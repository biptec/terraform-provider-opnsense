package haproxy

import (
	"github.com/biptec/opnsense-go/pkg/api"
	apihaproxy "github.com/biptec/opnsense-go/pkg/haproxy"
	"github.com/biptec/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
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

type frontendModel struct {
	Enabled        types.Bool   `tfsdk:"enabled"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	Bind           types.Set    `tfsdk:"bind"`
	BindOptions    types.String `tfsdk:"bind_options"`
	Mode           types.String `tfsdk:"mode"`
	DefaultBackend types.String `tfsdk:"default_backend"`
	SSLEnabled     types.Bool   `tfsdk:"ssl_enabled"`
	CustomOptions  types.String `tfsdk:"custom_options"`
	LinkedActions  types.Set    `tfsdk:"linked_actions"`
	ID             types.String `tfsdk:"id"`
}

func frontendResourceSchema() schema.Schema {
	return schema.Schema{Version: 1, MarkdownDescription: "Manages an OPNsense HAProxy frontend. TCP mode can be used for L4 TLS passthrough without terminating TLS on the firewall.", Attributes: map[string]schema.Attribute{
		"enabled":         schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true), MarkdownDescription: "Whether to enable the frontend. Defaults to `true`."},
		"name":            schema.StringAttribute{Required: true, MarkdownDescription: "Unique HAProxy frontend name."},
		"description":     schema.StringAttribute{Optional: true, MarkdownDescription: "Optional description."},
		"bind":            schema.SetAttribute{Required: true, ElementType: types.StringType, Validators: []validator.Set{setvalidator.SizeAtLeast(1)}, MarkdownDescription: "Listen addresses in HAProxy address:port syntax."},
		"bind_options":    schema.StringAttribute{Optional: true, MarkdownDescription: "Optional raw bind options."},
		"mode":            schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("http"), Validators: []validator.String{stringvalidator.OneOf("http", "ssl", "tcp")}, MarkdownDescription: "Frontend mode: `http`, `ssl`, or `tcp`. Use `tcp` for raw TLS passthrough. Defaults to `http`."},
		"default_backend": schema.StringAttribute{Optional: true, MarkdownDescription: "UUID of the default backend."},
		"ssl_enabled":     schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), MarkdownDescription: "Whether HAProxy terminates TLS on this frontend. Keep disabled for L4 passthrough. Defaults to `false`."},
		"custom_options":  schema.StringAttribute{Optional: true, MarkdownDescription: "Optional raw frontend directives, for example TCP inspection directives required before SNI routing."},
		"linked_actions":  schema.SetAttribute{Optional: true, Computed: true, Default: setdefault.StaticValue(tools.EmptySetValue(types.StringType)), ElementType: types.StringType, MarkdownDescription: "HAProxy action UUIDs attached to the frontend."},
		"id":              schema.StringAttribute{Computed: true, MarkdownDescription: "UUID of the HAProxy frontend.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
	}}
}

func frontendDataSourceSchema() dschema.Schema {
	return dschema.Schema{MarkdownDescription: "Reads an OPNsense HAProxy frontend.", Attributes: map[string]dschema.Attribute{
		"id": dschema.StringAttribute{Required: true}, "enabled": dschema.BoolAttribute{Computed: true}, "name": dschema.StringAttribute{Computed: true},
		"description": dschema.StringAttribute{Computed: true}, "bind": dschema.SetAttribute{Computed: true, ElementType: types.StringType},
		"bind_options": dschema.StringAttribute{Computed: true}, "mode": dschema.StringAttribute{Computed: true}, "default_backend": dschema.StringAttribute{Computed: true},
		"ssl_enabled": dschema.BoolAttribute{Computed: true}, "custom_options": dschema.StringAttribute{Computed: true}, "linked_actions": dschema.SetAttribute{Computed: true, ElementType: types.StringType},
	}}
}

func frontendModelToAPI(d *frontendModel) (*apihaproxy.Frontend, error) {
	return &apihaproxy.Frontend{Enabled: tools.BoolToString(d.Enabled.ValueBool()), Name: d.Name.ValueString(), Description: d.Description.ValueString(), Bind: api.SelectedMapList(tools.SetToStringSlice(d.Bind)), BindOptions: d.BindOptions.ValueString(), Mode: api.SelectedMap(d.Mode.ValueString()), DefaultBackend: api.SelectedMap(d.DefaultBackend.ValueString()), SSLEnabled: tools.BoolToString(d.SSLEnabled.ValueBool()), CustomOptions: d.CustomOptions.ValueString(), LinkedActions: api.SelectedMapList(tools.SetToStringSlice(d.LinkedActions))}, nil
}
func frontendAPIToModel(d *apihaproxy.Frontend) (*frontendModel, error) {
	return &frontendModel{Enabled: types.BoolValue(tools.StringToBool(d.Enabled)), Name: types.StringValue(d.Name), Description: tools.StringOrNull(d.Description), Bind: tools.StringSliceToSet([]string(d.Bind)), BindOptions: tools.StringOrNull(d.BindOptions), Mode: types.StringValue(d.Mode.String()), DefaultBackend: tools.StringOrNull(d.DefaultBackend.String()), SSLEnabled: types.BoolValue(tools.StringToBool(d.SSLEnabled)), CustomOptions: tools.StringOrNull(d.CustomOptions), LinkedActions: tools.StringSliceToSet([]string(d.LinkedActions))}, nil
}
