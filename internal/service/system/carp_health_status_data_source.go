package system

import (
	"context"
	"fmt"

	"github.com/biptec/opnsense-go/pkg/api"
	apiextensions "github.com/biptec/opnsense-go/pkg/api_extensions"
	"github.com/biptec/opnsense-go/pkg/opnsense"
	"github.com/biptec/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type carpHealthStatusCheckModel struct {
	UUID                  types.String `tfsdk:"uuid"`
	Name                  types.String `tfsdk:"name"`
	Interface             types.String `tfsdk:"interface"`
	Device                types.String `tfsdk:"device"`
	Target                types.String `tfsdk:"target"`
	Scope                 types.String `tfsdk:"scope"`
	VHID                  types.Int64  `tfsdk:"vhid"`
	VHIDTargets           types.Set    `tfsdk:"vhid_targets"`
	ConfiguredVHIDTargets types.Set    `tfsdk:"configured_vhid_targets"`
	FailureAdvSkew        types.Int64  `tfsdk:"failure_advskew"`
	VHIDStates            types.List   `tfsdk:"vhid_states"`
	FallbackRoutes        types.List   `tfsdk:"fallback_routes"`
	CarpState             types.String `tfsdk:"carp_state"`
	ConfiguredAdvSkew     types.Int64  `tfsdk:"configured_advskew"`
	CurrentAdvSkew        types.Int64  `tfsdk:"current_advskew"`
	ControlOK             types.Bool   `tfsdk:"control_ok"`
	Healthy               types.Bool   `tfsdk:"healthy"`
	Failures              types.Int64  `tfsdk:"failures"`
	Successes             types.Int64  `tfsdk:"successes"`
}

type carpHealthStatusGlobalModel struct {
	Active     types.Bool  `tfsdk:"active"`
	CheckCount types.Int64 `tfsdk:"check_count"`
	Ready      types.Bool  `tfsdk:"ready"`
	Healthy    types.Bool  `tfsdk:"healthy"`
}

type carpHealthStatusRouteModel struct {
	Key              types.String `tfsdk:"key"`
	CheckUUID        types.String `tfsdk:"check_uuid"`
	Check            types.String `tfsdk:"check"`
	Family           types.String `tfsdk:"family"`
	Destination      types.String `tfsdk:"destination"`
	Gateway          types.String `tfsdk:"gateway"`
	DesiredInstalled types.Bool   `tfsdk:"desired_installed"`
	Installed        types.Bool   `tfsdk:"installed"`
	Managed          types.Bool   `tfsdk:"managed"`
	ControlOK        types.Bool   `tfsdk:"control_ok"`
	Retired          types.Bool   `tfsdk:"retired"`
	Error            types.String `tfsdk:"error"`
}

type carpHealthStatusVHIDModel struct {
	Key               types.String `tfsdk:"key"`
	Interface         types.String `tfsdk:"interface"`
	Device            types.String `tfsdk:"device"`
	VHID              types.Int64  `tfsdk:"vhid"`
	Checks            types.List   `tfsdk:"checks"`
	Ready             types.Bool   `tfsdk:"ready"`
	Healthy           types.Bool   `tfsdk:"healthy"`
	DesiredDemoted    types.Bool   `tfsdk:"desired_demoted"`
	DesiredAdvSkew    types.Int64  `tfsdk:"desired_advskew"`
	ConfiguredAdvSkew types.Int64  `tfsdk:"configured_advskew"`
	CurrentAdvSkew    types.Int64  `tfsdk:"current_advskew"`
	CarpState         types.String `tfsdk:"carp_state"`
	ControlOK         types.Bool   `tfsdk:"control_ok"`
	Retired           types.Bool   `tfsdk:"retired"`
	Error             types.String `tfsdk:"error"`
}

type carpHealthStatusModel struct {
	ID              types.String  `tfsdk:"id"`
	Status          types.String  `tfsdk:"status"`
	Enabled         types.Bool    `tfsdk:"enabled"`
	Ready           types.Bool    `tfsdk:"ready"`
	Healthy         types.Bool    `tfsdk:"healthy"`
	ProbeHealthy    types.Bool    `tfsdk:"probe_healthy"`
	ControlOK       types.Bool    `tfsdk:"control_ok"`
	Running         types.Bool    `tfsdk:"running"`
	Timestamp       types.Float64 `tfsdk:"timestamp"`
	ConfigSignature types.String  `tfsdk:"config_signature"`
	Global          types.Object  `tfsdk:"global"`
	VHIDs           types.List    `tfsdk:"vhids"`
	Routes          types.List    `tfsdk:"routes"`
	Checks          types.List    `tfsdk:"checks"`
}

