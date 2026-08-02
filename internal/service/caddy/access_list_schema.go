package caddy

import (
	"github.com/biptec/opnsense-go/pkg/api"
	apicaddy "github.com/biptec/opnsense-go/pkg/caddy"
	"github.com/biptec/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type accessListResourceModel struct {
	ID                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	ClientIPs           types.Set    `tfsdk:"client_ips"`
	Invert              types.Bool   `tfsdk:"invert"`
	HTTPResponseCode    types.Int64  `tfsdk:"http_response_code"`
	HTTPResponseMessage types.String `tfsdk:"http_response_message"`
	RequestMatcher      types.String `tfsdk:"request_matcher"`
	Description         types.String `tfsdk:"description"`
}

func accessListResourceSchema() schema.Schema {
	return schema.Schema{MarkdownDescription: "Manages a Caddy access list.", Version: 1, Attributes: map[string]schema.Attribute{
		"id":                    schema.StringAttribute{Computed: true, MarkdownDescription: "UUID of the access list.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"name":                  schema.StringAttribute{Required: true, MarkdownDescription: "Name of the access list."},
		"client_ips":            schema.SetAttribute{Required: true, ElementType: types.StringType, MarkdownDescription: "Client IP addresses or networks allowed or denied by this access list."},
		"invert":                schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), MarkdownDescription: "When enabled, invert the client IP match. Defaults to `false`."},
		"http_response_code":    schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(-1), MarkdownDescription: "Optional HTTP response status returned for denied requests. `-1` uses the Caddy default.", Validators: []validator.Int64{int64validator.Any(int64validator.OneOf(-1), int64validator.Between(100, 599))}},
		"http_response_message": schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), MarkdownDescription: "Optional response message returned for denied requests. Defaults to `\"\"`."},
		"request_matcher":       schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("client_ip"), MarkdownDescription: "Address matcher used by Caddy: `client_ip` or `remote_ip`. Defaults to `client_ip`.", Validators: []validator.String{stringvalidator.OneOf("client_ip", "remote_ip")}},
		"description":           schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), MarkdownDescription: "Optional description. Defaults to `\"\"`."},
	}}
}

func accessListDataSourceSchema() dschema.Schema {
	return dschema.Schema{MarkdownDescription: "Reads a Caddy access list by UUID.", Attributes: map[string]dschema.Attribute{
		"id":   dschema.StringAttribute{Required: true, MarkdownDescription: "UUID of the access list."},
		"name": dschema.StringAttribute{Computed: true}, "client_ips": dschema.SetAttribute{Computed: true, ElementType: types.StringType},
		"invert": dschema.BoolAttribute{Computed: true}, "http_response_code": dschema.Int64Attribute{Computed: true},
		"http_response_message": dschema.StringAttribute{Computed: true}, "request_matcher": dschema.StringAttribute{Computed: true}, "description": dschema.StringAttribute{Computed: true},
	}}
}

func convertAccessListSchemaToStruct(d *accessListResourceModel) (*apicaddy.AccessList, error) {
	return &apicaddy.AccessList{Name: d.Name.ValueString(), ClientIPs: api.SelectedMapList(tools.SetToStringSlice(d.ClientIPs)), Invert: tools.BoolToString(d.Invert.ValueBool()), HTTPResponseCode: tools.Int64ToStringNegative(d.HTTPResponseCode.ValueInt64()), HTTPResponseMessage: d.HTTPResponseMessage.ValueString(), RequestMatcher: api.SelectedMap(d.RequestMatcher.ValueString()), Description: d.Description.ValueString()}, nil
}
func convertAccessListStructToSchema(d *apicaddy.AccessList) (*accessListResourceModel, error) {
	code := types.Int64Value(-1)
	if d.HTTPResponseCode != "" {
		code = types.Int64Value(tools.StringToInt64(d.HTTPResponseCode))
	}
	return &accessListResourceModel{Name: types.StringValue(d.Name), ClientIPs: tools.StringSliceToSet([]string(d.ClientIPs)), Invert: types.BoolValue(tools.StringToBool(d.Invert)), HTTPResponseCode: code, HTTPResponseMessage: types.StringValue(d.HTTPResponseMessage), RequestMatcher: types.StringValue(d.RequestMatcher.String()), Description: types.StringValue(d.Description)}, nil
}
