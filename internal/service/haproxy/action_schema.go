package haproxy

import (
	"github.com/biptec/opnsense-go/pkg/api"
	apihaproxy "github.com/biptec/opnsense-go/pkg/haproxy"
	"github.com/biptec/terraform-provider-opnsense/internal/tools"
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

type actionModel struct {
	Enabled                types.Bool   `tfsdk:"enabled"`
	Name                   types.String `tfsdk:"name"`
	Description            types.String `tfsdk:"description"`
	TestType               types.String `tfsdk:"test_type"`
	LinkedACLs             types.Set    `tfsdk:"linked_acls"`
	Operator               types.String `tfsdk:"operator"`
	Type                   types.String `tfsdk:"type"`
	UseBackend             types.String `tfsdk:"use_backend"`
	Custom                 types.String `tfsdk:"custom"`
	TCPRequestAction       types.String `tfsdk:"tcp_request_action"`
	TCPRequestOption       types.String `tfsdk:"tcp_request_option"`
	TCPRequestInspectDelay types.String `tfsdk:"tcp_request_inspect_delay"`
	ID                     types.String `tfsdk:"id"`
}

func actionResourceSchema() schema.Schema {
	return schema.Schema{Version: 1, MarkdownDescription: "Manages an OPNsense HAProxy action, including conditional `use_backend` routing used by SNI passthrough frontends.", Attributes: map[string]schema.Attribute{
		"enabled": schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true), MarkdownDescription: "Whether to enable the action. Defaults to `true`."}, "name": schema.StringAttribute{Required: true, MarkdownDescription: "Unique action name."}, "description": schema.StringAttribute{Optional: true, MarkdownDescription: "Optional description."},
		"test_type":   schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("if"), Validators: []validator.String{stringvalidator.OneOf("if", "unless")}, MarkdownDescription: "ACL test type. Defaults to `if`."},
		"linked_acls": schema.SetAttribute{Optional: true, Computed: true, Default: setdefault.StaticValue(tools.EmptySetValue(types.StringType)), ElementType: types.StringType, MarkdownDescription: "ACL UUIDs evaluated by this action."},
		"operator":    schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("and"), Validators: []validator.String{stringvalidator.OneOf("and", "or")}, MarkdownDescription: "Operator joining linked ACLs. Defaults to `and`."},
		"type":        schema.StringAttribute{Required: true, MarkdownDescription: "HAProxy action type, for example `use_backend`, `tcp-request`, or `custom`."}, "use_backend": schema.StringAttribute{Optional: true, MarkdownDescription: "Backend UUID for a `use_backend` action."}, "custom": schema.StringAttribute{Optional: true, MarkdownDescription: "Raw action content for a `custom` action."},
		"tcp_request_action": schema.StringAttribute{Optional: true, MarkdownDescription: "TCP request action selector."}, "tcp_request_option": schema.StringAttribute{Optional: true, MarkdownDescription: "TCP request action option."}, "tcp_request_inspect_delay": schema.StringAttribute{Optional: true, MarkdownDescription: "TCP inspection delay used by applicable request actions."},
		"id": schema.StringAttribute{Computed: true, MarkdownDescription: "UUID of the HAProxy action.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
	}}
}
func actionDataSourceSchema() dschema.Schema {
	return dschema.Schema{MarkdownDescription: "Reads an OPNsense HAProxy action.", Attributes: map[string]dschema.Attribute{
		"id": dschema.StringAttribute{Required: true}, "enabled": dschema.BoolAttribute{Computed: true}, "name": dschema.StringAttribute{Computed: true}, "description": dschema.StringAttribute{Computed: true}, "test_type": dschema.StringAttribute{Computed: true}, "linked_acls": dschema.SetAttribute{Computed: true, ElementType: types.StringType}, "operator": dschema.StringAttribute{Computed: true}, "type": dschema.StringAttribute{Computed: true}, "use_backend": dschema.StringAttribute{Computed: true}, "custom": dschema.StringAttribute{Computed: true}, "tcp_request_action": dschema.StringAttribute{Computed: true}, "tcp_request_option": dschema.StringAttribute{Computed: true}, "tcp_request_inspect_delay": dschema.StringAttribute{Computed: true},
	}}
}
func actionModelToAPI(d *actionModel) (*apihaproxy.Action, error) {
	return &apihaproxy.Action{Enabled: tools.BoolToString(d.Enabled.ValueBool()), Name: d.Name.ValueString(), Description: d.Description.ValueString(), TestType: api.SelectedMap(d.TestType.ValueString()), LinkedACLs: api.SelectedMapList(tools.SetToStringSlice(d.LinkedACLs)), Operator: api.SelectedMap(d.Operator.ValueString()), Type: api.SelectedMap(d.Type.ValueString()), UseBackend: api.SelectedMap(d.UseBackend.ValueString()), Custom: d.Custom.ValueString(), TCPRequestAction: api.SelectedMap(d.TCPRequestAction.ValueString()), TCPRequestOption: d.TCPRequestOption.ValueString(), TCPRequestInspectDelay: d.TCPRequestInspectDelay.ValueString()}, nil
}
func actionAPIToModel(d *apihaproxy.Action) (*actionModel, error) {
	return &actionModel{Enabled: types.BoolValue(tools.StringToBool(d.Enabled)), Name: types.StringValue(d.Name), Description: tools.StringOrNull(d.Description), TestType: types.StringValue(d.TestType.String()), LinkedACLs: tools.StringSliceToSet([]string(d.LinkedACLs)), Operator: types.StringValue(d.Operator.String()), Type: types.StringValue(d.Type.String()), UseBackend: tools.StringOrNull(d.UseBackend.String()), Custom: tools.StringOrNull(d.Custom), TCPRequestAction: tools.StringOrNull(d.TCPRequestAction.String()), TCPRequestOption: tools.StringOrNull(d.TCPRequestOption), TCPRequestInspectDelay: tools.StringOrNull(d.TCPRequestInspectDelay)}, nil
}