var carpHealthStatusGlobalTypes = map[string]attr.Type{
	"active": types.BoolType, "check_count": types.Int64Type,
	"ready": types.BoolType, "healthy": types.BoolType,
}
var carpHealthStatusRouteTypes = map[string]attr.Type{
	"key": types.StringType, "check_uuid": types.StringType, "check": types.StringType,
	"family": types.StringType, "destination": types.StringType, "gateway": types.StringType,
	"desired_installed": types.BoolType, "installed": types.BoolType, "managed": types.BoolType,
	"control_ok": types.BoolType, "retired": types.BoolType, "error": types.StringType,
}
var carpHealthStatusVHIDTypes = map[string]attr.Type{
	"key": types.StringType, "interface": types.StringType, "device": types.StringType,
	"vhid": types.Int64Type, "checks": types.ListType{ElemType: types.StringType},
	"ready": types.BoolType, "healthy": types.BoolType, "desired_demoted": types.BoolType,
	"desired_advskew": types.Int64Type, "configured_advskew": types.Int64Type, "current_advskew": types.Int64Type,
	"carp_state": types.StringType, "control_ok": types.BoolType,
	"retired": types.BoolType, "error": types.StringType,
}
var carpHealthStatusCheckTypes = map[string]attr.Type{
	"uuid": types.StringType, "name": types.StringType, "interface": types.StringType,
	"device": types.StringType, "target": types.StringType, "scope": types.StringType,
	"vhid": types.Int64Type, "vhid_targets": types.SetType{ElemType: types.StringType},
	"configured_vhid_targets": types.SetType{ElemType: types.StringType}, "failure_advskew": types.Int64Type,
	"vhid_states":     types.ListType{ElemType: types.ObjectType{AttrTypes: carpHealthStatusVHIDTypes}},
	"fallback_routes": types.ListType{ElemType: types.ObjectType{AttrTypes: carpHealthStatusRouteTypes}},
	"carp_state":      types.StringType, "configured_advskew": types.Int64Type, "current_advskew": types.Int64Type,
	"control_ok": types.BoolType, "healthy": types.BoolType,
	"failures": types.Int64Type, "successes": types.Int64Type,
}

var _ datasource.DataSource = &carpHealthStatusDataSource{}
var _ datasource.DataSourceWithConfigure = &carpHealthStatusDataSource{}

type carpHealthStatusDataSource struct{ client opnsense.Client }

func newCarpHealthStatusDataSource() datasource.DataSource { return &carpHealthStatusDataSource{} }
func (d *carpHealthStatusDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_carp_health_status"
}
func (d *carpHealthStatusDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{MarkdownDescription: "Reads runtime CARP health-monitor status from `os-api-extensions`, including resolved automatic targets, per-VHID control state, and managed fallback routes.", Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{Computed: true}, "status": schema.StringAttribute{Computed: true},
		"enabled": schema.BoolAttribute{Computed: true}, "ready": schema.BoolAttribute{Computed: true},
		"healthy": schema.BoolAttribute{Computed: true}, "probe_healthy": schema.BoolAttribute{Computed: true},
		"control_ok": schema.BoolAttribute{Computed: true}, "running": schema.BoolAttribute{Computed: true},
		"timestamp": schema.Float64Attribute{Computed: true}, "config_signature": schema.StringAttribute{Computed: true},
		"global": schema.SingleNestedAttribute{Computed: true, Attributes: map[string]schema.Attribute{
			"active": schema.BoolAttribute{Computed: true}, "check_count": schema.Int64Attribute{Computed: true},
			"ready": schema.BoolAttribute{Computed: true}, "healthy": schema.BoolAttribute{Computed: true},
		}},
		"vhids":  schema.ListNestedAttribute{Computed: true, NestedObject: schema.NestedAttributeObject{Attributes: carpHealthVHIDSchemaAttributes()}},
		"routes": schema.ListNestedAttribute{Computed: true, NestedObject: schema.NestedAttributeObject{Attributes: carpHealthRouteSchemaAttributes()}},
		"checks": schema.ListNestedAttribute{Computed: true, NestedObject: schema.NestedAttributeObject{Attributes: carpHealthCheckStatusSchemaAttributes()}},
	}}
}

func carpHealthVHIDSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"key": schema.StringAttribute{Computed: true}, "interface": schema.StringAttribute{Computed: true},
		"device": schema.StringAttribute{Computed: true}, "vhid": schema.Int64Attribute{Computed: true},
		"checks": schema.ListAttribute{Computed: true, ElementType: types.StringType},
		"ready":  schema.BoolAttribute{Computed: true}, "healthy": schema.BoolAttribute{Computed: true},
		"desired_demoted": schema.BoolAttribute{Computed: true}, "desired_advskew": schema.Int64Attribute{Computed: true},
		"configured_advskew": schema.Int64Attribute{Computed: true}, "current_advskew": schema.Int64Attribute{Computed: true},
		"carp_state": schema.StringAttribute{Computed: true}, "control_ok": schema.BoolAttribute{Computed: true},
		"retired": schema.BoolAttribute{Computed: true}, "error": schema.StringAttribute{Computed: true},
	}
}

func carpHealthRouteSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"key": schema.StringAttribute{Computed: true}, "check_uuid": schema.StringAttribute{Computed: true},
		"check": schema.StringAttribute{Computed: true}, "family": schema.StringAttribute{Computed: true},
		"destination": schema.StringAttribute{Computed: true}, "gateway": schema.StringAttribute{Computed: true},
		"desired_installed": schema.BoolAttribute{Computed: true}, "installed": schema.BoolAttribute{Computed: true},
		"managed": schema.BoolAttribute{Computed: true}, "control_ok": schema.BoolAttribute{Computed: true},
		"retired": schema.BoolAttribute{Computed: true}, "error": schema.StringAttribute{Computed: true},
	}
}

func carpHealthCheckStatusSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"uuid": schema.StringAttribute{Computed: true}, "name": schema.StringAttribute{Computed: true},
		"interface": schema.StringAttribute{Computed: true}, "device": schema.StringAttribute{Computed: true},
		"target": schema.StringAttribute{Computed: true}, "scope": schema.StringAttribute{Computed: true},
		"vhid": schema.Int64Attribute{Computed: true}, "vhid_targets": schema.SetAttribute{Computed: true, ElementType: types.StringType},
		"configured_vhid_targets": schema.SetAttribute{Computed: true, ElementType: types.StringType},
		"failure_advskew":         schema.Int64Attribute{Computed: true},
		"vhid_states":             schema.ListNestedAttribute{Computed: true, NestedObject: schema.NestedAttributeObject{Attributes: carpHealthVHIDSchemaAttributes()}},
		"fallback_routes":         schema.ListNestedAttribute{Computed: true, NestedObject: schema.NestedAttributeObject{Attributes: carpHealthRouteSchemaAttributes()}},
		"carp_state":              schema.StringAttribute{Computed: true},
		"configured_advskew":      schema.Int64Attribute{Computed: true}, "current_advskew": schema.Int64Attribute{Computed: true},
		"control_ok": schema.BoolAttribute{Computed: true}, "healthy": schema.BoolAttribute{Computed: true},
		"failures": schema.Int64Attribute{Computed: true}, "successes": schema.Int64Attribute{Computed: true},
	}
}

func (d *carpHealthStatusDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected *api.Client, got %T.", req.ProviderData))
		return
	}
	d.client = opnsense.NewClient(client)
}

