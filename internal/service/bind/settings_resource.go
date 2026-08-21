package bind

import (
	"context"
	"fmt"
	"reflect"

	"github.com/biptec/opnsense-go/pkg/api"
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
func (r *settingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan settingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	state, err := r.applySettings(ctx, &plan)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Adopt BIND Settings", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	tflog.Info(ctx, "adopted existing BIND settings", map[string]any{"id": "bind_settings"})
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
	state, err := r.applySettings(ctx, &plan)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Update BIND Settings", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *settingsResource) applySettings(ctx context.Context, plan *settingsResourceModel) (*settingsResourceModel, error) {
	remote, err := r.client.Bind().SettingsGet(ctx)
	if err != nil {
		return nil, fmt.Errorf("read existing BIND settings: %w", err)
	}
	original := remote.General
	applySettingsModel(&remote.General, plan)
	if reflect.DeepEqual(remote.General, original) {
		return settingsAPIToModel(remote)
	}
	result, err := r.client.Bind().SettingsSet(ctx, &remote.General)
	if err != nil {
		return nil, fmt.Errorf("save BIND settings: %w", err)
	}
	if err := validateSettingsSetResult(result); err != nil {
		return nil, err
	}
	if _, err = r.client.Bind().ServiceReconfigure(ctx); err != nil {
		return nil, fmt.Errorf("reconfigure BIND: %w", err)
	}
	updated, err := r.client.Bind().SettingsGet(ctx)
	if err != nil {
		return nil, fmt.Errorf("read updated BIND settings: %w", err)
	}
	return settingsAPIToModel(updated)
}

func validateSettingsSetResult(result *api.ActionResult) error {
	if result == nil {
		return fmt.Errorf("BIND settings API returned an empty response")
	}
	if result.Result != "saved" {
		return fmt.Errorf("BIND settings API returned result %q instead of %q", result.Result, "saved")
	}
	return nil
}

func (r *settingsResource) Delete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning("Singleton Resource Removed From State Only", "BIND settings remain unchanged in OPNsense. Re-add the resource to adopt the built-in singleton again, or import it explicitly with ID `bind_settings`.")
}
func (r *settingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != "bind_settings" {
		resp.Diagnostics.AddError("Invalid Import ID", fmt.Sprintf("BIND settings must be imported with ID bind_settings, got %q.", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	tflog.Info(ctx, "imported BIND settings")
}
