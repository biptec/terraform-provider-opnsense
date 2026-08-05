package bind

import (
	"github.com/biptec/opnsense-go/pkg/api"
	apibind "github.com/biptec/opnsense-go/pkg/bind"
	"github.com/biptec/terraform-provider-opnsense/internal/tools"
	"github.com/biptec/terraform-provider-opnsense/internal/validators"
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

type secondaryDomainResourceModel struct {
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

func secondaryDomainResourceSchema() schema.Schema {
	uuidSet := []validator.Set{setvalidator.ValueStringsAre(validators.IsUUIDv4())}
	ipSet := []validator.Set{setvalidator.ValueStringsAre(validators.IPAddress())}
	return schema.Schema{MarkdownDescription: "Manages a secondary BIND zone inside a selected view.", Attributes: map[string]schema.Attribute{
		"id":                     schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"view_id":                schema.StringAttribute{Required: true, Validators: []validator.String{validators.IsUUIDv4()}},
		"domain_name":            schema.StringAttribute{Required: true, Validators: []validator.String{stringvalidator.LengthBetween(1, 255)}},
		"enabled":                schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true)},
		"primary_ips":            schema.SetAttribute{Required: true, ElementType: types.StringType, Validators: ipSet, MarkdownDescription: "Primary nameserver IP addresses."},
		"allow_notify":           schema.SetAttribute{Optional: true, Computed: true, Default: setdefault.StaticValue(emptyStringSet()), ElementType: types.StringType, Validators: ipSet},
		"transfer_key_algorithm": schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), Validators: []validator.String{stringvalidator.OneOf("", "hmac-sha512", "hmac-sha384", "hmac-sha256", "hmac-sha224", "hmac-sha1")}},
		"transfer_key_name":      schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("")},
		"transfer_key":           schema.StringAttribute{Optional: true, Computed: true, Sensitive: true, Default: stringdefault.StaticString(""), MarkdownDescription: "Base64 transfer TSIG secret."},
		"allow_transfer_acl_ids": schema.SetAttribute{Optional: true, Computed: true, Default: setdefault.StaticValue(emptyStringSet()), ElementType: types.StringType, Validators: uuidSet},
		"allow_query_acl_ids":    schema.SetAttribute{Optional: true, Computed: true, Default: setdefault.StaticValue(emptyStringSet()), ElementType: types.StringType, Validators: uuidSet},
	}}
}
func secondaryDomainDataSourceSchema() dschema.Schema {
	return dschema.Schema{MarkdownDescription: "Reads a secondary BIND zone.", Attributes: map[string]dschema.Attribute{
		"id": dschema.StringAttribute{Required: true}, "view_id": dschema.StringAttribute{Computed: true}, "domain_name": dschema.StringAttribute{Computed: true}, "enabled": dschema.BoolAttribute{Computed: true},
		"primary_ips": dschema.SetAttribute{Computed: true, ElementType: types.StringType}, "allow_notify": dschema.SetAttribute{Computed: true, ElementType: types.StringType},
		"transfer_key_algorithm": dschema.StringAttribute{Computed: true}, "transfer_key_name": dschema.StringAttribute{Computed: true}, "transfer_key": dschema.StringAttribute{Computed: true, Sensitive: true},
		"allow_transfer_acl_ids": dschema.SetAttribute{Computed: true, ElementType: types.StringType}, "allow_query_acl_ids": dschema.SetAttribute{Computed: true, ElementType: types.StringType},
	}}
}
func secondaryDomainModelToAPI(d *secondaryDomainResourceModel) (*apibind.SecondaryDomain, error) {
	return &apibind.SecondaryDomain{View: api.SelectedMap(d.ViewID.ValueString()), DomainName: d.DomainName.ValueString(), Enabled: tools.BoolToString(d.Enabled.ValueBool()), PrimaryIP: api.SelectedMapList(tools.SetToStringSlice(d.PrimaryIPs)), AllowNotify: api.SelectedMapList(tools.SetToStringSlice(d.AllowNotify)), TransferKeyAlgorithm: api.SelectedMap(d.TransferKeyAlgorithm.ValueString()), TransferKeyName: d.TransferKeyName.ValueString(), TransferKey: d.TransferKey.ValueString(), AllowTransfer: api.SelectedMapList(tools.SetToStringSlice(d.AllowTransferACLs)), AllowQuery: api.SelectedMapList(tools.SetToStringSlice(d.AllowQueryACLs))}, nil
}
func secondaryDomainAPIToModel(d *apibind.SecondaryDomain) (*secondaryDomainResourceModel, error) {
	return &secondaryDomainResourceModel{ViewID: types.StringValue(d.View.String()), DomainName: types.StringValue(d.DomainName), Enabled: types.BoolValue(tools.StringToBool(d.Enabled)), PrimaryIPs: tools.StringSliceToSet([]string(d.PrimaryIP)), AllowNotify: tools.StringSliceToSet([]string(d.AllowNotify)), TransferKeyAlgorithm: types.StringValue(d.TransferKeyAlgorithm.String()), TransferKeyName: types.StringValue(d.TransferKeyName), TransferKey: types.StringValue(d.TransferKey), AllowTransferACLs: tools.StringSliceToSet([]string(d.AllowTransfer)), AllowQueryACLs: tools.StringSliceToSet([]string(d.AllowQuery))}, nil
}
