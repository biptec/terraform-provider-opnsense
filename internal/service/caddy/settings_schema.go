package caddy

import (
	"github.com/biptec/opnsense-go/pkg/api"
	apicaddy "github.com/biptec/opnsense-go/pkg/caddy"
	"github.com/biptec/terraform-provider-opnsense/internal/tools"
	"github.com/biptec/terraform-provider-opnsense/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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
	"regexp"
	"sort"
)

type settingsResourceModel struct {
	ID                            types.String `tfsdk:"id"`
	Enabled                       types.Bool   `tfsdk:"enabled"`
	EnableLayer4                  types.Bool   `tfsdk:"enable_layer4"`
	HTTPPort                      types.Int64  `tfsdk:"http_port"`
	HTTPSPort                     types.Int64  `tfsdk:"https_port"`
	ListenAddresses               types.Set    `tfsdk:"listen_addresses"`
	ACMEEmail                     types.String `tfsdk:"acme_email"`
	AutoHTTPS                     types.String `tfsdk:"auto_https"`
	DNSProvider                   types.String `tfsdk:"dns_provider"`
	DNSAPIKey                     types.String `tfsdk:"dns_api_key"`
	DNSAPIKeyConfigured           types.Bool   `tfsdk:"dns_api_key_configured"`
	DNSRFC2136Server              types.String `tfsdk:"dns_rfc2136_server"`
	DNSRFC2136Port                types.Int64  `tfsdk:"dns_rfc2136_port"`
	DNSRFC2136KeyName             types.String `tfsdk:"dns_rfc2136_key_name"`
	DNSRFC2136KeyAlgorithm        types.String `tfsdk:"dns_rfc2136_key_algorithm"`
	DNSRFC2136Key                 types.String `tfsdk:"dns_rfc2136_key"`
	DNSRFC2136KeyConfigured       types.Bool   `tfsdk:"dns_rfc2136_key_configured"`
	DNSCredentialsVersion         types.Int64  `tfsdk:"dns_credentials_version"`
	DNSPropagationTimeoutDisabled types.Bool   `tfsdk:"dns_propagation_timeout_disabled"`
	DNSPropagationTimeout         types.Int64  `tfsdk:"dns_propagation_timeout"`
	DNSPropagationDelay           types.Int64  `tfsdk:"dns_propagation_delay"`
	DNSPropagationResolvers       types.String `tfsdk:"dns_propagation_resolvers"`
	RunAsUser                     types.String `tfsdk:"run_as_user"`
	GracePeriod                   types.Int64  `tfsdk:"grace_period"`
	HTTPVersions                  types.Set    `tfsdk:"http_versions"`
	LogLevel                      types.String `tfsdk:"log_level"`
	PlainAccessLog                types.Bool   `tfsdk:"plain_access_log"`
	PlainAccessLogKeep            types.Int64  `tfsdk:"plain_access_log_keep"`
}

type settingsDataSourceModel struct {
	ID                            types.String `tfsdk:"id"`
	Enabled                       types.Bool   `tfsdk:"enabled"`
	EnableLayer4                  types.Bool   `tfsdk:"enable_layer4"`
	HTTPPort                      types.Int64  `tfsdk:"http_port"`
	HTTPSPort                     types.Int64  `tfsdk:"https_port"`
	ListenAddresses               types.Set    `tfsdk:"listen_addresses"`
	ACMEEmail                     types.String `tfsdk:"acme_email"`
	AutoHTTPS                     types.String `tfsdk:"auto_https"`
	DNSProvider                   types.String `tfsdk:"dns_provider"`
	DNSAPIKeyConfigured           types.Bool   `tfsdk:"dns_api_key_configured"`
	DNSRFC2136Server              types.String `tfsdk:"dns_rfc2136_server"`
	DNSRFC2136Port                types.Int64  `tfsdk:"dns_rfc2136_port"`
	DNSRFC2136KeyName             types.String `tfsdk:"dns_rfc2136_key_name"`
	DNSRFC2136KeyAlgorithm        types.String `tfsdk:"dns_rfc2136_key_algorithm"`
	DNSRFC2136KeyConfigured       types.Bool   `tfsdk:"dns_rfc2136_key_configured"`
	DNSPropagationTimeoutDisabled types.Bool   `tfsdk:"dns_propagation_timeout_disabled"`
	DNSPropagationTimeout         types.Int64  `tfsdk:"dns_propagation_timeout"`
	DNSPropagationDelay           types.Int64  `tfsdk:"dns_propagation_delay"`
	DNSPropagationResolvers       types.String `tfsdk:"dns_propagation_resolvers"`
	RunAsUser                     types.String `tfsdk:"run_as_user"`
	GracePeriod                   types.Int64  `tfsdk:"grace_period"`
	HTTPVersions                  types.Set    `tfsdk:"http_versions"`
	LogLevel                      types.String `tfsdk:"log_level"`
	PlainAccessLog                types.Bool   `tfsdk:"plain_access_log"`
	PlainAccessLogKeep            types.Int64  `tfsdk:"plain_access_log_keep"`
}

func settingsResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manages Caddy global settings. This singleton must be imported with `caddy_settings` before use. Destroy removes it from Terraform state without changing OPNsense.", Version: 1,
		Attributes: map[string]schema.Attribute{
			"id":                               schema.StringAttribute{Computed: true, MarkdownDescription: "Always `caddy_settings`.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"enabled":                          schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), MarkdownDescription: "Whether to enable Caddy. Defaults to `false`."},
			"enable_layer4":                    schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), MarkdownDescription: "Whether to enable Layer 4 routing. Defaults to `false`."},
			"http_port":                        schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(80), MarkdownDescription: "Local HTTP listen port. Defaults to `80`.", Validators: []validator.Int64{int64validator.Between(1, 65535)}},
			"https_port":                       schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(443), MarkdownDescription: "Local HTTPS listen port. Defaults to `443`.", Validators: []validator.Int64{int64validator.Between(1, 65535)}},
			"listen_addresses":                 schema.SetAttribute{Required: true, ElementType: types.StringType, Validators: []validator.Set{setvalidator.SizeAtLeast(1), setvalidator.ValueStringsAre(validators.IPAddress())}, MarkdownDescription: "Explicit IPv4 and IPv6 addresses on which all Caddy HTTP and HTTPS frontends listen. A non-empty set is required to prevent wildcard listeners."},
			"acme_email":                       schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), MarkdownDescription: "Email address used by public ACME issuers such as Let's Encrypt and ZeroSSL. Defaults to `\"\"`."},
			"auto_https":                       schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), MarkdownDescription: "Global automatic HTTPS mode: empty for enabled, `off`, `disable_redirects`, `disable_certs`, or `ignore_loaded_certs`. Defaults to enabled.", Validators: []validator.String{stringvalidator.OneOf("", "off", "disable_redirects", "disable_certs", "ignore_loaded_certs")}},
			"dns_provider":                     schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), Validators: []validator.String{stringvalidator.OneOf("", "cloudflare", "rfc2136")}, MarkdownDescription: "DNS provider used for Caddy DNS operations: empty, `cloudflare`, or `rfc2136`."},
			"dns_api_key":                      schema.StringAttribute{Optional: true, Sensitive: true, WriteOnly: true, MarkdownDescription: "Write-only Cloudflare API credential. The value is never stored in Terraform plan or state. Increment `dns_credentials_version` whenever it changes."},
			"dns_api_key_configured":           schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether OPNsense currently has a Cloudflare DNS API credential configured."},
			"dns_rfc2136_server":               schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), Validators: []validator.String{validators.IPAddress()}, MarkdownDescription: "Authoritative DNS server address used for RFC2136 updates."},
			"dns_rfc2136_port":                 schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(53), Validators: []validator.Int64{int64validator.Between(1, 65535)}, MarkdownDescription: "Authoritative DNS server port used for RFC2136 updates. Defaults to `53`."},
			"dns_rfc2136_key_name":             schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), Validators: []validator.String{stringvalidator.RegexMatches(regexp.MustCompile(`^[0-9A-Za-z_.-]{0,255}$`), "must contain only letters, digits, dots, underscores, or hyphens")}, MarkdownDescription: "TSIG key name authorized for RFC2136 updates."},
			"dns_rfc2136_key_algorithm":        schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("hmac-sha256"), Validators: []validator.String{stringvalidator.OneOf("hmac-sha512", "hmac-sha384", "hmac-sha256", "hmac-sha224", "hmac-sha1")}, MarkdownDescription: "RFC2136 TSIG HMAC algorithm. Defaults to `hmac-sha256`."},
			"dns_rfc2136_key":                  schema.StringAttribute{Optional: true, Sensitive: true, WriteOnly: true, MarkdownDescription: "Write-only Base64 TSIG secret. The value is never stored in Terraform plan or state. Increment `dns_credentials_version` whenever it changes."},
			"dns_rfc2136_key_configured":       schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether OPNsense currently has an RFC2136 TSIG secret configured."},
			"dns_credentials_version":          schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(0), Validators: []validator.Int64{int64validator.AtLeast(0)}, MarkdownDescription: "Stateful rotation marker for write-only DNS credentials. Increment whenever `dns_api_key` or `dns_rfc2136_key` changes."},
			"dns_propagation_timeout_disabled": schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), MarkdownDescription: "Disable the DNS challenge propagation timeout."},
			"dns_propagation_timeout":          schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(0), Validators: []validator.Int64{int64validator.AtLeast(0)}, MarkdownDescription: "DNS propagation timeout in seconds. `0` keeps the Caddy default."},
			"dns_propagation_delay":            schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(0), Validators: []validator.Int64{int64validator.AtLeast(0)}, MarkdownDescription: "Delay in seconds before DNS propagation checks. `0` means no explicit delay."},
			"dns_propagation_resolvers":        schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), Validators: []validator.String{validators.IPAddress()}, MarkdownDescription: "Optional DNS resolver address used for propagation checks."},
			"run_as_user":                      schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("root"), MarkdownDescription: "Operating-system account used to run Caddy: `root` or `www`. Defaults to `root`.", Validators: []validator.String{stringvalidator.OneOf("root", "www")}},
			"grace_period":                     schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(10), MarkdownDescription: "Grace period in seconds for closing connections during reload. Defaults to `10`.", Validators: []validator.Int64{int64validator.Between(1, 20)}},
			"http_versions":                    schema.SetAttribute{Optional: true, Computed: true, ElementType: types.StringType, Default: setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{types.StringValue("h1"), types.StringValue("h2")})), MarkdownDescription: "Frontend HTTP versions: `h1`, `h2`, and optionally `h3`. Defaults to `h1` and `h2`."},
			"log_level":                        schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), MarkdownDescription: "Global log level. Empty means INFO; alternatives are DEBUG, WARN, ERROR, PANIC, or FATAL. Defaults to INFO.", Validators: []validator.String{stringvalidator.OneOf("", "DEBUG", "WARN", "ERROR", "PANIC", "FATAL")}},
			"plain_access_log":                 schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), MarkdownDescription: "Whether to write HTTP access logs to plain JSON files. Defaults to `false`."},
			"plain_access_log_keep":            schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(10), MarkdownDescription: "Number of rotated plain access logs to retain. Defaults to `10`.", Validators: []validator.Int64{int64validator.AtLeast(1)}},
		},
	}
}

