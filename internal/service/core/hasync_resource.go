package core

import (
	"context"
	"fmt"

	"github.com/biptec/opnsense-go/pkg/api"
	"github.com/biptec/opnsense-go/pkg/opnsense"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
func (r *hasyncResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var old hasyncModel
	resp.Diagnostics.Append(req.State.Get(ctx, &old)...)
	if resp.Diagnostics.HasError() {
		return
	}
	remote, err := r.client.Core().HasyncGet(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read OPNsense HA Settings", err.Error())
		return
	}
	state := hasyncAPIToModel(&remote.Hasync)
	state.PasswordVersion = old.PasswordVersion
	if state.PasswordVersion.IsNull() || state.PasswordVersion.IsUnknown() {
		state.PasswordVersion = types.Int64Value(0)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
func (r *hasyncResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, old hasyncModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &old)...)
	var password types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("password"), &password)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !plan.PasswordVersion.IsUnknown() && !old.PasswordVersion.IsUnknown() && !plan.PasswordVersion.IsNull() && !old.PasswordVersion.IsNull() && plan.PasswordVersion.ValueInt64() != old.PasswordVersion.ValueInt64() {
		if password.IsNull() || password.IsUnknown() || password.ValueString() == "" {
			resp.Diagnostics.AddError("Missing Rotated XMLRPC Password", "password_version changed but no write-only password was supplied. Provide the new password together with the incremented password_version.")
			return
		}
	}
	remote, err := r.client.Core().HasyncGet(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read OPNsense HA Settings", err.Error())
		return
	}
	applyHasyncModel(&remote.Hasync, &plan, password)
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
	state := hasyncAPIToModel(&updated.Hasync)
	state.PasswordVersion = plan.PasswordVersion
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
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
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("password_version"), int64(0))...)
	tflog.Info(ctx, "imported OPNsense HA settings")
}
