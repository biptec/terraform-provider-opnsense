package bind

import "github.com/hashicorp/terraform-plugin-framework/types"

type primaryDomainDataSourceModel struct {
	ID                types.String `tfsdk:"id"`
	ViewID            types.String `tfsdk:"view_id"`
	ViewName          types.String `tfsdk:"view_name"`
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

func primaryDomainDataSourceModelFromResource(m *primaryDomainResourceModel, id, viewName string) *primaryDomainDataSourceModel {
	return &primaryDomainDataSourceModel{
		ID: types.StringValue(id), ViewID: m.ViewID, ViewName: types.StringValue(viewName),
		DomainName: types.StringValue(normalizeBindDomainName(m.DomainName.ValueString())), Enabled: m.Enabled,
		AllowTransferACLs: m.AllowTransferACLs, AllowRndcTransfer: m.AllowRndcTransfer,
		TransferKeyID: m.TransferKeyID, AlsoNotify: m.AlsoNotify, AllowQueryACLs: m.AllowQueryACLs,
		AllowRndcUpdate: m.AllowRndcUpdate, UpdateKeyIDs: m.UpdateKeyIDs, UpdatePolicy: m.UpdatePolicy,
		DNSSEC: m.DNSSEC, Serial: m.Serial, TTL: m.TTL, Refresh: m.Refresh, Retry: m.Retry,
		Expire: m.Expire, NegativeTTL: m.NegativeTTL, MailAdmin: m.MailAdmin, DNSServer: m.DNSServer,
	}
}
