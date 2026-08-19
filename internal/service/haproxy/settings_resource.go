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
func (r *settingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan settingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	state, err := r.applySettings(ctx, &plan)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Adopt HAProxy Settings", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	tflog.Info(ctx, "adopted existing HAProxy settings", map[string]any{"id": settingsID})
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
	state, err := r.applySettings(ctx, &plan)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Update HAProxy Settings", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *settingsResource) applySettings(ctx context.Context, plan *settingsModel) (*settingsModel, error) {
	remote, err := r.client.Haproxy().SettingsGet(ctx)
	if err != nil {
		return nil, fmt.Errorf("read existing HAProxy settings: %w", err)
	}
	current := settingsAPIToModel(remote)
	completeSettingsModel(plan, current)
	original := remote.HAProxy.General
	applySettingsModel(&remote.HAProxy.General, plan)
	if remote.HAProxy.General == original {
		return settingsAPIToModel(remote), nil
	}
	result, err := r.client.Haproxy().SettingsSet(ctx, &remote.HAProxy)
	if err != nil {
		return nil, fmt.Errorf("save HAProxy settings: %w", err)
	}
	if result == nil || result.Result != "saved" {
		return nil, fmt.Errorf("unexpected API result: %#v", result)
	}
	checked, err := r.client.Haproxy().ServiceConfigtest(ctx)
	if err != nil {
		return nil, fmt.Errorf("configuration test: %w", err)
	}
	if checked == nil || strings.Contains(strings.ToLower(checked.Result), "error") || strings.Contains(strings.ToLower(checked.Result), "failed") {
		return nil, fmt.Errorf("configuration test result: %#v", checked)
	}
	if _, err = r.client.Haproxy().ServiceReconfigure(ctx); err != nil {
		return nil, fmt.Errorf("reconfigure HAProxy: %w", err)
	}
	updated, err := r.client.Haproxy().SettingsGet(ctx)
	if err != nil {
		return nil, fmt.Errorf("read updated HAProxy settings: %w", err)
	}
	return settingsAPIToModel(updated), nil
}

func completeSettingsModel(plan, current *settingsModel) {
	if plan.Enabled.IsNull() || plan.Enabled.IsUnknown() {
		plan.Enabled = current.Enabled
	}
	if plan.GracefulStop.IsNull() || plan.GracefulStop.IsUnknown() {
		plan.GracefulStop = current.GracefulStop
	}
	if plan.HardStopAfter.IsNull() || plan.HardStopAfter.IsUnknown() {
		plan.HardStopAfter = current.HardStopAfter
	}
	if plan.CloseSpreadTime.IsNull() || plan.CloseSpreadTime.IsUnknown() {
		plan.CloseSpreadTime = current.CloseSpreadTime
	}
	if plan.SeamlessReload.IsNull() || plan.SeamlessReload.IsUnknown() {
		plan.SeamlessReload = current.SeamlessReload
	}
}
func (r *settingsResource) Delete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning("Singleton Resource Removed From State Only", "HAProxy settings remain unchanged in OPNsense. Re-add the resource to adopt the built-in singleton again, or import it explicitly with ID `haproxy_settings`.")
}
func (r *settingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != settingsID {
		resp.Diagnostics.AddError("Invalid Import ID", fmt.Sprintf("HAProxy settings must be imported with ID %s, got %q.", settingsID, req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	tflog.Info(ctx, "imported HAProxy settings")
}
