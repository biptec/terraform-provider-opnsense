package caddy

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &settingsResource{}
var _ resource.ResourceWithConfigure = &settingsResource{}
var _ resource.ResourceWithImportState = &settingsResource{}
var _ resource.ResourceWithConfigValidators = &settingsResource{}

type settingsResource struct{ resourceClient }

func newSettingsResource() resource.Resource { return &settingsResource{} }
func (r *settingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_caddy_settings"
}
func (r *settingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = settingsResourceSchema()
}
func (r *settingsResource) ConfigValidators(context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{settingsDNSConfigValidator{}}
}
func (r *settingsResource) Create(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.AddError("Cannot Create Singleton Resource", "Caddy settings already exist in OPNsense. Import them first with: terraform import opnsense_caddy_settings.<name> caddy_settings")
}
func (r *settingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var old settingsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &old)...)
	if resp.Diagnostics.HasError() {
		return
	}
	remote, err := r.client.Caddy().SettingsGet(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Caddy Settings", err.Error())
		return
	}
	state, err := settingsStructToSchema(remote)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Decode Caddy Settings", err.Error())
		return
	}
	state.DNSCredentialsVersion = old.DNSCredentialsVersion
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
func (r *settingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan settingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	var dnsAPIKey, dnsRFC2136Key types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("dns_api_key"), &dnsAPIKey)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("dns_rfc2136_key"), &dnsRFC2136Key)...)
	if resp.Diagnostics.HasError() {
		return
	}
	remote, err := r.client.Caddy().SettingsGet(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Caddy Settings", err.Error())
		return
	}
	applySettingsModel(&remote.Caddy.General, &plan, dnsAPIKey, dnsRFC2136Key)
	if _, err = r.client.Caddy().SettingsSet(ctx, &remote.Caddy); err != nil {
		resp.Diagnostics.AddError("Unable to Update Caddy Settings", err.Error())
		return
	}
	if _, err = r.client.Caddy().ServiceReconfigure(ctx); err != nil {
		resp.Diagnostics.AddError("Unable to Reconfigure Caddy", err.Error())
		return
	}
	updated, err := r.client.Caddy().SettingsGet(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Caddy Settings Updated but Read Failed", err.Error())
		return
	}
	state, err := settingsStructToSchema(updated)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Decode Caddy Settings", err.Error())
		return
	}
	state.DNSCredentialsVersion = plan.DNSCredentialsVersion
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
func (r *settingsResource) Delete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning("Singleton Resource Removed From State Only", "Caddy settings remain unchanged in OPNsense. Re-import with ID `caddy_settings` to manage them again.")
}
func (r *settingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != "caddy_settings" {
		resp.Diagnostics.AddError("Invalid Import ID", fmt.Sprintf("Caddy settings must be imported with ID caddy_settings, got %q.", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	tflog.Info(ctx, "imported Caddy settings")
}
