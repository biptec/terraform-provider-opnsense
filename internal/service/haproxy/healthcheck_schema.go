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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type healthcheckModel struct {
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	Type          types.String `tfsdk:"type"`
	Interval      types.String `tfsdk:"interval"`
	SSL           types.String `tfsdk:"ssl"`
	SSLSNI        types.String `tfsdk:"ssl_sni"`
	ForceSSL      types.Bool   `tfsdk:"force_ssl"`
	CheckPort     types.Int64  `tfsdk:"check_port"`
	HTTPMethod    types.String `tfsdk:"http_method"`
	HTTPURI       types.String `tfsdk:"http_uri"`
	HTTPHost      types.String `tfsdk:"http_host"`
	TCPSendValue  types.String `tfsdk:"tcp_send_value"`
	TCPMatchType  types.String `tfsdk:"tcp_match_type"`
	TCPMatchValue types.String `tfsdk:"tcp_match_value"`
	ID            types.String `tfsdk:"id"`
}

func healthcheckResourceSchema() schema.Schema {
	return schema.Schema{Version: 1, MarkdownDescription: "Manages an OPNsense HAProxy health check.", Attributes: map[string]schema.Attribute{
		"name": schema.StringAttribute{Required: true, MarkdownDescription: "Unique health-check name."}, "description": schema.StringAttribute{Optional: true, MarkdownDescription: "Optional description."},
		"type":     schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("http"), Validators: []validator.String{stringvalidator.OneOf("tcp", "http", "agent", "ldap", "mysql", "pgsql", "redis", "smtp", "esmtp", "ssl")}, MarkdownDescription: "Health-check type. Use `tcp` for L4 services. Defaults to `http`."},
		"interval": schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("2s"), MarkdownDescription: "Health-check interval. Defaults to `2s`."},
		"ssl":      schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("nopref"), Validators: []validator.String{stringvalidator.OneOf("nopref", "ssl", "sslsni", "nossl")}, MarkdownDescription: "TLS preference for the check. Defaults to `nopref`."},
		"ssl_sni":  schema.StringAttribute{Optional: true, MarkdownDescription: "Optional SNI name sent by an SSL health check."}, "force_ssl": schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), MarkdownDescription: "Whether to force SSL for the health check. Defaults to `false`."},
		"check_port":  schema.Int64Attribute{Optional: true, Validators: []validator.Int64{int64validator.Between(1, 65535)}, MarkdownDescription: "Optional dedicated check port."},
		"http_method": schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("options"), Validators: []validator.String{stringvalidator.OneOf("options", "head", "get", "put", "post", "delete", "trace")}, MarkdownDescription: "HTTP method for HTTP checks. Defaults to `options`, matching OPNsense."}, "http_uri": schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("/"), MarkdownDescription: "URI for HTTP checks. Defaults to `/`, matching OPNsense."}, "http_host": schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("localhost"), MarkdownDescription: "Host header for HTTP checks. Defaults to `localhost`, matching OPNsense."},
		"tcp_send_value": schema.StringAttribute{Optional: true, MarkdownDescription: "Optional value to send during a TCP check."}, "tcp_match_type": schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("string"), Validators: []validator.String{stringvalidator.OneOf("string", "rstring", "binary")}, MarkdownDescription: "TCP response match type. Defaults to `string`, matching OPNsense."}, "tcp_match_value": schema.StringAttribute{Optional: true, MarkdownDescription: "Optional expected TCP response value."},
		"id": schema.StringAttribute{Computed: true, MarkdownDescription: "UUID of the HAProxy health check.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
	}}
}
func healthcheckDataSourceSchema() dschema.Schema {
	return dschema.Schema{MarkdownDescription: "Reads an OPNsense HAProxy health check.", Attributes: map[string]dschema.Attribute{
		"id": dschema.StringAttribute{Required: true}, "name": dschema.StringAttribute{Computed: true}, "description": dschema.StringAttribute{Computed: true}, "type": dschema.StringAttribute{Computed: true}, "interval": dschema.StringAttribute{Computed: true}, "ssl": dschema.StringAttribute{Computed: true}, "ssl_sni": dschema.StringAttribute{Computed: true}, "force_ssl": dschema.BoolAttribute{Computed: true}, "check_port": dschema.Int64Attribute{Computed: true}, "http_method": dschema.StringAttribute{Computed: true}, "http_uri": dschema.StringAttribute{Computed: true}, "http_host": dschema.StringAttribute{Computed: true}, "tcp_send_value": dschema.StringAttribute{Computed: true}, "tcp_match_type": dschema.StringAttribute{Computed: true}, "tcp_match_value": dschema.StringAttribute{Computed: true},
	}}
}
func healthcheckModelToAPI(d *healthcheckModel) (*apihaproxy.Healthcheck, error) {
	return &apihaproxy.Healthcheck{Name: d.Name.ValueString(), Description: d.Description.ValueString(), Type: api.SelectedMap(d.Type.ValueString()), Interval: d.Interval.ValueString(), SSL: api.SelectedMap(d.SSL.ValueString()), SSLSNI: d.SSLSNI.ValueString(), ForceSSL: tools.BoolToString(d.ForceSSL.ValueBool()), CheckPort: optionalIntString(d.CheckPort), HTTPMethod: api.SelectedMap(d.HTTPMethod.ValueString()), HTTPURI: d.HTTPURI.ValueString(), HTTPHost: d.HTTPHost.ValueString(), TCPSendValue: d.TCPSendValue.ValueString(), TCPMatchType: api.SelectedMap(d.TCPMatchType.ValueString()), TCPMatchValue: d.TCPMatchValue.ValueString()}, nil
}
func healthcheckAPIToModel(d *apihaproxy.Healthcheck) (*healthcheckModel, error) {
	return &healthcheckModel{Name: types.StringValue(d.Name), Description: tools.StringOrNull(d.Description), Type: types.StringValue(d.Type.String()), Interval: types.StringValue(d.Interval), SSL: types.StringValue(d.SSL.String()), SSLSNI: tools.StringOrNull(d.SSLSNI), ForceSSL: types.BoolValue(tools.StringToBool(d.ForceSSL)), CheckPort: tools.StringToInt64Null(d.CheckPort), HTTPMethod: tools.StringOrNull(d.HTTPMethod.String()), HTTPURI: tools.StringOrNull(d.HTTPURI), HTTPHost: tools.StringOrNull(d.HTTPHost), TCPSendValue: tools.StringOrNull(d.TCPSendValue), TCPMatchType: tools.StringOrNull(d.TCPMatchType.String()), TCPMatchValue: tools.StringOrNull(d.TCPMatchValue)}, nil
}
