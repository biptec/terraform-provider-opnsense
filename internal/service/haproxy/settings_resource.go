package haproxy

import (
	"context"
	"fmt"
	"strings"

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
	resp.TypeName = req.ProviderTypeName + "_haproxy_settings"
}
func (r *settingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = settingsResourceSchema()
}
func (r *settingsResource) Create(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.AddError("Cannot Create Singleton Resource", "HAProxy settings already exist in OPNsense. Import them first with: terraform import opnsense_haproxy_settings.<name> haproxy_settings")
}
func (r *settingsResource) Read(ctx context.Context, _ resource.ReadRequest, resp *resource.ReadResponse) {
	remote, err := r.client.Haproxy().SettingsGet(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read HAProxy Settings", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, settingsAPIToModel(remote))...)
}
func (r *settingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan settingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	remote, err := r.client.Haproxy().SettingsGet(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read HAProxy Settings", err.Error())
		return
	}
	applySettingsModel(&remote.HAProxy.General, &plan)
	result, err := r.client.Haproxy().SettingsSet(ctx, &remote.HAProxy)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Update HAProxy Settings", err.Error())
		return
	}
	if result == nil || result.Result != "saved" {
		resp.Diagnostics.AddError("Unable to Update HAProxy Settings", fmt.Sprintf("unexpected API result: %#v", result))
		return
	}
	checked, err := r.client.Haproxy().ServiceConfigtest(ctx)
	if err != nil {
		resp.Diagnostics.AddError("HAProxy Configuration Test Failed", err.Error())
		return
	}
	if checked == nil || strings.Contains(strings.ToLower(checked.Result), "error") || strings.Contains(strings.ToLower(checked.Result), "failed") {
		resp.Diagnostics.AddError("HAProxy Configuration Test Failed", fmt.Sprintf("configtest result: %#v", checked))
		return
	}
	if _, err = r.client.Haproxy().ServiceReconfigure(ctx); err != nil {
		resp.Diagnostics.AddError("Unable to Reconfigure HAProxy", err.Error())
		return
	}
	updated, err := r.client.Haproxy().SettingsGet(ctx)
	if err != nil {
		resp.Diagnostics.AddError("HAProxy Settings Updated but Read Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, settingsAPIToModel(updated))...)
}
func (r *settingsResource) Delete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning("Singleton Resource Removed From State Only", "HAProxy settings remain unchanged in OPNsense. Re-import with ID `haproxy_settings` to manage them again.")
}
func (r *settingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != settingsID {
		resp.Diagnostics.AddError("Invalid Import ID", fmt.Sprintf("HAProxy settings must be imported with ID %s, got %q.", settingsID, req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	tflog.Info(ctx, "imported HAProxy settings")
}
