package haproxy

import (
	"strconv"

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

type serverModel struct {
	Enabled           types.Bool   `tfsdk:"enabled"`
	Name              types.String `tfsdk:"name"`
	Description       types.String `tfsdk:"description"`
	Address           types.String `tfsdk:"address"`
	Port              types.Int64  `tfsdk:"port"`
	CheckPort         types.Int64  `tfsdk:"check_port"`
	Mode              types.String `tfsdk:"mode"`
	Type              types.String `tfsdk:"type"`
	SSL               types.Bool   `tfsdk:"ssl"`
	MaxConnections    types.Int64  `tfsdk:"max_connections"`
	Weight            types.Int64  `tfsdk:"weight"`
	CheckInterval     types.String `tfsdk:"check_interval"`
	CheckDownInterval types.String `tfsdk:"check_down_interval"`
	Advanced          types.String `tfsdk:"advanced"`
	ID                types.String `tfsdk:"id"`
}

func serverResourceSchema() schema.Schema {
	portValidator := []validator.Int64{int64validator.Between(1, 65535)}
	nonNegative := []validator.Int64{int64validator.AtLeast(0)}
	return schema.Schema{Version: 1, MarkdownDescription: "Manages an OPNsense HAProxy server endpoint.", Attributes: map[string]schema.Attribute{
		"enabled": schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true), MarkdownDescription: "Whether to enable the server. Defaults to `true`."},
		"name":    schema.StringAttribute{Required: true, MarkdownDescription: "Unique HAProxy server name."}, "description": schema.StringAttribute{Optional: true, MarkdownDescription: "Optional description."},
		"address":         schema.StringAttribute{Required: true, MarkdownDescription: "Backend server address or hostname."},
		"port":            schema.Int64Attribute{Optional: true, Validators: portValidator, MarkdownDescription: "Backend server TCP port."},
		"check_port":      schema.Int64Attribute{Optional: true, Validators: portValidator, MarkdownDescription: "Optional dedicated health-check port."},
		"mode":            schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("active"), Validators: []validator.String{stringvalidator.OneOf("active", "backup", "disabled")}, MarkdownDescription: "Server mode: `active`, `backup`, or `disabled`. Defaults to `active`."},
		"type":            schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("static"), Validators: []validator.String{stringvalidator.OneOf("static", "template", "unix")}, MarkdownDescription: "Server type. Defaults to `static`."},
		"ssl":             schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), MarkdownDescription: "Whether HAProxy establishes TLS to the backend. Keep disabled for raw TCP/TLS passthrough. Defaults to `false`."},
		"max_connections": schema.Int64Attribute{Optional: true, Validators: nonNegative, MarkdownDescription: "Optional maximum concurrent connections."},
		"weight":          schema.Int64Attribute{Optional: true, Validators: nonNegative, MarkdownDescription: "Optional server weight."},
		"check_interval":  schema.StringAttribute{Optional: true, MarkdownDescription: "Optional per-server health-check interval."}, "check_down_interval": schema.StringAttribute{Optional: true, MarkdownDescription: "Optional per-server health-check interval while down."},
		"advanced": schema.StringAttribute{Optional: true, MarkdownDescription: "Optional raw HAProxy server parameters."},
		"id":       schema.StringAttribute{Computed: true, MarkdownDescription: "UUID of the HAProxy server.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
	}}
}
func serverDataSourceSchema() dschema.Schema {
	return dschema.Schema{MarkdownDescription: "Reads an OPNsense HAProxy server.", Attributes: map[string]dschema.Attribute{
		"id": dschema.StringAttribute{Required: true}, "enabled": dschema.BoolAttribute{Computed: true}, "name": dschema.StringAttribute{Computed: true}, "description": dschema.StringAttribute{Computed: true}, "address": dschema.StringAttribute{Computed: true},
		"port": dschema.Int64Attribute{Computed: true}, "check_port": dschema.Int64Attribute{Computed: true}, "mode": dschema.StringAttribute{Computed: true}, "type": dschema.StringAttribute{Computed: true}, "ssl": dschema.BoolAttribute{Computed: true},
		"max_connections": dschema.Int64Attribute{Computed: true}, "weight": dschema.Int64Attribute{Computed: true}, "check_interval": dschema.StringAttribute{Computed: true}, "check_down_interval": dschema.StringAttribute{Computed: true}, "advanced": dschema.StringAttribute{Computed: true},
	}}
}
func optionalIntString(v types.Int64) string {
	if v.IsNull() || v.IsUnknown() {
		return ""
	}
	return strconv.FormatInt(v.ValueInt64(), 10)
}
func serverModelToAPI(d *serverModel) (*apihaproxy.Server, error) {
	return &apihaproxy.Server{Enabled: tools.BoolToString(d.Enabled.ValueBool()), Name: d.Name.ValueString(), Description: d.Description.ValueString(), Address: d.Address.ValueString(), Port: optionalIntString(d.Port), CheckPort: optionalIntString(d.CheckPort), Mode: api.SelectedMap(d.Mode.ValueString()), Type: api.SelectedMap(d.Type.ValueString()), SSL: tools.BoolToString(d.SSL.ValueBool()), MaxConnections: optionalIntString(d.MaxConnections), Weight: optionalIntString(d.Weight), CheckInterval: d.CheckInterval.ValueString(), CheckDownInterval: d.CheckDownInterval.ValueString(), Advanced: d.Advanced.ValueString()}, nil
}
func serverAPIToModel(d *apihaproxy.Server) (*serverModel, error) {
	return &serverModel{Enabled: types.BoolValue(tools.StringToBool(d.Enabled)), Name: types.StringValue(d.Name), Description: tools.StringOrNull(d.Description), Address: types.StringValue(d.Address), Port: tools.StringToInt64Null(d.Port), CheckPort: tools.StringToInt64Null(d.CheckPort), Mode: types.StringValue(d.Mode.String()), Type: types.StringValue(d.Type.String()), SSL: types.BoolValue(tools.StringToBool(d.SSL)), MaxConnections: tools.StringToInt64Null(d.MaxConnections), Weight: tools.StringToInt64Null(d.Weight), CheckInterval: tools.StringOrNull(d.CheckInterval), CheckDownInterval: tools.StringOrNull(d.CheckDownInterval), Advanced: tools.StringOrNull(d.Advanced)}, nil
}
