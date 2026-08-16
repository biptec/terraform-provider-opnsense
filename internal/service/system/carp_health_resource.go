package system

import (
	"context"
	"fmt"

	"github.com/biptec/opnsense-go/pkg/api"
	apiextensions "github.com/biptec/opnsense-go/pkg/api_extensions"
	"github.com/biptec/opnsense-go/pkg/opnsense"
	"github.com/biptec/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &carpHealthResource{}
var _ resource.ResourceWithConfigure = &carpHealthResource{}
var _ resource.ResourceWithImportState = &carpHealthResource{}

type carpHealthResource struct{ client opnsense.Client }

func newCarpHealthResource() resource.Resource { return &carpHealthResource{} }
func (r *carpHealthResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_carp_health"
}
func (r *carpHealthResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = carpHealthResourceSchema()
}
func (r *carpHealthResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureClient(req, resp)
}

func (r *carpHealthResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data carpHealthResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.apply(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Unable to Configure CARP Health", err.Error())
		return
	}
	data.ID = types.StringValue("carp_health")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
func (r *carpHealthResource) Read(ctx context.Context, _ resource.ReadRequest, resp *resource.ReadResponse) {
	remote, err := r.client.ApiExtensions().CarpHealthGet(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read CARP Health", err.Error())
		return
	}
	state := carpHealthFromAPI(&remote.CarpHealth)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
func (r *carpHealthResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data carpHealthResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.apply(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Unable to Configure CARP Health", err.Error())
		return
	}
	data.ID = types.StringValue("carp_health")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
func (r *carpHealthResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	current, err := r.client.ApiExtensions().CarpHealthGet(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Disable CARP Health", err.Error())
		return
	}
	current.CarpHealth.Enabled = api.BoolString("0")
	result, err := r.client.ApiExtensions().CarpHealthSet(ctx, &current.CarpHealth)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Disable CARP Health", err.Error())
		return
	}
	if err := validateCarpHealthSet(result); err != nil {
		resp.Diagnostics.AddError("Unable to Disable CARP Health", err.Error())
		return
	}
	reconfigured, err := r.client.ApiExtensions().CarpHealthReconfigure(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Reconfigure CARP Health", err.Error())
		return
	}
	if err := validateCarpHealthReconfigure(reconfigured); err != nil {
		resp.Diagnostics.AddError("Unable to Reconfigure CARP Health", err.Error())
	}
}
func (r *carpHealthResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != "carp_health" {
		resp.Diagnostics.AddError("Invalid Import ID", "The CARP health singleton must be imported with ID `carp_health`.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
func (r *carpHealthResource) apply(ctx context.Context, data *carpHealthResourceModel) error {
	desired := carpHealthToAPI(data)
	current, err := r.client.ApiExtensions().CarpHealthGet(ctx)
	if err != nil {
		return err
	}
	if carpHealthEqual(&current.CarpHealth, desired) {
		return nil
	}
	result, err := r.client.ApiExtensions().CarpHealthSet(ctx, desired)
	if err != nil {
		return err
	}
	if err := validateCarpHealthSet(result); err != nil {
		return err
	}
	reconfigured, err := r.client.ApiExtensions().CarpHealthReconfigure(ctx)
	if err != nil {
		return err
	}
	return validateCarpHealthReconfigure(reconfigured)
}
func carpHealthToAPI(data *carpHealthResourceModel) *apiextensions.CarpHealthSettings {
	return &apiextensions.CarpHealthSettings{Enabled: api.BoolString(tools.BoolToString(data.Enabled.ValueBool())), Interval: tools.Int64ToString(data.Interval.ValueInt64()), FailureThreshold: tools.Int64ToString(data.FailureThreshold.ValueInt64()), RecoveryThreshold: tools.Int64ToString(data.RecoveryThreshold.ValueInt64())}
}
func carpHealthFromAPI(data *apiextensions.CarpHealthSettings) *carpHealthResourceModel {
	return &carpHealthResourceModel{Enabled: types.BoolValue(data.Enabled.Bool()), Interval: types.Int64Value(tools.StringToInt64(data.Interval)), FailureThreshold: types.Int64Value(tools.StringToInt64(data.FailureThreshold)), RecoveryThreshold: types.Int64Value(tools.StringToInt64(data.RecoveryThreshold)), ID: types.StringValue("carp_health")}
}
func carpHealthEqual(a, b *apiextensions.CarpHealthSettings) bool {
	return a.Enabled == b.Enabled && a.Interval == b.Interval && a.FailureThreshold == b.FailureThreshold && a.RecoveryThreshold == b.RecoveryThreshold
}
func validateCarpHealthSet(result *apiextensions.CarpHealthActionResult) error {
	if result == nil {
		return fmt.Errorf("CARP health settings API returned an empty response")
	}
	if result.Result != "saved" {
		return fmt.Errorf("CARP health settings API returned result %q instead of %q", result.Result, "saved")
	}
	return nil
}
func validateCarpHealthReconfigure(result *apiextensions.CarpHealthActionResult) error {
	if result == nil {
		return fmt.Errorf("CARP health reconfigure API returned an empty response")
	}
	if result.Status != "ok" && result.Result != "ok" {
		return fmt.Errorf("CARP health reconfigure API returned status %q result %q", result.Status, result.Result)
	}
	return nil
}