func settingsDataSourceSchema() dschema.Schema {
	return dschema.Schema{MarkdownDescription: "Reads Caddy global settings.", Attributes: map[string]dschema.Attribute{
		"id": dschema.StringAttribute{Computed: true}, "enabled": dschema.BoolAttribute{Computed: true}, "enable_layer4": dschema.BoolAttribute{Computed: true},
		"http_port": dschema.Int64Attribute{Computed: true}, "https_port": dschema.Int64Attribute{Computed: true},
		"listen_addresses": dschema.SetAttribute{Computed: true, ElementType: types.StringType}, "acme_email": dschema.StringAttribute{Computed: true},
		"auto_https": dschema.StringAttribute{Computed: true}, "dns_provider": dschema.StringAttribute{Computed: true},
		"dns_api_key_configured": dschema.BoolAttribute{Computed: true}, "dns_rfc2136_server": dschema.StringAttribute{Computed: true},
		"dns_rfc2136_port": dschema.Int64Attribute{Computed: true}, "dns_rfc2136_key_name": dschema.StringAttribute{Computed: true},
		"dns_rfc2136_key_algorithm": dschema.StringAttribute{Computed: true}, "dns_rfc2136_key_configured": dschema.BoolAttribute{Computed: true},
		"dns_propagation_timeout_disabled": dschema.BoolAttribute{Computed: true}, "dns_propagation_timeout": dschema.Int64Attribute{Computed: true},
		"dns_propagation_delay": dschema.Int64Attribute{Computed: true}, "dns_propagation_resolvers": dschema.StringAttribute{Computed: true},
		"run_as_user": dschema.StringAttribute{Computed: true}, "grace_period": dschema.Int64Attribute{Computed: true},
		"http_versions": dschema.SetAttribute{Computed: true, ElementType: types.StringType}, "log_level": dschema.StringAttribute{Computed: true},
		"plain_access_log": dschema.BoolAttribute{Computed: true}, "plain_access_log_keep": dschema.Int64Attribute{Computed: true},
	}}
}

