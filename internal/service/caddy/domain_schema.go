package caddy

import (
	"fmt"
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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type domainResourceModel struct {
	ID                              types.String `tfsdk:"id"`
	Enabled                         types.Bool   `tfsdk:"enabled"`
	Domain                          types.String `tfsdk:"domain"`
	Port                            types.Int64  `tfsdk:"port"`
	Protocol                        types.String `tfsdk:"protocol"`
	CertificateMode                 types.String `tfsdk:"certificate_mode"`
	CertificateRefID                types.String `tfsdk:"certificate_ref_id"`
	InternalCAName                  types.String `tfsdk:"internal_ca_name"`
	InternalCertificateLifetimeDays types.Int64  `tfsdk:"internal_certificate_lifetime_days"`
	InternalCertificateKeyType      types.String `tfsdk:"internal_certificate_key_type"`
	InternalCertificateDigest       types.String `tfsdk:"internal_certificate_digest"`
	GeneratedCertificateID          types.String `tfsdk:"generated_certificate_id"`
	AccessListID                    types.String `tfsdk:"access_list_id"`
	BasicAuthIDs                    types.Set    `tfsdk:"basic_auth_ids"`
	Description                     types.String `tfsdk:"description"`
	DNSChallenge                    types.Bool   `tfsdk:"dns_challenge"`
	DNSChallengeOverrideDomain      types.String `tfsdk:"dns_challenge_override_domain"`
	AccessLog                       types.Bool   `tfsdk:"access_log"`
	DynamicDNS                      types.Bool   `tfsdk:"dynamic_dns"`
	ACMEPassthrough                 types.Bool   `tfsdk:"acme_passthrough"`
	ClientAuthMode                  types.String `tfsdk:"client_auth_mode"`
	ClientAuthCARefIDs              types.Set    `tfsdk:"client_auth_ca_ref_ids"`
}

func domainResourceSchema() schema.Schema {
	return schema.Schema{MarkdownDescription: "Manages a Caddy frontend domain. HTTPS can use automatic public ACME, a dynamically issued certificate from an existing OPNsense CA, or an existing custom certificate.", Version: 1, Attributes: map[string]schema.Attribute{
		"id":                                 schema.StringAttribute{Computed: true, MarkdownDescription: "UUID of the Caddy domain.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"enabled":                            schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true), MarkdownDescription: "Whether to enable this domain. Defaults to `true`."},
		"domain":                             schema.StringAttribute{Required: true, MarkdownDescription: "Frontend FQDN or wildcard domain handled by Caddy."},
		"port":                               schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(-1), MarkdownDescription: "Optional frontend port. `-1` uses the global HTTP or HTTPS port.", Validators: []validator.Int64{int64validator.Any(int64validator.OneOf(-1), int64validator.Between(1, 65535))}},
		"protocol":                           schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("https"), MarkdownDescription: "Frontend protocol: `https` or `http`. Defaults to `https`.", Validators: []validator.String{stringvalidator.OneOf("https", "http")}},
		"certificate_mode":                   schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("acme"), MarkdownDescription: "Certificate mode: `acme` for automatic public ACME, `internal` for a dynamically issued certificate from an existing OPNsense CA, `custom` for an existing certificate reference, or `none` for HTTP. Defaults to `acme`.", Validators: []validator.String{stringvalidator.OneOf("acme", "internal", "custom", "none")}},
		"certificate_ref_id":                 schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "OPNsense certificate reference ID. Required for `custom`; computed for `internal`; empty for `acme` and `none`."},
		"internal_ca_name":                   schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), MarkdownDescription: "Exact existing OPNsense CA name, common name, description, or reference ID used in `internal` mode. Terraform never creates or exports the CA private key."},
		"internal_certificate_lifetime_days": schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(3650), MarkdownDescription: "Validity of dynamically issued internal certificates in days. Defaults to `3650`.", Validators: []validator.Int64{int64validator.AtLeast(1)}},
		"internal_certificate_key_type":      schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("4096"), MarkdownDescription: "Key type for dynamically issued internal certificates. Defaults to `4096`.", Validators: []validator.String{stringvalidator.OneOf("2048", "3072", "4096", "8192", "prime256v1", "secp384r1", "secp521r1")}},
		"internal_certificate_digest":        schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("sha256"), MarkdownDescription: "Digest used for dynamically issued internal certificates. Defaults to `sha256`.", Validators: []validator.String{stringvalidator.OneOf("sha224", "sha256", "sha384", "sha512")}},
		"generated_certificate_id":           schema.StringAttribute{Computed: true, MarkdownDescription: "UUID of the certificate created and owned by this resource in `internal` mode."},
		"access_list_id":                     schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), MarkdownDescription: "Optional Caddy access-list UUID. Defaults to `\"\"`."},
		"basic_auth_ids":                     schema.SetAttribute{Optional: true, Computed: true, ElementType: types.StringType, Default: setdefault.StaticValue(tools.EmptySetValue(types.StringType)), MarkdownDescription: "Caddy basic-auth entry UUIDs. Defaults to an empty set."},
		"description":                        schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), MarkdownDescription: "Optional description. Defaults to `\"\"`."},
		"dns_challenge":                      schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), MarkdownDescription: "Whether automatic ACME uses DNS-01. Defaults to `false`."},
		"dns_challenge_override_domain":      schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), MarkdownDescription: "Optional delegated DNS-01 challenge domain. Defaults to `\"\"`."},
		"access_log":                         schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), MarkdownDescription: "Whether to enable HTTP access logging for this domain. Defaults to `false`."},
		"dynamic_dns":                        schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), MarkdownDescription: "Whether Caddy updates DNS for this domain. Defaults to `false`."},
		"acme_passthrough":                   schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), MarkdownDescription: "Whether to pass ACME HTTP-01 challenges to the upstream. Defaults to `false`."},
		"client_auth_mode":                   schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), MarkdownDescription: "Optional mTLS client-auth mode: empty for `require_and_verify`, or `request`, `require`, or `verify_if_given`.", Validators: []validator.String{stringvalidator.OneOf("", "request", "require", "verify_if_given")}},
		"client_auth_ca_ref_ids":             schema.SetAttribute{Optional: true, Computed: true, ElementType: types.StringType, Default: setdefault.StaticValue(tools.EmptySetValue(types.StringType)), MarkdownDescription: "OPNsense CA reference IDs trusted for client certificate authentication. Defaults to an empty set."},
	}}
}

