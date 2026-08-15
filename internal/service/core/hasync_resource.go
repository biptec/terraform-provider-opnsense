package core

import (
	"context"
	"fmt"

	"github.com/biptec/opnsense-go/pkg/api"
	"github.com/biptec/opnsense-go/pkg/opnsense"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &hasyncResource{}
var _ resource.ResourceWithConfigure = &hasyncResource{}
var _ resource.ResourceWithImportState = &hasyncResource{}

type hasyncResource struct{ client opnsense.Client }

func newHasyncResource() resource.Resource { return &hasyncResource{} }
func (r *hasyncResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_core_hasync"
}
func (r *hasyncResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = hasyncResourceSchema()
}
func (r *hasyncResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *api.Client, got %T.", req.ProviderData))
		return
	}
	r.client = opnsense.NewClient(c)
}
func (r *hasyncResource) Create(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.AddError("Cannot Create Singleton Resource", "OPNsense HA settings already exist. Import them first with: terraform import opnsense_core_hasync.<name> core_hasync")
}
func (r *hasyncResource) Read(ctx context.Context, _ resource.ReadRequest, resp *resource.ReadResponse) {
	remote, err := r.client.Core().HasyncGet(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read OPNsense HA Settings", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, hasyncAPIToModel(&remote.Hasync))...)
}
func (r *hasyncResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan hasyncModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	remote, err := r.client.Core().HasyncGet(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read OPNsense HA Settings", err.Error())
		return
	}
	applyHasyncModel(&remote.Hasync, &plan)
	result, err := r.client.Core().HasyncSet(ctx, &remote.Hasync)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Update OPNsense HA Settings", err.Error())
		return
	}
	if result == nil || result.Result != "saved" {
		resp.Diagnostics.AddError("Unable to Update OPNsense HA Settings", fmt.Sprintf("unexpected API result: %#v", result))
		return
	}
	if _, err := r.client.Core().HasyncReconfigure(ctx); err != nil {
		resp.Diagnostics.AddError("Unable to Reconfigure pfsync", err.Error())
		return
	}
	updated, err := r.client.Core().HasyncGet(ctx)
	if err != nil {
		resp.Diagnostics.AddError("OPNsense HA Settings Updated but Read Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, hasyncAPIToModel(&updated.Hasync))...)
}
func (r *hasyncResource) Delete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning("Singleton Resource Removed From State Only", "OPNsense HA settings remain unchanged. Re-import with ID `core_hasync` to manage them again.")
}
func (r *hasyncResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != hasyncID {
		resp.Diagnostics.AddError("Invalid Import ID", fmt.Sprintf("OPNsense HA settings must be imported with ID %s, got %q.", hasyncID, req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	tflog.Info(ctx, "imported OPNsense HA settings")
}
