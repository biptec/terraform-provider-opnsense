package haproxy

import (
	"github.com/biptec/opnsense-go/pkg/api"
	apihaproxy "github.com/biptec/opnsense-go/pkg/haproxy"
	"github.com/biptec/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type backendModel struct {
	Enabled                  types.Bool   `tfsdk:"enabled"`
	Name                     types.String `tfsdk:"name"`
	Description              types.String `tfsdk:"description"`
	HAPolicy                 types.String `tfsdk:"ha_policy"`
	Mode                     types.String `tfsdk:"mode"`
	Algorithm                types.String `tfsdk:"algorithm"`
	ProxyProtocol            types.String `tfsdk:"proxy_protocol"`
	LinkedServers            types.Set    `tfsdk:"linked_servers"`
	HealthCheckEnabled       types.Bool   `tfsdk:"health_check_enabled"`
	HealthCheck              types.String `tfsdk:"health_check"`
	HealthCheckProxyProtocol types.String `tfsdk:"health_check_proxy_protocol"`
	CheckInterval            types.String `tfsdk:"check_interval"`
	CheckDownInterval        types.String `tfsdk:"check_down_interval"`
	HealthCheckFall          types.Int64  `tfsdk:"health_check_fall"`
	HealthCheckRise          types.Int64  `tfsdk:"health_check_rise"`
	CustomOptions            types.String `tfsdk:"custom_options"`
	ID                       types.String `tfsdk:"id"`
}