func domainDataSourceSchema() dschema.Schema {
	return dschema.Schema{MarkdownDescription: "Reads a Caddy domain by UUID. Dynamically generated certificate ownership cannot be inferred for imported or data-source-only domains.", Attributes: map[string]dschema.Attribute{
		"id": dschema.StringAttribute{Required: true}, "enabled": dschema.BoolAttribute{Computed: true}, "domain": dschema.StringAttribute{Computed: true}, "port": dschema.Int64Attribute{Computed: true}, "protocol": dschema.StringAttribute{Computed: true}, "certificate_mode": dschema.StringAttribute{Computed: true}, "certificate_ref_id": dschema.StringAttribute{Computed: true}, "internal_ca_name": dschema.StringAttribute{Computed: true}, "internal_certificate_lifetime_days": dschema.Int64Attribute{Computed: true}, "internal_certificate_key_type": dschema.StringAttribute{Computed: true}, "internal_certificate_digest": dschema.StringAttribute{Computed: true}, "generated_certificate_id": dschema.StringAttribute{Computed: true}, "access_list_id": dschema.StringAttribute{Computed: true}, "basic_auth_ids": dschema.SetAttribute{Computed: true, ElementType: types.StringType}, "description": dschema.StringAttribute{Computed: true}, "dns_challenge": dschema.BoolAttribute{Computed: true}, "dns_challenge_override_domain": dschema.StringAttribute{Computed: true}, "access_log": dschema.BoolAttribute{Computed: true}, "dynamic_dns": dschema.BoolAttribute{Computed: true}, "acme_passthrough": dschema.BoolAttribute{Computed: true}, "client_auth_mode": dschema.StringAttribute{Computed: true}, "client_auth_ca_ref_ids": dschema.SetAttribute{Computed: true, ElementType: types.StringType},
	}}
}

