package bind

import (
	"github.com/biptec/opnsense-go/pkg/api"
	apibind "github.com/biptec/opnsense-go/pkg/bind"
	"github.com/biptec/terraform-provider-opnsense/internal/tools"
	"github.com/biptec/terraform-provider-opnsense/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type secondaryDomainResourceModel struct {
	ID                    types.String `tfsdk:"id"`
	ViewID                types.String `tfsdk:"view_id"`
	DomainName            types.String `tfsdk:"domain_name"`
	Enabled               types.Bool   `tfsdk:"enabled"`
	PrimaryIPs            types.Set    `tfsdk:"primary_ips"`
	AllowNotify           types.Set    `tfsdk:"allow_notify"`
	TransferKeyID         types.String `tfsdk:"transfer_key_id"`
	TransferKeyAlgorithm  types.String `tfsdk:"transfer_key_algorithm"`
	TransferKeyName       types.String `tfsdk:"transfer_key_name"`
	TransferKey           types.String `tfsdk:"transfer_key"`
	TransferKeyVersion    types.Int64  `tfsdk:"transfer_key_version"`
	TransferKeyConfigured types.Bool   `tfsdk:"transfer_key_configured"`
	AllowTransferACLs     types.Set    `tfsdk:"allow_transfer_acl_ids"`
	AllowQueryACLs        types.Set    `tfsdk:"allow_query_acl_ids"`
}

type secondaryDomainResourceModelV0 struct {
	ID                   types.String `tfsdk:"id"`
	ViewID               types.String `tfsdk:"view_id"`
	DomainName           types.String `tfsdk:"domain_name"`
	Enabled              types.Bool   `tfsdk:"enabled"`
	PrimaryIPs           types.Set    `tfsdk:"primary_ips"`
	AllowNotify          types.Set    `tfsdk:"allow_notify"`
	TransferKeyAlgorithm types.String `tfsdk:"transfer_key_algorithm"`
	TransferKeyName      types.String `tfsdk:"transfer_key_name"`
	TransferKey          types.String `tfsdk:"transfer_key"`
	AllowTransferACLs    types.Set    `tfsdk:"allow_transfer_acl_ids"`
	AllowQueryACLs       types.Set    `tfsdk:"allow_query_acl_ids"`
}

type secondaryDomainDataSourceModel struct {
	ID                    types.String `tfsdk:"id"`
	ViewID                types.String `tfsdk:"view_id"`
	DomainName            types.String `tfsdk:"domain_name"`
	Enabled               types.Bool   `tfsdk:"enabled"`
	PrimaryIPs            types.Set    `tfsdk:"primary_ips"`
	AllowNotify           types.Set    `tfsdk:"allow_notify"`
	TransferKeyID         types.String `tfsdk:"transfer_key_id"`
	TransferKeyAlgorithm  types.String `tfsdk:"transfer_key_algorithm"`
	TransferKeyName       types.String `tfsdk:"transfer_key_name"`
	TransferKey           types.String `tfsdk:"transfer_key"`
	TransferKeyConfigured types.Bool   `tfsdk:"transfer_key_configured"`
	AllowTransferACLs     types.Set    `tfsdk:"allow_transfer_acl_ids"`
	AllowQueryACLs        types.Set    `tfsdk:"allow_query_acl_ids"`
}

