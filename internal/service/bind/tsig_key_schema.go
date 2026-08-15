package bind

import (
	"github.com/biptec/opnsense-go/pkg/api"
	apibind "github.com/biptec/opnsense-go/pkg/bind"
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

type tsigKeyResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Enabled          types.Bool   `tfsdk:"enabled"`
	Name             types.String `tfsdk:"name"`
	Algorithm        types.String `tfsdk:"algorithm"`
	Secret           types.String `tfsdk:"secret"`
	SecretVersion    types.Int64  `tfsdk:"secret_version"`
	SecretConfigured types.Bool   `tfsdk:"secret_configured"`
}

type tsigKeyDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	Enabled          types.Bool   `tfsdk:"enabled"`
	Name             types.String `tfsdk:"name"`
	Algorithm        types.String `tfsdk:"algorithm"`
	SecretConfigured types.Bool   `tfsdk:"secret_configured"`
}

func tsigKeyResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manages a BIND TSIG key for RFC2136 dynamic updates or authenticated transfers. The secret resource attribute is write-only and is never stored in Terraform/OpenTofu plan or state. Use an ephemeral input variable when the source value must also be excluded from saved plan artifacts.",
		Version:             1,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID of the TSIG key.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"enabled": schema.BoolAttribute{
				Optional: true, Computed: true, Default: booldefault.StaticBool(true),
				MarkdownDescription: "Whether the key is enabled.",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Unique TSIG identity. For update_policy=self_txt this must equal the exact TXT owner the key is allowed to update.",
				Validators:          []validator.String{stringvalidator.LengthBetween(1, 255)},
			},
			"algorithm": schema.StringAttribute{
				Optional: true, Computed: true, Default: stringdefault.StaticString("hmac-sha256"),
				MarkdownDescription: "TSIG HMAC algorithm.",
				Validators:          []validator.String{stringvalidator.OneOf("hmac-sha512", "hmac-sha384", "hmac-sha256", "hmac-sha224", "hmac-sha1")},
			},
			"secret": schema.StringAttribute{
				Optional: true, Sensitive: true, WriteOnly: true,
				MarkdownDescription: "Write-only Base64 TSIG secret. Required when creating a key. The resource attribute is never stored in Terraform/OpenTofu plan or state; supply it through an ephemeral variable to keep the source value out of saved plans too. Increment secret_version whenever the secret changes.",
			},
			"secret_version": schema.Int64Attribute{
				Optional: true, Computed: true, Default: int64default.StaticInt64(0),
				MarkdownDescription: "Stateful rotation marker for the write-only secret. Increment whenever secret changes so Terraform/OpenTofu schedules an update.",
				Validators:          []validator.Int64{int64validator.AtLeast(0)},
			},
			"secret_configured": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether OPNsense currently has a non-empty TSIG secret configured. The secret value itself is never returned to state.",
			},
		},
	}
}

func tsigKeyResourceSchemaV0() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Legacy BIND TSIG key schema whose secret was stored as sensitive Terraform state.",
		Version:             0,
		Attributes: map[string]schema.Attribute{
			"id":        schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"enabled":   schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true)},
			"name":      schema.StringAttribute{Required: true, Validators: []validator.String{stringvalidator.LengthBetween(1, 255)}},
			"algorithm": schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("hmac-sha256"), Validators: []validator.String{stringvalidator.OneOf("hmac-sha512", "hmac-sha384", "hmac-sha256", "hmac-sha224", "hmac-sha1")}},
			"secret":    schema.StringAttribute{Required: true, Sensitive: true},
		},
	}
}

func tsigKeyDataSourceSchema() dschema.Schema {
	return dschema.Schema{MarkdownDescription: "Reads BIND TSIG key metadata without exposing the secret.", Attributes: map[string]dschema.Attribute{
		"id":                dschema.StringAttribute{Required: true},
		"enabled":           dschema.BoolAttribute{Computed: true},
		"name":              dschema.StringAttribute{Computed: true},
		"algorithm":         dschema.StringAttribute{Computed: true},
		"secret_configured": dschema.BoolAttribute{Computed: true},
	}}
}

func tsigKeyModelToAPI(d *tsigKeyResourceModel, secret string) *apibind.TsigKey {
	return &apibind.TsigKey{
		Enabled:   tools.BoolToString(d.Enabled.ValueBool()),
		Name:      d.Name.ValueString(),
		Algorithm: api.SelectedMap(d.Algorithm.ValueString()),
		Secret:    secret,
	}
}

func applyTsigKeyModel(remote *apibind.TsigKey, d *tsigKeyResourceModel, secret types.String) {
	remote.Enabled = tools.BoolToString(d.Enabled.ValueBool())
	remote.Name = d.Name.ValueString()
	remote.Algorithm = api.SelectedMap(d.Algorithm.ValueString())
	if !secret.IsNull() && !secret.IsUnknown() {
		remote.Secret = secret.ValueString()
	}
}

func tsigKeyAPIToModel(d *apibind.TsigKey) *tsigKeyResourceModel {
	return &tsigKeyResourceModel{
		Enabled:          types.BoolValue(tools.StringToBool(d.Enabled)),
		Name:             types.StringValue(d.Name),
		Algorithm:        types.StringValue(d.Algorithm.String()),
		Secret:           types.StringNull(),
		SecretVersion:    types.Int64Null(),
		SecretConfigured: types.BoolValue(d.Secret != ""),
	}
}

func tsigKeyAPIToDataSourceModel(d *apibind.TsigKey) (*tsigKeyDataSourceModel, error) {
	return &tsigKeyDataSourceModel{
		Enabled:          types.BoolValue(tools.StringToBool(d.Enabled)),
		Name:             types.StringValue(d.Name),
		Algorithm:        types.StringValue(d.Algorithm.String()),
		SecretConfigured: types.BoolValue(d.Secret != ""),
	}, nil
}