func settingsStructToSchema(d *apicaddy.SettingsResponse) (*settingsResourceModel, error) {
	g := d.Caddy.General
	httpPort := int64(80)
	if g.HTTPPort != "" {
		httpPort = tools.StringToInt64(g.HTTPPort)
	}
	httpsPort := int64(443)
	if g.HTTPSPort != "" {
		httpsPort = tools.StringToInt64(g.HTTPSPort)
	}
	runAs := "root"
	if g.RunAsUser.String() == "1" {
		runAs = "www"
	}
	rfcPort := int64(53)
	if g.DNSRFC2136Port != "" {
		rfcPort = tools.StringToInt64(g.DNSRFC2136Port)
	}
	rfcAlgorithm := g.DNSRFC2136KeyAlgorithm.String()
	if rfcAlgorithm == "" {
		rfcAlgorithm = "hmac-sha256"
	}
	propagationTimeout := int64(0)
	if g.DNSPropagationTimeout != "" {
		propagationTimeout = tools.StringToInt64(g.DNSPropagationTimeout)
	}
	propagationDelay := int64(0)
	if g.DNSPropagationDelay != "" {
		propagationDelay = tools.StringToInt64(g.DNSPropagationDelay)
	}
	return &settingsResourceModel{
		ID: types.StringValue("caddy_settings"), Enabled: types.BoolValue(tools.StringToBool(g.Enabled)),
		EnableLayer4: types.BoolValue(tools.StringToBool(g.EnableLayer4)), HTTPPort: types.Int64Value(httpPort),
		HTTPSPort: types.Int64Value(httpsPort), ListenAddresses: tools.StringSliceToSet([]string(g.ListenAddresses)),
		ACMEEmail: types.StringValue(g.ACMEEmail), AutoHTTPS: types.StringValue(g.AutoHTTPS.String()),
		DNSProvider: types.StringValue(g.DNSProvider.String()), DNSAPIKey: types.StringNull(),
		DNSAPIKeyConfigured: types.BoolValue(g.DNSAPIKey != ""), DNSRFC2136Server: types.StringValue(g.DNSRFC2136Server),
		DNSRFC2136Port: types.Int64Value(rfcPort), DNSRFC2136KeyName: types.StringValue(g.DNSRFC2136KeyName),
		DNSRFC2136KeyAlgorithm: types.StringValue(rfcAlgorithm), DNSRFC2136Key: types.StringNull(),
		DNSRFC2136KeyConfigured: types.BoolValue(g.DNSRFC2136Key != ""), DNSCredentialsVersion: types.Int64Null(),
		DNSPropagationTimeoutDisabled: types.BoolValue(tools.StringToBool(g.DNSPropagationTimeoutDisabled)),
		DNSPropagationTimeout:         types.Int64Value(propagationTimeout), DNSPropagationDelay: types.Int64Value(propagationDelay),
		DNSPropagationResolvers: types.StringValue(g.DNSPropagationResolvers), RunAsUser: types.StringValue(runAs),
		GracePeriod: types.Int64Value(tools.StringToInt64(g.GracePeriod)), HTTPVersions: tools.StringSliceToSet([]string(g.HTTPVersions)),
		LogLevel: types.StringValue(g.LogLevel.String()), PlainAccessLog: types.BoolValue(tools.StringToBool(g.PlainAccessLog)),
		PlainAccessLogKeep: types.Int64Value(tools.StringToInt64(g.PlainAccessLogKeep)),
	}, nil
}

func settingsStructToDataSource(d *apicaddy.SettingsResponse) (*settingsDataSourceModel, error) {
	r, err := settingsStructToSchema(d)
	if err != nil {
		return nil, err
	}
	return &settingsDataSourceModel{
		ID: r.ID, Enabled: r.Enabled, EnableLayer4: r.EnableLayer4, HTTPPort: r.HTTPPort, HTTPSPort: r.HTTPSPort,
		ListenAddresses: r.ListenAddresses, ACMEEmail: r.ACMEEmail, AutoHTTPS: r.AutoHTTPS, DNSProvider: r.DNSProvider,
		DNSAPIKeyConfigured: r.DNSAPIKeyConfigured, DNSRFC2136Server: r.DNSRFC2136Server, DNSRFC2136Port: r.DNSRFC2136Port,
		DNSRFC2136KeyName: r.DNSRFC2136KeyName, DNSRFC2136KeyAlgorithm: r.DNSRFC2136KeyAlgorithm,
		DNSRFC2136KeyConfigured: r.DNSRFC2136KeyConfigured, DNSPropagationTimeoutDisabled: r.DNSPropagationTimeoutDisabled,
		DNSPropagationTimeout: r.DNSPropagationTimeout, DNSPropagationDelay: r.DNSPropagationDelay,
		DNSPropagationResolvers: r.DNSPropagationResolvers, RunAsUser: r.RunAsUser, GracePeriod: r.GracePeriod,
		HTTPVersions: r.HTTPVersions, LogLevel: r.LogLevel, PlainAccessLog: r.PlainAccessLog, PlainAccessLogKeep: r.PlainAccessLogKeep,
	}, nil
}

