package acmeclient

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type settingsResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	AutoRenewal types.Bool   `tfsdk:"auto_renewal"`
	Environment types.String `tfsdk:"environment"`
	LogLevel    types.String `tfsdk:"log_level"`
	ShowIntro   types.Bool   `tfsdk:"show_intro"`
}

func settingsResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manages native OPNsense os-acme-client global settings. This singleton must be imported with ID `acmeclient_settings` before use.",
		Attributes: map[string]schema.Attribute{
			"id":           schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"enabled":      schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true)},
			"auto_renewal": schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true)},
			"environment":  schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("prod"), Validators: []validator.String{stringvalidator.OneOf("prod", "stg")}},
			"log_level":    schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("normal"), Validators: []validator.String{stringvalidator.OneOf("normal", "extended", "debug", "debug2", "debug3")}},
			"show_intro":   schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false)},
		},
	}
}

type accountResourceModel struct {
	ID                  types.String `tfsdk:"id"`
	Enabled             types.Bool   `tfsdk:"enabled"`
	Name                types.String `tfsdk:"name"`
	Description         types.String `tfsdk:"description"`
	Email               types.String `tfsdk:"email"`
	CA                  types.String `tfsdk:"ca"`
	Register            types.Bool   `tfsdk:"register"`
	RegistrationVersion types.Int64  `tfsdk:"registration_version"`
	Registered          types.Bool   `tfsdk:"registered"`
	StatusCode          types.String `tfsdk:"status_code"`
	StatusLastUpdate    types.String `tfsdk:"status_last_update"`
}

func accountResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manages an os-acme-client account while keeping its generated private account key on OPNsense. Registration is an explicit lifecycle action controlled by `register` and `registration_version`.",
		Attributes: map[string]schema.Attribute{
			"id":                   schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"enabled":              schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true)},
			"name":                 schema.StringAttribute{Required: true, Validators: []validator.String{stringvalidator.LengthBetween(1, 255)}},
			"description":          schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("")},
			"email":                schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("")},
			"ca":                   schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("letsencrypt"), Validators: []validator.String{stringvalidator.OneOf("letsencrypt", "letsencrypt_test", "google", "google_test", "buypass", "buypass_test", "sslcom", "zerossl", "custom")}},
			"register":             schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false)},
			"registration_version": schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(0), Validators: []validator.Int64{int64validator.AtLeast(0)}},
			"registered":           schema.BoolAttribute{Computed: true},
			"status_code":          schema.StringAttribute{Computed: true},
			"status_last_update":   schema.StringAttribute{Computed: true},
		},
	}
}

type validationResourceModel struct {
	ID                       types.String `tfsdk:"id"`
	Enabled                  types.Bool   `tfsdk:"enabled"`
	Name                     types.String `tfsdk:"name"`
	Description              types.String `tfsdk:"description"`
	Method                   types.String `tfsdk:"method"`
	DNSService               types.String `tfsdk:"dns_service"`
	DNSNsupdateServer        types.String `tfsdk:"dns_nsupdate_server"`
	DNSNsupdateZone          types.String `tfsdk:"dns_nsupdate_zone"`
	DNSNsupdateKey           types.String `tfsdk:"dns_nsupdate_key"`
	DNSNsupdateKeyVersion    types.Int64  `tfsdk:"dns_nsupdate_key_version"`
	DNSNsupdateKeyConfigured types.Bool   `tfsdk:"dns_nsupdate_key_configured"`
}

func validationResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manages an os-acme-client DNS-01 validation. `dns_nsupdate_key` is sensitive and write-only; pair it with an ephemeral input and bump `dns_nsupdate_key_version` on rotation.",
		Attributes: map[string]schema.Attribute{
			"id":                          schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"enabled":                     schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true)},
			"name":                        schema.StringAttribute{Required: true, Validators: []validator.String{stringvalidator.LengthBetween(1, 255)}},
			"description":                 schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("")},
			"method":                      schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("dns01"), Validators: []validator.String{stringvalidator.OneOf("dns01")}},
			"dns_service":                 schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("dns_nsupdate"), Validators: []validator.String{stringvalidator.OneOf("dns_nsupdate")}},
			"dns_nsupdate_server":         schema.StringAttribute{Required: true},
			"dns_nsupdate_zone":           schema.StringAttribute{Required: true},
			"dns_nsupdate_key":            schema.StringAttribute{Optional: true, Sensitive: true, WriteOnly: true, MarkdownDescription: "Write-only nsupdate key-file content. The value is never stored in plan or state."},
			"dns_nsupdate_key_version":    schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(0), Validators: []validator.Int64{int64validator.AtLeast(0)}},
			"dns_nsupdate_key_configured": schema.BoolAttribute{Computed: true},
		},
	}
}

type actionResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Type        types.String `tfsdk:"type"`
}

func actionResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manages an os-acme-client post-issuance automation action.",
		Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"enabled":     schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true)},
			"name":        schema.StringAttribute{Required: true},
			"description": schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("")},
			"type":        schema.StringAttribute{Required: true, Validators: []validator.String{stringvalidator.OneOf("configd_restart_gui")}},
		},
	}
}

type certificateResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Enabled          types.Bool   `tfsdk:"enabled"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	AltNames         types.Set    `tfsdk:"alt_names"`
	AccountID        types.String `tfsdk:"account_id"`
	ValidationID     types.String `tfsdk:"validation_id"`
	KeyLength        types.String `tfsdk:"key_length"`
	RestartActionIDs types.Set    `tfsdk:"restart_action_ids"`
	AutoRenewal      types.Bool   `tfsdk:"auto_renewal"`
	RenewInterval    types.Int64  `tfsdk:"renew_interval"`
	AliasMode        types.String `tfsdk:"alias_mode"`
	ChallengeAlias   types.String `tfsdk:"challenge_alias"`
	Issue            types.Bool   `tfsdk:"issue"`
	IssuanceVersion  types.Int64  `tfsdk:"issuance_version"`
	CertRefID        types.String `tfsdk:"cert_ref_id"`
	Issued           types.Bool   `tfsdk:"issued"`
	LastUpdate       types.String `tfsdk:"last_update"`
	StatusCode       types.String `tfsdk:"status_code"`
	StatusLastUpdate types.String `tfsdk:"status_last_update"`
}

func certificateResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manages an os-acme-client certificate. Certificate/account private keys stay endpoint-local. `issue` performs the explicit ACME issuance lifecycle action; bump `issuance_version` to retry or force renewal.",
		Attributes: map[string]schema.Attribute{
			"id":                 schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"enabled":            schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true)},
			"name":               schema.StringAttribute{Required: true},
			"description":        schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("")},
			"alt_names":          schema.SetAttribute{Optional: true, ElementType: types.StringType, PlanModifiers: []planmodifier.Set{setplanmodifier.UseStateForUnknown()}},
			"account_id":         schema.StringAttribute{Required: true},
			"validation_id":      schema.StringAttribute{Required: true},
			"key_length":         schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("key_ec256"), Validators: []validator.String{stringvalidator.OneOf("key_2048", "key_3072", "key_4096", "key_ec256", "key_ec384")}},
			"restart_action_ids": schema.SetAttribute{Optional: true, ElementType: types.StringType},
			"auto_renewal":       schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true)},
			"renew_interval":     schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(60), Validators: []validator.Int64{int64validator.Between(0, 5000)}},
			"alias_mode":         schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("challenge"), Validators: []validator.String{stringvalidator.OneOf("none", "automatic", "domain", "challenge")}},
			"challenge_alias":    schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("")},
			"issue":              schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false)},
			"issuance_version":   schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(0), Validators: []validator.Int64{int64validator.AtLeast(0)}},
			"cert_ref_id":        schema.StringAttribute{Computed: true},
			"issued":             schema.BoolAttribute{Computed: true},
			"last_update":        schema.StringAttribute{Computed: true},
			"status_code":        schema.StringAttribute{Computed: true},
			"status_last_update": schema.StringAttribute{Computed: true},
		},
	}
}
