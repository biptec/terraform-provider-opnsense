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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type aclModel struct {
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	Expression    types.String `tfsdk:"expression"`
	Negate        types.Bool   `tfsdk:"negate"`
	CaseSensitive types.Bool   `tfsdk:"case_sensitive"`
	SSLFCSNI      types.String `tfsdk:"ssl_fc_sni"`
	SSLSNI        types.String `tfsdk:"ssl_sni"`
	SSLSNISub     types.String `tfsdk:"ssl_sni_sub"`
	SSLSNIBeg     types.String `tfsdk:"ssl_sni_beg"`
	SSLSNIEnd     types.String `tfsdk:"ssl_sni_end"`
	SSLSNIReg     types.String `tfsdk:"ssl_sni_reg"`
	SSLHelloType  types.String `tfsdk:"ssl_hello_type"`
	CustomACL     types.String `tfsdk:"custom_acl"`
	Value         types.String `tfsdk:"value"`
	ID            types.String `tfsdk:"id"`
}

func aclResourceSchema() schema.Schema {
	return schema.Schema{Version: 1, MarkdownDescription: "Manages an OPNsense HAProxy ACL. The `ssl_sni*` expressions are suitable for SNI-based L4 TLS passthrough routing.", Attributes: map[string]schema.Attribute{
		"name": schema.StringAttribute{Required: true, MarkdownDescription: "Unique HAProxy ACL name."}, "description": schema.StringAttribute{Optional: true, MarkdownDescription: "Optional description."},
		"expression": schema.StringAttribute{Required: true, MarkdownDescription: "HAProxy ACL expression selector, for example `ssl_sni`, `ssl_sni_reg`, `ssl_hello_type`, or `custom_acl`."},
		"negate":     schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), MarkdownDescription: "Whether to negate the ACL. Defaults to `false`."}, "case_sensitive": schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), MarkdownDescription: "Whether matching is case-sensitive. Defaults to `false`."},
		"ssl_fc_sni": schema.StringAttribute{Optional: true, MarkdownDescription: "Exact SNI value for the `ssl_fc_sni` expression."}, "ssl_sni": schema.StringAttribute{Optional: true, MarkdownDescription: "Exact ClientHello SNI value for the `ssl_sni` expression."},
		"ssl_sni_sub": schema.StringAttribute{Optional: true, MarkdownDescription: "Substring for `ssl_sni_sub`."}, "ssl_sni_beg": schema.StringAttribute{Optional: true, MarkdownDescription: "Prefix for `ssl_sni_beg`."}, "ssl_sni_end": schema.StringAttribute{Optional: true, MarkdownDescription: "Suffix for `ssl_sni_end`."}, "ssl_sni_reg": schema.StringAttribute{Optional: true, MarkdownDescription: "Regular expression for `ssl_sni_reg`."},
		"ssl_hello_type": schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("x1"), Validators: []validator.String{stringvalidator.OneOf("x0", "x1", "x2")}, MarkdownDescription: "TLS hello type used by the `ssl_hello_type` expression. Defaults to `x1` (client hello), matching OPNsense."}, "custom_acl": schema.StringAttribute{Optional: true, MarkdownDescription: "Raw custom ACL expression."}, "value": schema.StringAttribute{Optional: true, MarkdownDescription: "Generic expression value used by applicable ACL types."},
		"id": schema.StringAttribute{Computed: true, MarkdownDescription: "UUID of the HAProxy ACL.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
	}}
}
func aclDataSourceSchema() dschema.Schema {
	return dschema.Schema{MarkdownDescription: "Reads an OPNsense HAProxy ACL.", Attributes: map[string]dschema.Attribute{
		"id": dschema.StringAttribute{Required: true}, "name": dschema.StringAttribute{Computed: true}, "description": dschema.StringAttribute{Computed: true}, "expression": dschema.StringAttribute{Computed: true}, "negate": dschema.BoolAttribute{Computed: true}, "case_sensitive": dschema.BoolAttribute{Computed: true},
		"ssl_fc_sni": dschema.StringAttribute{Computed: true}, "ssl_sni": dschema.StringAttribute{Computed: true}, "ssl_sni_sub": dschema.StringAttribute{Computed: true}, "ssl_sni_beg": dschema.StringAttribute{Computed: true}, "ssl_sni_end": dschema.StringAttribute{Computed: true}, "ssl_sni_reg": dschema.StringAttribute{Computed: true}, "ssl_hello_type": dschema.StringAttribute{Computed: true}, "custom_acl": dschema.StringAttribute{Computed: true}, "value": dschema.StringAttribute{Computed: true},
	}}
}
func aclModelToAPI(d *aclModel) (*apihaproxy.ACL, error) {
	return &apihaproxy.ACL{Name: d.Name.ValueString(), Description: d.Description.ValueString(), Expression: api.SelectedMap(d.Expression.ValueString()), Negate: tools.BoolToString(d.Negate.ValueBool()), CaseSensitive: tools.BoolToString(d.CaseSensitive.ValueBool()), SSLFCSNI: d.SSLFCSNI.ValueString(), SSLSNI: d.SSLSNI.ValueString(), SSLSNISub: d.SSLSNISub.ValueString(), SSLSNIBeg: d.SSLSNIBeg.ValueString(), SSLSNIEnd: d.SSLSNIEnd.ValueString(), SSLSNIReg: d.SSLSNIReg.ValueString(), SSLHelloType: api.SelectedMap(d.SSLHelloType.ValueString()), CustomACL: d.CustomACL.ValueString(), Value: d.Value.ValueString()}, nil
}
func aclAPIToModel(d *apihaproxy.ACL) (*aclModel, error) {
	return &aclModel{Name: types.StringValue(d.Name), Description: tools.StringOrNull(d.Description), Expression: types.StringValue(d.Expression.String()), Negate: types.BoolValue(tools.StringToBool(d.Negate)), CaseSensitive: types.BoolValue(tools.StringToBool(d.CaseSensitive)), SSLFCSNI: tools.StringOrNull(d.SSLFCSNI), SSLSNI: tools.StringOrNull(d.SSLSNI), SSLSNISub: tools.StringOrNull(d.SSLSNISub), SSLSNIBeg: tools.StringOrNull(d.SSLSNIBeg), SSLSNIEnd: tools.StringOrNull(d.SSLSNIEnd), SSLSNIReg: tools.StringOrNull(d.SSLSNIReg), SSLHelloType: tools.StringOrNull(d.SSLHelloType.String()), CustomACL: tools.StringOrNull(d.CustomACL), Value: tools.StringOrNull(d.Value)}, nil
}