func applySettingsModel(g *apicaddy.GeneralSettings, d *settingsResourceModel, dnsAPIKey, dnsRFC2136Key types.String) {
	g.Enabled = tools.BoolToString(d.Enabled.ValueBool())
	g.EnableLayer4 = tools.BoolToString(d.EnableLayer4.ValueBool())
	g.HTTPPort = tools.Int64ToString(d.HTTPPort.ValueInt64())
	g.HTTPSPort = tools.Int64ToString(d.HTTPSPort.ValueInt64())
	listenAddresses := tools.SetToStringSlice(d.ListenAddresses)
	sort.Strings(listenAddresses)
	g.ListenAddresses = api.SelectedMapList(listenAddresses)
	g.ACMEEmail = d.ACMEEmail.ValueString()
	g.AutoHTTPS = api.SelectedMap(d.AutoHTTPS.ValueString())
	g.DNSProvider = api.SelectedMap(d.DNSProvider.ValueString())
	g.DNSPropagationTimeoutDisabled = tools.BoolToString(d.DNSPropagationTimeoutDisabled.ValueBool())
	if d.DNSPropagationTimeout.ValueInt64() == 0 {
		g.DNSPropagationTimeout = ""
	} else {
		g.DNSPropagationTimeout = tools.Int64ToString(d.DNSPropagationTimeout.ValueInt64())
	}
	if d.DNSPropagationDelay.ValueInt64() == 0 {
		g.DNSPropagationDelay = ""
	} else {
		g.DNSPropagationDelay = tools.Int64ToString(d.DNSPropagationDelay.ValueInt64())
	}
	g.DNSPropagationResolvers = d.DNSPropagationResolvers.ValueString()
	switch d.DNSProvider.ValueString() {
	case "rfc2136":
		g.DNSAPIKey = ""
		g.DNSRFC2136Server = d.DNSRFC2136Server.ValueString()
		g.DNSRFC2136Port = tools.Int64ToString(d.DNSRFC2136Port.ValueInt64())
		g.DNSRFC2136KeyName = d.DNSRFC2136KeyName.ValueString()
		g.DNSRFC2136KeyAlgorithm = api.SelectedMap(d.DNSRFC2136KeyAlgorithm.ValueString())
		if !dnsRFC2136Key.IsNull() && !dnsRFC2136Key.IsUnknown() {
			g.DNSRFC2136Key = dnsRFC2136Key.ValueString()
		}
	case "cloudflare":
		if !dnsAPIKey.IsNull() && !dnsAPIKey.IsUnknown() {
			g.DNSAPIKey = dnsAPIKey.ValueString()
		}
		g.DNSRFC2136Server = ""
		g.DNSRFC2136Port = ""
		g.DNSRFC2136KeyName = ""
		g.DNSRFC2136KeyAlgorithm = api.SelectedMap("")
		g.DNSRFC2136Key = ""
	default:
		g.DNSAPIKey = ""
		g.DNSRFC2136Server = ""
		g.DNSRFC2136Port = ""
		g.DNSRFC2136KeyName = ""
		g.DNSRFC2136KeyAlgorithm = api.SelectedMap("")
		g.DNSRFC2136Key = ""
	}
	if d.RunAsUser.ValueString() == "www" {
		g.RunAsUser = api.SelectedMap("1")
	} else {
		g.RunAsUser = api.SelectedMap("0")
	}
	g.GracePeriod = tools.Int64ToString(d.GracePeriod.ValueInt64())
	g.HTTPVersions = api.SelectedMapList(tools.SetToStringSlice(d.HTTPVersions))
	g.LogLevel = api.SelectedMap(d.LogLevel.ValueString())
	g.PlainAccessLog = tools.BoolToString(d.PlainAccessLog.ValueBool())
	g.PlainAccessLogKeep = tools.Int64ToString(d.PlainAccessLogKeep.ValueInt64())
}
