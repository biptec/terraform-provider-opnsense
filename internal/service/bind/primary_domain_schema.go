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

type primaryDomainResourceModel struct {
	ID                types.String `tfsdk:"id"`
	ViewID            types.String `tfsdk:"view_id"`
	DomainName        types.String `tfsdk:"domain_name"`
	Enabled           types.Bool   `tfsdk:"enabled"`
	AllowTransferACLs types.Set    `tfsdk:"allow_transfer_acl_ids"`
	AllowRndcTransfer types.Bool   `tfsdk:"allow_rndc_transfer"`
	TransferKeyID     types.String `tfsdk:"transfer_key_id"`
	AlsoNotify        types.Set    `tfsdk:"also_notify"`
	AllowQueryACLs    types.Set    `tfsdk:"allow_query_acl_ids"`
	AllowRndcUpdate   types.Bool   `tfsdk:"allow_rndc_update"`
	UpdateKeyIDs      types.Set    `tfsdk:"update_key_ids"`
	UpdatePolicy      types.String `tfsdk:"update_policy"`
	DNSSEC            types.Bool   `tfsdk:"dnssec"`
	Serial            types.String `tfsdk:"serial"`
	TTL               types.Int64  `tfsdk:"ttl"`
	Refresh           types.Int64  `tfsdk:"refresh"`
	Retry             types.Int64  `tfsdk:"retry"`
	Expire            types.Int64  `tfsdk:"expire"`
	NegativeTTL       types.Int64  `tfsdk:"negative_ttl"`
	MailAdmin         types.String `tfsdk:"mail_admin"`
	DNSServer         types.String `tfsdk:"dns_server"`
}

func primaryDomainResourceSchema() schema.Schema {
	uuidSet := []validator.Set{setvalidator.ValueStringsAre(validators.IsUUIDv4())}
	return schema.Schema{MarkdownDescription: "Manages a primary authoritative BIND zone inside a selected view.", Attributes: map[string]schema.Attribute{
		"id":                     schema.StringAttribute{Computed: true, MarkdownDescription: "UUID of the zone.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"view_id":                schema.StringAttribute{Required: true, MarkdownDescription: "UUID of the BIND view containing this zone.", Validators: []validator.String{validators.IsUUIDv4()}},
		"domain_name":            schema.StringAttribute{Required: true, MarkdownDescription: "Authoritative zone name.", Validators: []validator.String{stringvalidator.LengthBetween(1, 255)}},
		"enabled":                schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true)},
		"allow_transfer_acl_ids": schema.SetAttribute{Optional: true, Computed: true, Default: setdefault.StaticValue(emptyStringSet()), ElementType: types.StringType, Validators: uuidSet, MarkdownDescription: "ACL UUIDs allowed to transfer this zone."},
		"allow_rndc_transfer":    schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), MarkdownDescription: "Allow transfer using the built-in rndc key."},
		"transfer_key_id": schema.StringAttribute{
			Optional: true, Computed: true, Default: stringdefault.StaticString(""),
			Validators:          []validator.String{stringvalidator.Any(stringvalidator.OneOf(""), validators.IsUUIDv4())},
			MarkdownDescription: "Optional TSIG key UUID used for authenticated AXFR/IXFR and also-notify. The key secret remains owned by opnsense_bind_tsig_key.",
		},
		"also_notify": schema.SetAttribute{
			Optional: true, Computed: true, Default: setdefault.StaticValue(emptyStringSet()), ElementType: types.StringType,
			Validators:          []validator.Set{setvalidator.ValueStringsAre(validators.IPAddress())},
			MarkdownDescription: "Secondary nameserver IP addresses that receive NOTIFY. A non-empty set requires transfer_key_id so notifications are authenticated.",
		},
		"allow_query_acl_ids": schema.SetAttribute{Optional: true, Computed: true, Default: setdefault.StaticValue(emptyStringSet()), ElementType: types.StringType, Validators: uuidSet, MarkdownDescription: "Optional ACL UUIDs allowed to query this zone. View policy applies when empty."},
		"allow_rndc_update":   schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), MarkdownDescription: "Allow unrestricted dynamic updates using the built-in rndc key."},
		"update_key_ids":      schema.SetAttribute{Optional: true, Computed: true, ElementType: types.StringType, Validators: uuidSet, MarkdownDescription: "TSIG key UUIDs allowed to perform RFC2136 updates. Omit this attribute when memberships are owned additively by opnsense_bind_primary_domain_update_key resources; configure it explicitly only when this resource owns the complete set."},
		"update_policy":       schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("self_txt"), Validators: []validator.String{stringvalidator.OneOf("self_txt", "zonesub_txt", "zonesub_any")}, MarkdownDescription: "RFC2136 update policy. `self_txt` allows each TSIG key to update only the TXT owner matching the key name and is recommended for ACME DNS-01. The `zonesub_*` policies grant broader zone-wide access."},
		"dnssec":              schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), MarkdownDescription: "Enable automatic DNSSEC signing with the BIND default policy."},
		"serial":              schema.StringAttribute{Computed: true, MarkdownDescription: "Current SOA serial generated by OPNsense."},
		"ttl":                 schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(300), Validators: []validator.Int64{int64validator.AtLeast(1)}, MarkdownDescription: "Default zone TTL in seconds."},
		"refresh":             schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(300), Validators: []validator.Int64{int64validator.AtLeast(1)}},
		"retry":               schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(300), Validators: []validator.Int64{int64validator.AtLeast(1)}},
		"expire":              schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(86400), Validators: []validator.Int64{int64validator.AtLeast(1)}},
		"negative_ttl":        schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(300), Validators: []validator.Int64{int64validator.AtLeast(1)}},
		"mail_admin":          schema.StringAttribute{Required: true, MarkdownDescription: "SOA responsible-party email address."},
		"dns_server":          schema.StringAttribute{Required: true, MarkdownDescription: "Primary authoritative nameserver stored in the SOA record."},
	}}
}