func secondaryDomainResourceSchema() schema.Schema {
	uuidSet := []validator.Set{setvalidator.ValueStringsAre(validators.IsUUIDv4())}
	ipSet := []validator.Set{setvalidator.ValueStringsAre(validators.IPAddress())}
	return schema.Schema{
		Version:             1,
		MarkdownDescription: "Manages a secondary BIND zone inside a selected view. Transfer TSIG material is write-only and is never stored in Terraform/OpenTofu plan or state.",
		Attributes: map[string]schema.Attribute{
			"id":           schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"view_id":      schema.StringAttribute{Required: true, Validators: []validator.String{validators.IsUUIDv4()}},
			"domain_name":  schema.StringAttribute{Required: true, Validators: []validator.String{stringvalidator.LengthBetween(1, 255)}},
			"enabled":      schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true)},
			"primary_ips":  schema.SetAttribute{Required: true, ElementType: types.StringType, Validators: ipSet, MarkdownDescription: "Primary nameserver IP addresses."},
			"allow_notify": schema.SetAttribute{Optional: true, Computed: true, Default: setdefault.StaticValue(emptyStringSet()), ElementType: types.StringType, Validators: ipSet},
			"transfer_key_id": schema.StringAttribute{
				Optional: true, Computed: true, Default: stringdefault.StaticString(""),
				Validators:          []validator.String{stringvalidator.Any(stringvalidator.OneOf(""), validators.IsUUIDv4())},
				MarkdownDescription: "Preferred shared BIND TSIG key UUID used to authenticate AXFR/IXFR and NOTIFY. The same key can be referenced by BIND view selectors. Mutually exclusive with the legacy inline transfer-key attributes.",
			},
			"transfer_key_algorithm": schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), Validators: []validator.String{stringvalidator.OneOf("", "hmac-sha512", "hmac-sha384", "hmac-sha256", "hmac-sha224", "hmac-sha1")}, MarkdownDescription: "Legacy inline TSIG algorithm. Prefer transfer_key_id for new configurations."},
			"transfer_key_name":      schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), MarkdownDescription: "Legacy inline TSIG name. Prefer transfer_key_id for new configurations."},
			"transfer_key": schema.StringAttribute{
				Optional: true, Sensitive: true, WriteOnly: true,
				MarkdownDescription: "Legacy write-only Base64 transfer TSIG secret. Required only for the inline algorithm/name path. Prefer transfer_key_id for new configurations. Supply through an ephemeral variable and increment transfer_key_version whenever it changes.",
			},
			"transfer_key_version": schema.Int64Attribute{
				Optional: true, Computed: true, Default: int64default.StaticInt64(0),
				Validators:          []validator.Int64{int64validator.AtLeast(0)},
				MarkdownDescription: "Stateful rotation marker for the write-only transfer key. Increment whenever the secret changes.",
			},
			"transfer_key_configured": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the secondary zone has authenticated transfer TSIG configured, either through transfer_key_id or a legacy inline secret. Secret material is never returned to state.",
			},
			"allow_transfer_acl_ids": schema.SetAttribute{Optional: true, Computed: true, Default: setdefault.StaticValue(emptyStringSet()), ElementType: types.StringType, Validators: uuidSet},
			"allow_query_acl_ids":    schema.SetAttribute{Optional: true, Computed: true, Default: setdefault.StaticValue(emptyStringSet()), ElementType: types.StringType, Validators: uuidSet},
		},
	}
}

func secondaryDomainResourceSchemaV0() schema.Schema {
	uuidSet := []validator.Set{setvalidator.ValueStringsAre(validators.IsUUIDv4())}
	ipSet := []validator.Set{setvalidator.ValueStringsAre(validators.IPAddress())}
	return schema.Schema{MarkdownDescription: "Legacy BIND secondary-zone schema whose transfer key was stored as sensitive Terraform state.", Version: 0, Attributes: map[string]schema.Attribute{
		"id":                     schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"view_id":                schema.StringAttribute{Required: true, Validators: []validator.String{validators.IsUUIDv4()}},
		"domain_name":            schema.StringAttribute{Required: true, Validators: []validator.String{stringvalidator.LengthBetween(1, 255)}},
		"enabled":                schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true)},
		"primary_ips":            schema.SetAttribute{Required: true, ElementType: types.StringType, Validators: ipSet},
		"allow_notify":           schema.SetAttribute{Optional: true, Computed: true, Default: setdefault.StaticValue(emptyStringSet()), ElementType: types.StringType, Validators: ipSet},
		"transfer_key_algorithm": schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), Validators: []validator.String{stringvalidator.OneOf("", "hmac-sha512", "hmac-sha384", "hmac-sha256", "hmac-sha224", "hmac-sha1")}},
		"transfer_key_name":      schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("")},
		"transfer_key":           schema.StringAttribute{Optional: true, Computed: true, Sensitive: true, Default: stringdefault.StaticString("")},
		"allow_transfer_acl_ids": schema.SetAttribute{Optional: true, Computed: true, Default: setdefault.StaticValue(emptyStringSet()), ElementType: types.StringType, Validators: uuidSet},
		"allow_query_acl_ids":    schema.SetAttribute{Optional: true, Computed: true, Default: setdefault.StaticValue(emptyStringSet()), ElementType: types.StringType, Validators: uuidSet},
	}}
}

func secondaryDomainDataSourceSchema() dschema.Schema {
	return dschema.Schema{MarkdownDescription: "Reads a secondary BIND zone without exposing transfer TSIG secret material.", Attributes: map[string]dschema.Attribute{
		"id": dschema.StringAttribute{Required: true}, "view_id": dschema.StringAttribute{Computed: true}, "domain_name": dschema.StringAttribute{Computed: true}, "enabled": dschema.BoolAttribute{Computed: true},
		"primary_ips": dschema.SetAttribute{Computed: true, ElementType: types.StringType}, "allow_notify": dschema.SetAttribute{Computed: true, ElementType: types.StringType},
		"transfer_key_id":        dschema.StringAttribute{Computed: true, MarkdownDescription: "Shared BIND TSIG key UUID when the preferred reference-based transfer authentication is configured."},
		"transfer_key_algorithm": dschema.StringAttribute{Computed: true}, "transfer_key_name": dschema.StringAttribute{Computed: true},
		"transfer_key":            dschema.StringAttribute{Computed: true, Sensitive: true, MarkdownDescription: "Deprecated compatibility attribute. Always null; transfer TSIG secrets are never returned by the provider."},
		"transfer_key_configured": dschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether a non-empty transfer TSIG secret is configured."},
		"allow_transfer_acl_ids":  dschema.SetAttribute{Computed: true, ElementType: types.StringType}, "allow_query_acl_ids": dschema.SetAttribute{Computed: true, ElementType: types.StringType},
	}}
}

