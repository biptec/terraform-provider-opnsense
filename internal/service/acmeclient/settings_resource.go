package acmeclient

import (
	"context"
	"fmt"
	"strings"

	apiacme "github.com/biptec/opnsense-go/pkg/acmeclient"
	"github.com/biptec/opnsense-go/pkg/api"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &settingsResource{}
var _ resource.ResourceWithConfigure = &settingsResource{}
var _ resource.ResourceWithImportState = &settingsResource{}

type settingsResource struct{ resourceClient }

func newSettingsResource() resource.Resource { return &settingsResource{} }
func (r *settingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_acme_settings"
}
func (r *settingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = settingsResourceSchema()
}
func (r *settingsResource) Create(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.AddError("Cannot Create Singleton Resource", "Native ACME settings already exist. Import with ID acmeclient_settings before managing them.")
}

func settingsState(remote *apiacme.SettingsResponse) *settingsResourceModel {
	s := remote.AcmeClient.Settings
	env := s.Environment.String()
	if env == "" {
		env = "prod"
	}
	logLevel := s.LogLevel.String()
	if logLevel == "" {
		logLevel = "normal"
	}
	return &settingsResourceModel{ID: types.StringValue("acmeclient_settings"), Enabled: types.BoolValue(stringBool(s.Enabled)), AutoRenewal: types.BoolValue(stringBool(s.AutoRenewal)), Environment: types.StringValue(env), LogLevel: types.StringValue(logLevel), ShowIntro: types.BoolValue(stringBool(s.ShowIntro))}
}
func (r *settingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var old settingsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &old)...)
	if resp.Diagnostics.HasError() {
		return
	}
	remote, err := r.client.Acmeclient().SettingsGet(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read ACME Settings", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, settingsState(remote))...)
}
func (r *settingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan settingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	remote, err := r.client.Acmeclient().SettingsGet(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read ACME Settings", err.Error())
		return
	}
	s := &remote.AcmeClient.Settings
	s.Enabled = boolString(plan.Enabled.ValueBool())
	s.AutoRenewal = boolString(plan.AutoRenewal.ValueBool())
	s.Environment = api.SelectedMap(plan.Environment.ValueString())
	s.LogLevel = api.SelectedMap(plan.LogLevel.ValueString())
	s.ShowIntro = boolString(plan.ShowIntro.ValueBool())
	result, err := r.client.Acmeclient().SettingsSet(ctx, &remote.AcmeClient)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Update ACME Settings", err.Error())
		return
	}
	if result.Result != "saved" {
		resp.Diagnostics.AddError("Unable to Update ACME Settings", fmt.Sprintf("settings result=%q validations=%v", result.Result, result.Validations))
		return
	}
	check, err := r.client.Acmeclient().ServiceConfigtest(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Validate ACME Configuration", err.Error())
		return
	}
	if strings.Contains(strings.ToUpper(check.Result), "ALERT") {
		resp.Diagnostics.AddError("Invalid ACME Configuration", check.Result)
		return
	}
	reconfigured, err := r.client.Acmeclient().ServiceReconfigure(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Reconfigure ACME Client", err.Error())
		return
	}
	if reconfigured.Status != "" && !strings.EqualFold(reconfigured.Status, "ok") {
		resp.Diagnostics.AddError("Unable to Reconfigure ACME Client", fmt.Sprintf("status=%q result=%q", reconfigured.Status, reconfigured.Result))
		return
	}
	if _, err = r.client.Acmeclient().SettingsFetchCronIntegration(ctx); err != nil {
		resp.Diagnostics.AddError("Unable to Reconcile ACME Renewal Cron", err.Error())
		return
	}
	updated, err := r.client.Acmeclient().SettingsGet(ctx)
	if err != nil {
		resp.Diagnostics.AddError("ACME Settings Updated but Read Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, settingsState(updated))...)
}
func (r *settingsResource) Delete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning("Singleton Resource Removed From State Only", "Native ACME settings remain unchanged in OPNsense.")
}
func (r *settingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != "acmeclient_settings" {
		resp.Diagnostics.AddError("Invalid Import ID", fmt.Sprintf("ACME settings must be imported with ID acmeclient_settings, got %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