func primaryDomainDataSourceSchema() dschema.Schema {
	return dschema.Schema{MarkdownDescription: "Reads a primary BIND zone.", Attributes: map[string]dschema.Attribute{
		"id":                     dschema.StringAttribute{Required: true},
		"view_id":                dschema.StringAttribute{Computed: true},
		"domain_name":            dschema.StringAttribute{Computed: true},
		"enabled":                dschema.BoolAttribute{Computed: true},
		"allow_transfer_acl_ids": dschema.SetAttribute{Computed: true, ElementType: types.StringType},
		"allow_rndc_transfer":    dschema.BoolAttribute{Computed: true},
		"transfer_key_id":        dschema.StringAttribute{Computed: true},
		"also_notify":            dschema.SetAttribute{Computed: true, ElementType: types.StringType},
		"allow_query_acl_ids":    dschema.SetAttribute{Computed: true, ElementType: types.StringType},
		"allow_rndc_update":      dschema.BoolAttribute{Computed: true},
		"update_key_ids":         dschema.SetAttribute{Computed: true, ElementType: types.StringType},
		"update_policy":          dschema.StringAttribute{Computed: true},
		"dnssec":                 dschema.BoolAttribute{Computed: true},
		"serial":                 dschema.StringAttribute{Computed: true},
		"ttl":                    dschema.Int64Attribute{Computed: true},
		"refresh":                dschema.Int64Attribute{Computed: true},
		"retry":                  dschema.Int64Attribute{Computed: true},
		"expire":                 dschema.Int64Attribute{Computed: true},
		"negative_ttl":           dschema.Int64Attribute{Computed: true},
		"mail_admin":             dschema.StringAttribute{Computed: true},
		"dns_server":             dschema.StringAttribute{Computed: true},
	}}
}

func primaryDomainModelToAPI(d *primaryDomainResourceModel) (*apibind.PrimaryDomain, error) {
	serial := ""
	if !d.Serial.IsNull() && !d.Serial.IsUnknown() {
		serial = d.Serial.ValueString()
	}
	return &apibind.PrimaryDomain{
		View: api.SelectedMap(d.ViewID.ValueString()), DomainName: d.DomainName.ValueString(), Enabled: tools.BoolToString(d.Enabled.ValueBool()),
		AllowTransfer: api.SelectedMapList(tools.SetToStringSlice(d.AllowTransferACLs)), AllowRndcTransfer: tools.BoolToString(d.AllowRndcTransfer.ValueBool()),
		TransferKey: api.SelectedMap(d.TransferKeyID.ValueString()), AlsoNotify: api.SelectedMapList(tools.SetToStringSlice(d.AlsoNotify)),
		AllowQuery: api.SelectedMapList(tools.SetToStringSlice(d.AllowQueryACLs)), AllowRndcUpdate: tools.BoolToString(d.AllowRndcUpdate.ValueBool()),
		UpdateKeys: api.SelectedMapList(tools.SetToStringSlice(d.UpdateKeyIDs)), UpdatePolicy: api.SelectedMap(d.UpdatePolicy.ValueString()),
		DNSSEC: tools.BoolToString(d.DNSSEC.ValueBool()), Serial: serial, TimeToLive: tools.Int64ToString(d.TTL.ValueInt64()),
		Refresh: tools.Int64ToString(d.Refresh.ValueInt64()), Retry: tools.Int64ToString(d.Retry.ValueInt64()), Expire: tools.Int64ToString(d.Expire.ValueInt64()),
		Negative: tools.Int64ToString(d.NegativeTTL.ValueInt64()), MailAdmin: d.MailAdmin.ValueString(), DnsServer: d.DNSServer.ValueString(),
	}, nil
}

func primaryDomainAPIToModel(d *apibind.PrimaryDomain) (*primaryDomainResourceModel, error) {
	return &primaryDomainResourceModel{
		ViewID: types.StringValue(d.View.String()), DomainName: types.StringValue(d.DomainName), Enabled: types.BoolValue(tools.StringToBool(d.Enabled)),
		AllowTransferACLs: tools.StringSliceToSet([]string(d.AllowTransfer)), AllowRndcTransfer: types.BoolValue(tools.StringToBool(d.AllowRndcTransfer)),
		TransferKeyID: types.StringValue(d.TransferKey.String()), AlsoNotify: tools.StringSliceToSet([]string(d.AlsoNotify)),
		AllowQueryACLs: tools.StringSliceToSet([]string(d.AllowQuery)), AllowRndcUpdate: types.BoolValue(tools.StringToBool(d.AllowRndcUpdate)),
		UpdateKeyIDs: tools.StringSliceToSet([]string(d.UpdateKeys)), UpdatePolicy: types.StringValue(d.UpdatePolicy.String()), DNSSEC: types.BoolValue(tools.StringToBool(d.DNSSEC)),
		Serial: types.StringValue(d.Serial), TTL: types.Int64Value(tools.StringToInt64(d.TimeToLive)), Refresh: types.Int64Value(tools.StringToInt64(d.Refresh)),
		Retry: types.Int64Value(tools.StringToInt64(d.Retry)), Expire: types.Int64Value(tools.StringToInt64(d.Expire)), NegativeTTL: types.Int64Value(tools.StringToInt64(d.Negative)),
		MailAdmin: types.StringValue(d.MailAdmin), DNSServer: types.StringValue(d.DnsServer),
	}, nil
}
