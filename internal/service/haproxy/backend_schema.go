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
	Enabled            types.Bool   `tfsdk:"enabled"`
	Name               types.String `tfsdk:"name"`
	Description        types.String `tfsdk:"description"`
	Mode               types.String `tfsdk:"mode"`
	Algorithm          types.String `tfsdk:"algorithm"`
	LinkedServers      types.Set    `tfsdk:"linked_servers"`
	HealthCheckEnabled types.Bool   `tfsdk:"health_check_enabled"`
	HealthCheck        types.String `tfsdk:"health_check"`
	CheckInterval      types.String `tfsdk:"check_interval"`
	CheckDownInterval  types.String `tfsdk:"check_down_interval"`
	HealthCheckFall    types.Int64  `tfsdk:"health_check_fall"`
	HealthCheckRise    types.Int64  `tfsdk:"health_check_rise"`
	CustomOptions      types.String `tfsdk:"custom_options"`
	ID                 types.String `tfsdk:"id"`
}

func backendResourceSchema() schema.Schema {
	return schema.Schema{Version: 1, MarkdownDescription: "Manages an OPNsense HAProxy backend.", Attributes: map[string]schema.Attribute{
		"enabled":              schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true), MarkdownDescription: "Whether to enable the backend. Defaults to `true`."},
		"name":                 schema.StringAttribute{Required: true, MarkdownDescription: "Unique HAProxy backend name."},
		"description":          schema.StringAttribute{Optional: true, MarkdownDescription: "Optional description."},
		"mode":                 schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("http"), Validators: []validator.String{stringvalidator.OneOf("http", "tcp")}, MarkdownDescription: "Backend mode: `http` or `tcp`. Use `tcp` for TLS passthrough. Defaults to `http`."},
		"algorithm":            schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("source"), Validators: []validator.String{stringvalidator.OneOf("source", "roundrobin", "static-rr", "leastconn", "uri", "random")}, MarkdownDescription: "Load-balancing algorithm. Defaults to `source`."},
		"linked_servers":       schema.SetAttribute{Optional: true, Computed: true, Default: setdefault.StaticValue(tools.EmptySetValue(types.StringType)), ElementType: types.StringType, MarkdownDescription: "HAProxy server UUIDs in this backend."},
		"health_check_enabled": schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true), MarkdownDescription: "Whether to enable backend health checks. Defaults to `true`."},
		"health_check":         schema.StringAttribute{Optional: true, MarkdownDescription: "UUID of the HAProxy health check."},
		"check_interval":       schema.StringAttribute{Optional: true, MarkdownDescription: "Health check interval, such as `5s`."},
		"check_down_interval":  schema.StringAttribute{Optional: true, MarkdownDescription: "Health check interval while a server is down."},
		"health_check_fall":    schema.Int64Attribute{Optional: true, Validators: []validator.Int64{int64validator.AtLeast(1)}, MarkdownDescription: "Consecutive failed checks required before marking a server down."},
		"health_check_rise":    schema.Int64Attribute{Optional: true, Validators: []validator.Int64{int64validator.AtLeast(1)}, MarkdownDescription: "Consecutive successful checks required before marking a server up."},
		"custom_options":       schema.StringAttribute{Optional: true, MarkdownDescription: "Optional raw backend directives."},
		"id":                   schema.StringAttribute{Computed: true, MarkdownDescription: "UUID of the HAProxy backend.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
	}}
}
func backendDataSourceSchema() dschema.Schema {
	return dschema.Schema{MarkdownDescription: "Reads an OPNsense HAProxy backend.", Attributes: map[string]dschema.Attribute{
		"id": dschema.StringAttribute{Required: true}, "enabled": dschema.BoolAttribute{Computed: true}, "name": dschema.StringAttribute{Computed: true}, "description": dschema.StringAttribute{Computed: true},
		"mode": dschema.StringAttribute{Computed: true}, "algorithm": dschema.StringAttribute{Computed: true}, "linked_servers": dschema.SetAttribute{Computed: true, ElementType: types.StringType},
		"health_check_enabled": dschema.BoolAttribute{Computed: true}, "health_check": dschema.StringAttribute{Computed: true}, "check_interval": dschema.StringAttribute{Computed: true}, "check_down_interval": dschema.StringAttribute{Computed: true}, "health_check_fall": dschema.Int64Attribute{Computed: true}, "health_check_rise": dschema.Int64Attribute{Computed: true}, "custom_options": dschema.StringAttribute{Computed: true},
	}}
}
func backendModelToAPI(d *backendModel) (*apihaproxy.Backend, error) {
	return &apihaproxy.Backend{Enabled: tools.BoolToString(d.Enabled.ValueBool()), Name: d.Name.ValueString(), Description: d.Description.ValueString(), Mode: api.SelectedMap(d.Mode.ValueString()), Algorithm: api.SelectedMap(d.Algorithm.ValueString()), LinkedServers: api.SelectedMapList(tools.SetToStringSlice(d.LinkedServers)), HealthCheckEnabled: tools.BoolToString(d.HealthCheckEnabled.ValueBool()), HealthCheck: api.SelectedMap(d.HealthCheck.ValueString()), CheckInterval: d.CheckInterval.ValueString(), CheckDownInterval: d.CheckDownInterval.ValueString(), HealthCheckFall: optionalIntString(d.HealthCheckFall), HealthCheckRise: optionalIntString(d.HealthCheckRise), CustomOptions: d.CustomOptions.ValueString()}, nil
}
func backendAPIToModel(d *apihaproxy.Backend) (*backendModel, error) {
	return &backendModel{Enabled: types.BoolValue(tools.StringToBool(d.Enabled)), Name: types.StringValue(d.Name), Description: tools.StringOrNull(d.Description), Mode: types.StringValue(d.Mode.String()), Algorithm: types.StringValue(d.Algorithm.String()), LinkedServers: tools.StringSliceToSet([]string(d.LinkedServers)), HealthCheckEnabled: types.BoolValue(tools.StringToBool(d.HealthCheckEnabled)), HealthCheck: tools.StringOrNull(d.HealthCheck.String()), CheckInterval: tools.StringOrNull(d.CheckInterval), CheckDownInterval: tools.StringOrNull(d.CheckDownInterval), HealthCheckFall: tools.StringToInt64Null(d.HealthCheckFall), HealthCheckRise: tools.StringToInt64Null(d.HealthCheckRise), CustomOptions: tools.StringOrNull(d.CustomOptions)}, nil
}