func validateDomainModel(d *domainResourceModel) error {
	protocol, mode := d.Protocol.ValueString(), d.CertificateMode.ValueString()
	if protocol == "http" && mode != "none" {
		return fmt.Errorf("certificate_mode must be `none` when protocol is `http`")
	}
	if protocol == "https" && mode == "none" {
		return fmt.Errorf("certificate_mode `none` is only valid with protocol `http`")
	}
	if mode == "internal" && d.InternalCAName.ValueString() == "" {
		return fmt.Errorf("internal_ca_name is required when certificate_mode is `internal`")
	}
	if mode == "custom" && d.CertificateRefID.ValueString() == "" {
		return fmt.Errorf("certificate_ref_id is required when certificate_mode is `custom`")
	}
	return nil
}
func buildDomainAPI(d *domainResourceModel, certificateRef string) (*apicaddy.Domain, error) {
	if err := validateDomainModel(d); err != nil {
		return nil, err
	}
	disable := "0"
	if d.Protocol.ValueString() == "http" {
		disable = "1"
	}
	return &apicaddy.Domain{Enabled: tools.BoolToString(d.Enabled.ValueBool()), Domain: d.Domain.ValueString(), Port: intToAPI(d.Port), AccessList: api.SelectedMap(d.AccessListID.ValueString()), BasicAuth: api.SelectedMapList(tools.SetToStringSlice(d.BasicAuthIDs)), Description: d.Description.ValueString(), DNSChallenge: tools.BoolToString(d.DNSChallenge.ValueBool()), DNSChallengeOverrideDomain: d.DNSChallengeOverrideDomain.ValueString(), CustomCertificate: api.SelectedMap(certificateRef), AccessLog: tools.BoolToString(d.AccessLog.ValueBool()), DynamicDNS: tools.BoolToString(d.DynamicDNS.ValueBool()), ACMEPassthrough: tools.BoolToString(d.ACMEPassthrough.ValueBool()), DisableTLS: api.SelectedMap(disable), ClientAuthMode: api.SelectedMap(d.ClientAuthMode.ValueString()), ClientAuthTrustPool: api.SelectedMapList(tools.SetToStringSlice(d.ClientAuthCARefIDs))}, nil
}
func domainStructToSchema(d *apicaddy.Domain, prior *domainResourceModel) (*domainResourceModel, error) {
	protocol := "https"
	mode := "acme"
	if d.DisableTLS.String() == "1" {
		protocol = "http"
		mode = "none"
	} else if d.CustomCertificate.String() != "" {
		mode = "custom"
	}
	m := &domainResourceModel{Enabled: types.BoolValue(tools.StringToBool(d.Enabled)), Domain: types.StringValue(d.Domain), Port: apiInt(d.Port), Protocol: types.StringValue(protocol), CertificateMode: types.StringValue(mode), CertificateRefID: types.StringValue(d.CustomCertificate.String()), InternalCAName: types.StringValue(""), InternalCertificateLifetimeDays: types.Int64Value(3650), InternalCertificateKeyType: types.StringValue("4096"), InternalCertificateDigest: types.StringValue("sha256"), GeneratedCertificateID: types.StringNull(), AccessListID: types.StringValue(d.AccessList.String()), BasicAuthIDs: tools.StringSliceToSet([]string(d.BasicAuth)), Description: types.StringValue(d.Description), DNSChallenge: types.BoolValue(tools.StringToBool(d.DNSChallenge)), DNSChallengeOverrideDomain: types.StringValue(d.DNSChallengeOverrideDomain), AccessLog: types.BoolValue(tools.StringToBool(d.AccessLog)), DynamicDNS: types.BoolValue(tools.StringToBool(d.DynamicDNS)), ACMEPassthrough: types.BoolValue(tools.StringToBool(d.ACMEPassthrough)), ClientAuthMode: types.StringValue(d.ClientAuthMode.String()), ClientAuthCARefIDs: tools.StringSliceToSet([]string(d.ClientAuthTrustPool))}
	if prior != nil && prior.CertificateMode.ValueString() == "internal" {
		m.CertificateMode = types.StringValue("internal")
		m.InternalCAName = prior.InternalCAName
		m.InternalCertificateLifetimeDays = prior.InternalCertificateLifetimeDays
		m.InternalCertificateKeyType = prior.InternalCertificateKeyType
		m.InternalCertificateDigest = prior.InternalCertificateDigest
		m.GeneratedCertificateID = prior.GeneratedCertificateID
		m.CertificateRefID = prior.CertificateRefID
	}
	return m, nil
}
