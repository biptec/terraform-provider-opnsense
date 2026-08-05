package bind

import (
	"github.com/biptec/opnsense-go/pkg/api"
	apibind "github.com/biptec/opnsense-go/pkg/bind"
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

type tsigKeyResourceModel struct {
	ID        types.String `tfsdk:"id"`
	Enabled   types.Bool   `tfsdk:"enabled"`
	Name      types.String `tfsdk:"name"`
	Algorithm types.String `tfsdk:"algorithm"`
	Secret    types.String `tfsdk:"secret"`
}

func tsigKeyResourceSchema() schema.Schema {
	return schema.Schema{MarkdownDescription: "Manages a BIND TSIG key for RFC2136 dynamic updates or authenticated transfers. The secret is stored as sensitive Terraform state.", Attributes: map[string]schema.Attribute{
		"id":        schema.StringAttribute{Computed: true, MarkdownDescription: "UUID of the TSIG key.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"enabled":   schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true), MarkdownDescription: "Whether the key is enabled."},
		"name":      schema.StringAttribute{Required: true, MarkdownDescription: "Unique TSIG identity.", Validators: []validator.String{stringvalidator.LengthBetween(1, 255)}},
		"algorithm": schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("hmac-sha256"), MarkdownDescription: "TSIG HMAC algorithm.", Validators: []validator.String{stringvalidator.OneOf("hmac-sha512", "hmac-sha384", "hmac-sha256", "hmac-sha224", "hmac-sha1")}},
		"secret":    schema.StringAttribute{Required: true, Sensitive: true, MarkdownDescription: "Base64-encoded TSIG secret. Supply it from a secret manager and protect the Terraform state."},
	}}
}

func tsigKeyDataSourceSchema() dschema.Schema {
	return dschema.Schema{MarkdownDescription: "Reads a BIND TSIG key.", Attributes: map[string]dschema.Attribute{
		"id":        dschema.StringAttribute{Required: true},
		"enabled":   dschema.BoolAttribute{Computed: true},
		"name":      dschema.StringAttribute{Computed: true},
		"algorithm": dschema.StringAttribute{Computed: true},
		"secret":    dschema.StringAttribute{Computed: true, Sensitive: true},
	}}
}

func tsigKeyModelToAPI(d *tsigKeyResourceModel) (*apibind.TsigKey, error) {
	return &apibind.TsigKey{Enabled: tools.BoolToString(d.Enabled.ValueBool()), Name: d.Name.ValueString(), Algorithm: api.SelectedMap(d.Algorithm.ValueString()), Secret: d.Secret.ValueString()}, nil
}

func tsigKeyAPIToModel(d *apibind.TsigKey) (*tsigKeyResourceModel, error) {
	return &tsigKeyResourceModel{Enabled: types.BoolValue(tools.StringToBool(d.Enabled)), Name: types.StringValue(d.Name), Algorithm: types.StringValue(d.Algorithm.String()), Secret: types.StringValue(d.Secret)}, nil
}