func backendResourceSchema() schema.Schema {
	return schema.Schema{Version: 1, MarkdownDescription: "Manages an OPNsense HAProxy backend.", Attributes: map[string]schema.Attribute{
		"enabled":                     schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true), MarkdownDescription: "Whether to enable the backend. Defaults to `true`."},
		"name":                        schema.StringAttribute{Required: true, MarkdownDescription: "Unique HAProxy backend name."},
		"description":                 schema.StringAttribute{Optional: true, MarkdownDescription: "Optional description."},
		"ha_policy":                   schema.StringAttribute{Optional: true, MarkdownDescription: "Optional HA policy ID used by the policy-managed HAProxy synchronizer."},
		"mode":                        schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("http"), Validators: []validator.String{stringvalidator.OneOf("http", "tcp")}, MarkdownDescription: "Backend mode: `http` or `tcp`. Use `tcp` for TLS passthrough. Defaults to `http`."},
		"algorithm":                   schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("source"), Validators: []validator.String{stringvalidator.OneOf("source", "roundrobin", "static-rr", "leastconn", "uri", "random")}, MarkdownDescription: "Load-balancing algorithm. Defaults to `source`."},
		"proxy_protocol":              schema.StringAttribute{Optional: true, Validators: []validator.String{stringvalidator.OneOf("v1", "v2")}, MarkdownDescription: "PROXY protocol version sent to backend servers. Omit to disable PROXY protocol."},
		"linked_servers":              schema.SetAttribute{Optional: true, Computed: true, Default: setdefault.StaticValue(tools.EmptySetValue(types.StringType)), ElementType: types.StringType, MarkdownDescription: "HAProxy server UUIDs in this backend."},
		"health_check_enabled":        schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true), MarkdownDescription: "Whether to enable backend health checks. Defaults to `true`."},
		"health_check":                schema.StringAttribute{Optional: true, MarkdownDescription: "UUID of the HAProxy health check."},
		"health_check_proxy_protocol": schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("backend"), Validators: []validator.String{stringvalidator.OneOf("backend", "enable", "disable")}, MarkdownDescription: "PROXY behavior for health checks: follow the backend setting, force enable, or disable. Defaults to `backend`."},
		"check_interval":              schema.StringAttribute{Optional: true, MarkdownDescription: "Health check interval, such as `5s`."},
		"check_down_interval":         schema.StringAttribute{Optional: true, MarkdownDescription: "Health check interval while a server is down."},
		"health_check_fall":           schema.Int64Attribute{Optional: true, Validators: []validator.Int64{int64validator.AtLeast(1)}, MarkdownDescription: "Consecutive failed checks required before marking a server down."},
		"health_check_rise":           schema.Int64Attribute{Optional: true, Validators: []validator.Int64{int64validator.AtLeast(1)}, MarkdownDescription: "Consecutive successful checks required before marking a server up."},
		"custom_options":              schema.StringAttribute{Optional: true, MarkdownDescription: "Optional raw backend directives."},
		"id":                          schema.StringAttribute{Computed: true, MarkdownDescription: "UUID of the HAProxy backend.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
	}}
}
func backendDataSourceSchema() dschema.Schema {
	return dschema.Schema{MarkdownDescription: "Reads an OPNsense HAProxy backend.", Attributes: map[string]dschema.Attribute{
		"id": dschema.StringAttribute{Required: true}, "enabled": dschema.BoolAttribute{Computed: true}, "name": dschema.StringAttribute{Computed: true}, "description": dschema.StringAttribute{Computed: true}, "ha_policy": dschema.StringAttribute{Computed: true},
		"mode": dschema.StringAttribute{Computed: true}, "algorithm": dschema.StringAttribute{Computed: true}, "proxy_protocol": dschema.StringAttribute{Computed: true}, "linked_servers": dschema.SetAttribute{Computed: true, ElementType: types.StringType},
		"health_check_enabled": dschema.BoolAttribute{Computed: true}, "health_check": dschema.StringAttribute{Computed: true}, "health_check_proxy_protocol": dschema.StringAttribute{Computed: true}, "check_interval": dschema.StringAttribute{Computed: true}, "check_down_interval": dschema.StringAttribute{Computed: true}, "health_check_fall": dschema.Int64Attribute{Computed: true}, "health_check_rise": dschema.Int64Attribute{Computed: true}, "custom_options": dschema.StringAttribute{Computed: true},
	}}
}
func backendModelToAPI(d *backendModel) (*apihaproxy.Backend, error) {
	return &apihaproxy.Backend{Enabled: tools.BoolToString(d.Enabled.ValueBool()), Name: d.Name.ValueString(), Description: d.Description.ValueString(), HAPolicy: api.SelectedMap(d.HAPolicy.ValueString()), Mode: api.SelectedMap(d.Mode.ValueString()), Algorithm: api.SelectedMap(d.Algorithm.ValueString()), ProxyProtocol: api.SelectedMap(d.ProxyProtocol.ValueString()), LinkedServers: api.SelectedMapList(tools.SetToStringSlice(d.LinkedServers)), HealthCheckEnabled: tools.BoolToString(d.HealthCheckEnabled.ValueBool()), HealthCheck: api.SelectedMap(d.HealthCheck.ValueString()), HealthCheckProxyProto: api.SelectedMap(d.HealthCheckProxyProtocol.ValueString()), CheckInterval: d.CheckInterval.ValueString(), CheckDownInterval: d.CheckDownInterval.ValueString(), HealthCheckFall: optionalIntString(d.HealthCheckFall), HealthCheckRise: optionalIntString(d.HealthCheckRise), CustomOptions: d.CustomOptions.ValueString()}, nil
}
func backendAPIToModel(d *apihaproxy.Backend) (*backendModel, error) {
	return &backendModel{Enabled: types.BoolValue(tools.StringToBool(d.Enabled)), Name: types.StringValue(d.Name), Description: tools.StringOrNull(d.Description), HAPolicy: tools.StringOrNull(d.HAPolicy.String()), Mode: types.StringValue(d.Mode.String()), Algorithm: types.StringValue(d.Algorithm.String()), ProxyProtocol: tools.StringOrNull(d.ProxyProtocol.String()), LinkedServers: tools.StringSliceToSet([]string(d.LinkedServers)), HealthCheckEnabled: types.BoolValue(tools.StringToBool(d.HealthCheckEnabled)), HealthCheck: tools.StringOrNull(d.HealthCheck.String()), HealthCheckProxyProtocol: tools.StringOrNull(d.HealthCheckProxyProto.String()), CheckInterval: tools.StringOrNull(d.CheckInterval), CheckDownInterval: tools.StringOrNull(d.CheckDownInterval), HealthCheckFall: tools.StringToInt64Null(d.HealthCheckFall), HealthCheckRise: tools.StringToInt64Null(d.HealthCheckRise), CustomOptions: tools.StringOrNull(d.CustomOptions)}, nil
}