func secondaryDomainModelToAPI(d *secondaryDomainResourceModel, secret string) *apibind.SecondaryDomain {
	return &apibind.SecondaryDomain{View: api.SelectedMap(d.ViewID.ValueString()), DomainName: d.DomainName.ValueString(), Enabled: tools.BoolToString(d.Enabled.ValueBool()), PrimaryIP: api.SelectedMapList(tools.SetToStringSlice(d.PrimaryIPs)), AllowNotify: api.SelectedMapList(tools.SetToStringSlice(d.AllowNotify)), TransferKeyID: api.SelectedMap(d.TransferKeyID.ValueString()), TransferKeyAlgorithm: api.SelectedMap(d.TransferKeyAlgorithm.ValueString()), TransferKeyName: d.TransferKeyName.ValueString(), TransferKey: secret, AllowTransfer: api.SelectedMapList(tools.SetToStringSlice(d.AllowTransferACLs)), AllowQuery: api.SelectedMapList(tools.SetToStringSlice(d.AllowQueryACLs))}
}

func applySecondaryDomainModel(remote *apibind.SecondaryDomain, d *secondaryDomainResourceModel, secret types.String) {
	remote.View = api.SelectedMap(d.ViewID.ValueString())
	remote.DomainName = d.DomainName.ValueString()
	remote.Enabled = tools.BoolToString(d.Enabled.ValueBool())
	remote.PrimaryIP = api.SelectedMapList(tools.SetToStringSlice(d.PrimaryIPs))
	remote.AllowNotify = api.SelectedMapList(tools.SetToStringSlice(d.AllowNotify))
	remote.TransferKeyID = api.SelectedMap(d.TransferKeyID.ValueString())
	if d.TransferKeyID.ValueString() != "" {
		remote.TransferKeyAlgorithm = api.SelectedMap("")
		remote.TransferKeyName = ""
		remote.TransferKey = ""
	} else {
		remote.TransferKeyAlgorithm = api.SelectedMap(d.TransferKeyAlgorithm.ValueString())
		remote.TransferKeyName = d.TransferKeyName.ValueString()
		if d.TransferKeyAlgorithm.ValueString() == "" && d.TransferKeyName.ValueString() == "" {
			remote.TransferKey = ""
		} else if !secret.IsNull() && !secret.IsUnknown() {
			remote.TransferKey = secret.ValueString()
		}
	}
	remote.AllowTransfer = api.SelectedMapList(tools.SetToStringSlice(d.AllowTransferACLs))
	remote.AllowQuery = api.SelectedMapList(tools.SetToStringSlice(d.AllowQueryACLs))
}

func secondaryDomainAPIToModel(d *apibind.SecondaryDomain) *secondaryDomainResourceModel {
	return &secondaryDomainResourceModel{
		ViewID: types.StringValue(d.View.String()), DomainName: types.StringValue(d.DomainName), Enabled: types.BoolValue(tools.StringToBool(d.Enabled)), PrimaryIPs: tools.StringSliceToSet([]string(d.PrimaryIP)), AllowNotify: tools.StringSliceToSet([]string(d.AllowNotify)), TransferKeyID: types.StringValue(d.TransferKeyID.String()), TransferKeyAlgorithm: types.StringValue(d.TransferKeyAlgorithm.String()), TransferKeyName: types.StringValue(d.TransferKeyName), TransferKey: types.StringNull(), TransferKeyVersion: types.Int64Null(), TransferKeyConfigured: types.BoolValue(d.TransferKeyID.String() != "" || d.TransferKey != ""), AllowTransferACLs: tools.StringSliceToSet([]string(d.AllowTransfer)), AllowQueryACLs: tools.StringSliceToSet([]string(d.AllowQuery)),
	}
}

func secondaryDomainAPIToDataSourceModel(d *apibind.SecondaryDomain) (*secondaryDomainDataSourceModel, error) {
	return &secondaryDomainDataSourceModel{
		ViewID: types.StringValue(d.View.String()), DomainName: types.StringValue(d.DomainName), Enabled: types.BoolValue(tools.StringToBool(d.Enabled)), PrimaryIPs: tools.StringSliceToSet([]string(d.PrimaryIP)), AllowNotify: tools.StringSliceToSet([]string(d.AllowNotify)), TransferKeyID: types.StringValue(d.TransferKeyID.String()), TransferKeyAlgorithm: types.StringValue(d.TransferKeyAlgorithm.String()), TransferKeyName: types.StringValue(d.TransferKeyName), TransferKey: types.StringNull(), TransferKeyConfigured: types.BoolValue(d.TransferKeyID.String() != "" || d.TransferKey != ""), AllowTransferACLs: tools.StringSliceToSet([]string(d.AllowTransfer)), AllowQueryACLs: tools.StringSliceToSet([]string(d.AllowQuery)),
	}, nil
}
