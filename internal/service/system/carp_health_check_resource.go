package system

import (
	"context"
	"errors"
	"fmt"
	"net/netip"

	"github.com/biptec/opnsense-go/pkg/api"
	apiextensions "github.com/biptec/opnsense-go/pkg/api_extensions"
	"github.com/biptec/opnsense-go/pkg/errs"
	"github.com/biptec/opnsense-go/pkg/opnsense"
	"github.com/biptec/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &carpHealthCheckResource{}
var _ resource.ResourceWithConfigure = &carpHealthCheckResource{}
var _ resource.ResourceWithImportState = &carpHealthCheckResource{}
var _ resource.ResourceWithConfigValidators = &carpHealthCheckResource{}

type carpHealthCheckResource struct{ client opnsense.Client }

type carpHealthCheckConfigValidator struct{}

func (carpHealthCheckConfigValidator) Description(_ context.Context) string {
	return "CARP health scope, explicit VHID targets, and fallback route pairs must be internally consistent"
}
func (v carpHealthCheckConfigValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}
func (carpHealthCheckConfigValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data carpHealthCheckResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := validateCarpHealthCheckConfiguration(&data); err != nil {
		resp.Diagnostics.AddError("Invalid CARP Health Check", err.Error())
	}
}

func newCarpHealthCheckResource() resource.Resource { return &carpHealthCheckResource{} }
func (r *carpHealthCheckResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_carp_health_check"
}
func (r *carpHealthCheckResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = carpHealthCheckResourceSchema()
}
func (r *carpHealthCheckResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{carpHealthCheckConfigValidator{}}
}
func (r *carpHealthCheckResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureClient(req, resp)
}
func (r *carpHealthCheckResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data carpHealthCheckResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	desired, err := carpHealthCheckToAPI(&data)
	if err != nil {
		resp.Diagnostics.AddError("Invalid CARP Health Check", err.Error())
		return
	}
	id, err := r.client.ApiExtensions().AddCarpHealthCheck(ctx, desired)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create CARP Health Check", err.Error())
		return
	}
	remote, err := r.client.ApiExtensions().GetCarpHealthCheck(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("CARP Health Check Created but Read Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, carpHealthCheckFromAPI(remote, id))...)
}
func (r *carpHealthCheckResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data carpHealthCheckResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	remote, err := r.client.ApiExtensions().GetCarpHealthCheck(ctx, data.ID.ValueString())
	if err != nil {
		var notFound *errs.NotFoundError
		if errors.As(err, &notFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to Read CARP Health Check", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, carpHealthCheckFromAPI(remote, data.ID.ValueString()))...)
}
func (r *carpHealthCheckResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data carpHealthCheckResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	desired, err := carpHealthCheckToAPI(&data)
	if err != nil {
		resp.Diagnostics.AddError("Invalid CARP Health Check", err.Error())
		return
	}
	if err := r.client.ApiExtensions().UpdateCarpHealthCheck(ctx, data.ID.ValueString(), desired); err != nil {
		resp.Diagnostics.AddError("Unable to Update CARP Health Check", err.Error())
		return
	}
	remote, err := r.client.ApiExtensions().GetCarpHealthCheck(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("CARP Health Check Updated but Read Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, carpHealthCheckFromAPI(remote, data.ID.ValueString()))...)
}
func (r *carpHealthCheckResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data carpHealthCheckResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.ApiExtensions().DeleteCarpHealthCheck(ctx, data.ID.ValueString()); err != nil {
		var notFound *errs.NotFoundError
		if !errors.As(err, &notFound) {
			resp.Diagnostics.AddError("Unable to Delete CARP Health Check", err.Error())
		}
	}
}
func (r *carpHealthCheckResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func validateCarpHealthCheckConfiguration(data *carpHealthCheckResourceModel) error {
	if data.Scope.IsUnknown() {
		return nil
	}
	scope := data.Scope.ValueString()
	if scope == "" {
		scope = "interface"
	}
	if !data.VHID.IsUnknown() {
		vhid := data.VHID.ValueInt64()
		if scope == "vhid" && (vhid < 1 || vhid > 255) {
			return fmt.Errorf("vhid must be between 1 and 255 when scope is vhid")
		}
		if scope != "vhid" && vhid != 0 {
			return fmt.Errorf("vhid must be zero unless scope is vhid")
		}
	}
	if !data.VHIDTargets.IsUnknown() {
		targets := carpHealthVHIDTargets(data.VHIDTargets)
		if scope == "vhid_group" && len(targets) == 0 {
			return fmt.Errorf("vhid_targets must contain at least one interface:VHID entry when scope is vhid_group")
		}
		if scope != "vhid_group" && len(targets) != 0 {
			return fmt.Errorf("vhid_targets must be empty unless scope is vhid_group")
		}
	}
	if err := validateFallbackPair(data.FallbackIPv4Target, data.FallbackIPv4Gateway, "IPv4"); err != nil {
		return err
	}
	if err := validateFallbackPair(data.FallbackIPv6Target, data.FallbackIPv6Gateway, "IPv6"); err != nil {
		return err
	}
	return nil
}

func carpHealthVHIDTargets(value types.Set) []string {
	if value.IsNull() || value.IsUnknown() || value.ElementType(context.Background()) == nil {
		return nil
	}
	return tools.SetToStringSlice(value)
}

func validateFallbackPair(target, gateway types.String, family string) error {
	if target.IsUnknown() || gateway.IsUnknown() {
		return nil
	}
	hasTarget := target.ValueString() != ""
	hasGateway := gateway.ValueString() != ""
	if hasTarget != hasGateway {
		return fmt.Errorf("%s fallback target and gateway must be configured together", family)
	}
	return nil
}

func carpHealthCheckToAPI(data *carpHealthCheckResourceModel) (*apiextensions.CarpHealthCheck, error) {
	if err := validateCarpHealthCheckConfiguration(data); err != nil {
		return nil, err
	}
	target, err := netip.ParseAddr(data.Target.ValueString())
	if err != nil || !target.Is4() {
		return nil, fmt.Errorf("target must be an IPv4 address")
	}
	scope := data.Scope.ValueString()
	if scope == "" {
		scope = "interface"
	}
	vhid := int64(0)
	if scope == "vhid" {
		vhid = data.VHID.ValueInt64()
	}
	failureAdvSkew := data.FailureAdvSkew.ValueInt64()
	if data.FailureAdvSkew.IsNull() || failureAdvSkew == 0 {
		failureAdvSkew = 254
	}
	vhidTargets := []string{}
	if scope == "vhid_group" {
		vhidTargets = carpHealthVHIDTargets(data.VHIDTargets)
	}
	fallbackIPv4Target, err := normalizeOptionalIP(data.FallbackIPv4Target.ValueString(), 4)
	if err != nil {
		return nil, fmt.Errorf("fallback_ipv4_target: %w", err)
	}
	fallbackIPv4Gateway, err := normalizeOptionalIP(data.FallbackIPv4Gateway.ValueString(), 4)
	if err != nil {
		return nil, fmt.Errorf("fallback_ipv4_gateway: %w", err)
	}
	fallbackIPv6Target, err := normalizeOptionalIP(data.FallbackIPv6Target.ValueString(), 6)
	if err != nil {
		return nil, fmt.Errorf("fallback_ipv6_target: %w", err)
	}
	fallbackIPv6Gateway, err := normalizeOptionalIP(data.FallbackIPv6Gateway.ValueString(), 6)
	if err != nil {
		return nil, fmt.Errorf("fallback_ipv6_gateway: %w", err)
	}
	return &apiextensions.CarpHealthCheck{
		Enabled: api.BoolString(tools.BoolToString(data.Enabled.ValueBool())), Name: data.Name.ValueString(),
		Interface: api.SelectedMap(data.Interface.ValueString()), Target: target.String(),
		Scope: api.SelectedMap(scope), VHID: tools.Int64ToString(vhid), FailureAdvSkew: tools.Int64ToString(failureAdvSkew),
		VHIDTargets:        api.SelectedMapList(vhidTargets),
		FallbackIPv4Target: fallbackIPv4Target, FallbackIPv4Gateway: fallbackIPv4Gateway,
		FallbackIPv6Target: fallbackIPv6Target, FallbackIPv6Gateway: fallbackIPv6Gateway,
	}, nil
}

func normalizeOptionalIP(value string, version int) (string, error) {
	if value == "" {
		return "", nil
	}
	addr, err := netip.ParseAddr(value)
	if err != nil || (version == 4 && !addr.Is4()) || (version == 6 && !addr.Is6()) {
		return "", fmt.Errorf("must be an IPv%d address", version)
	}
	return addr.String(), nil
}

func carpHealthCheckFromAPI(data *apiextensions.CarpHealthCheck, id string) *carpHealthCheckResourceModel {
	scope := data.Scope.String()
	if scope == "" {
		// Old checks without an explicit scope are legacy global checks. Do not reinterpret them as the new automatic default.
		scope = "global"
	}
	vhid := int64(0)
	if scope == "vhid" && data.VHID != "" {
		vhid = tools.StringToInt64(data.VHID)
	}
	failureAdvSkew := tools.StringToInt64(data.FailureAdvSkew)
	if failureAdvSkew < 1 || failureAdvSkew > 254 {
		failureAdvSkew = 254
	}
	return &carpHealthCheckResourceModel{
		Enabled: types.BoolValue(data.Enabled.Bool()), Name: types.StringValue(data.Name),
		Interface: types.StringValue(data.Interface.String()), Target: types.StringValue(data.Target),
		Scope: types.StringValue(scope), VHID: types.Int64Value(vhid), FailureAdvSkew: types.Int64Value(failureAdvSkew),
		VHIDTargets:        tools.StringSliceToSet([]string(data.VHIDTargets)),
		FallbackIPv4Target: types.StringValue(data.FallbackIPv4Target), FallbackIPv4Gateway: types.StringValue(data.FallbackIPv4Gateway),
		FallbackIPv6Target: types.StringValue(data.FallbackIPv6Target), FallbackIPv6Gateway: types.StringValue(data.FallbackIPv6Gateway),
		ID: types.StringValue(id),
	}
}
