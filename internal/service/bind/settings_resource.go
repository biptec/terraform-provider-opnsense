package bind

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &settingsResource{}
var _ resource.ResourceWithConfigure = &settingsResource{}
var _ resource.ResourceWithImportState = &settingsResource{}

type settingsResource struct{ resourceClient }

func newSettingsResource() resource.Resource { return &settingsResource{} }
func (r *settingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bind_settings"
}
func (r *settingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = settingsResourceSchema()
}
func (r *settingsResource) Create(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.AddError("Cannot Create Singleton Resource", "BIND settings already exist in OPNsense. Import them first with: terraform import opnsense_bind_settings.<name> bind_settings")
}
func (r *settingsResource) Read(ctx context.Context, _ resource.ReadRequest, resp *resource.ReadResponse) {
	remote, err := r.client.Bind().SettingsGet(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read BIND Settings", err.Error())
		return
	}
	state, err := settingsAPIToModel(remote)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Decode BIND Settings", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
func (r *settingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan settingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	remote, err := r.client.Bind().SettingsGet(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read BIND Settings", err.Error())
		return
	}
	applySettingsModel(&remote.General, &plan)
	if _, err = r.client.Bind().SettingsSet(ctx, &remote.General); err != nil {
		resp.Diagnostics.AddError("Unable to Update BIND Settings", err.Error())
		return
	}
	if _, err = r.client.Bind().ServiceReconfigure(ctx); err != nil {
		resp.Diagnostics.AddError("Unable to Reconfigure BIND", err.Error())
		return
	}
	updated, err := r.client.Bind().SettingsGet(ctx)
	if err != nil {
		resp.Diagnostics.AddError("BIND Settings Updated but Read Failed", err.Error())
		return
	}
	state, err := settingsAPIToModel(updated)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Decode BIND Settings", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
func (r *settingsResource) Delete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning("Singleton Resource Removed From State Only", "BIND settings remain unchanged in OPNsense. Re-import with ID `bind_settings` to manage them again.")
}
func (r *settingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != "bind_settings" {
		resp.Diagnostics.AddError("Invalid Import ID", fmt.Sprintf("BIND settings must be imported with ID bind_settings, got %q.", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	tflog.Info(ctx, "imported BIND settings")
}