func (d *carpHealthStatusDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	remote, err := d.client.ApiExtensions().CarpHealthStatus(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read CARP Health Status", err.Error())
		return
	}
	checks := make([]carpHealthStatusCheckModel, 0, len(remote.Checks))
	for _, item := range remote.Checks {
		vhidStates, diagnostics := carpHealthStatusVHIDList(ctx, item.VHIDStates)
		resp.Diagnostics.Append(diagnostics...)
		fallbackRoutes, diagnostics := carpHealthStatusRouteList(ctx, item.FallbackRoutes)
		resp.Diagnostics.Append(diagnostics...)
		checks = append(checks, carpHealthStatusCheckModel{
			UUID: types.StringValue(item.UUID), Name: types.StringValue(item.Name), Interface: types.StringValue(item.Interface),
			Device: types.StringValue(item.Device), Target: types.StringValue(item.Target), Scope: types.StringValue(item.Scope),
			VHID: types.Int64Value(int64(item.VHID)), VHIDTargets: tools.StringSliceToSet(item.VHIDTargets),
			ConfiguredVHIDTargets: tools.StringSliceToSet(item.ConfiguredVHIDTargets), FailureAdvSkew: types.Int64Value(int64(item.FailureAdvSkew)),
			VHIDStates: vhidStates, FallbackRoutes: fallbackRoutes, CarpState: types.StringValue(item.CarpState),
			ConfiguredAdvSkew: nullableInt(item.ConfiguredAdvSkew), CurrentAdvSkew: nullableInt(item.CurrentAdvSkew),
			ControlOK: types.BoolValue(item.ControlOK), Healthy: types.BoolValue(item.Healthy),
			Failures: types.Int64Value(int64(item.Failures)), Successes: types.Int64Value(int64(item.Successes)),
		})
	}
	checkList, diagnostics := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: carpHealthStatusCheckTypes}, checks)
	resp.Diagnostics.Append(diagnostics...)
	global, diagnostics := types.ObjectValueFrom(ctx, carpHealthStatusGlobalTypes, carpHealthStatusGlobalModel{
		Active: types.BoolValue(remote.Global.Active), CheckCount: types.Int64Value(int64(remote.Global.CheckCount)),
		Ready: types.BoolValue(remote.Global.Ready), Healthy: types.BoolValue(remote.Global.Healthy),
	})
	resp.Diagnostics.Append(diagnostics...)
	vhidList, diagnostics := carpHealthStatusVHIDList(ctx, remote.VHIDs)
	resp.Diagnostics.Append(diagnostics...)
	routeList, diagnostics := carpHealthStatusRouteList(ctx, remote.Routes)
	resp.Diagnostics.Append(diagnostics...)
	if resp.Diagnostics.HasError() {
		return
	}
	state := &carpHealthStatusModel{
		ID: types.StringValue("carp_health"), Status: types.StringValue(remote.Status), Enabled: types.BoolValue(remote.Enabled),
		Ready: types.BoolValue(remote.Ready), Healthy: types.BoolValue(remote.Healthy), ProbeHealthy: types.BoolValue(remote.ProbeHealthy),
		ControlOK: types.BoolValue(remote.ControlOK), Running: types.BoolValue(remote.Running), Timestamp: types.Float64Value(remote.Timestamp),
		ConfigSignature: types.StringValue(remote.ConfigSignature), Global: global, VHIDs: vhidList, Routes: routeList, Checks: checkList,
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func carpHealthStatusVHIDList(ctx context.Context, items []apiextensions.CarpHealthVHIDStatus) (types.List, diag.Diagnostics) {
	models := make([]carpHealthStatusVHIDModel, 0, len(items))
	for _, item := range items {
		checkIDs, diagnostics := types.ListValueFrom(ctx, types.StringType, item.Checks)
		if diagnostics.HasError() {
			return types.ListNull(types.ObjectType{AttrTypes: carpHealthStatusVHIDTypes}), diagnostics
		}
		models = append(models, carpHealthStatusVHIDModel{
			Key: types.StringValue(item.Key), Interface: types.StringValue(item.Interface), Device: types.StringValue(item.Device),
			VHID: types.Int64Value(int64(item.VHID)), Checks: checkIDs, Ready: types.BoolValue(item.Ready), Healthy: types.BoolValue(item.Healthy),
			DesiredDemoted: types.BoolValue(item.DesiredDemoted), DesiredAdvSkew: nullableInt(item.DesiredAdvSkew),
			ConfiguredAdvSkew: nullableInt(item.ConfiguredAdvSkew), CurrentAdvSkew: nullableInt(item.CurrentAdvSkew),
			CarpState: types.StringValue(item.CarpState), ControlOK: types.BoolValue(item.ControlOK), Retired: types.BoolValue(item.Retired), Error: types.StringValue(item.Error),
		})
	}
	return types.ListValueFrom(ctx, types.ObjectType{AttrTypes: carpHealthStatusVHIDTypes}, models)
}

func carpHealthStatusRouteList(ctx context.Context, items []apiextensions.CarpHealthRouteStatus) (types.List, diag.Diagnostics) {
	models := make([]carpHealthStatusRouteModel, 0, len(items))
	for _, item := range items {
		models = append(models, carpHealthStatusRouteModel{
			Key: types.StringValue(item.Key), CheckUUID: types.StringValue(item.CheckUUID), Check: types.StringValue(item.Check),
			Family: types.StringValue(item.Family), Destination: types.StringValue(item.Destination), Gateway: types.StringValue(item.Gateway),
			DesiredInstalled: types.BoolValue(item.DesiredInstalled), Installed: types.BoolValue(item.Installed), Managed: types.BoolValue(item.Managed),
			ControlOK: types.BoolValue(item.ControlOK), Retired: types.BoolValue(item.Retired), Error: types.StringValue(item.Error),
		})
	}
	return types.ListValueFrom(ctx, types.ObjectType{AttrTypes: carpHealthStatusRouteTypes}, models)
}

func nullableInt(value *int) types.Int64 {
	if value == nil {
		return types.Int64Null()
	}
	return types.Int64Value(int64(*value))
}
