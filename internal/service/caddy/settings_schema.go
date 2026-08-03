package caddy

import (
	"github.com/biptec/opnsense-go/pkg/api"
	apicaddy "github.com/biptec/opnsense-go/pkg/caddy"
	"github.com/biptec/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
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
)

type settingsResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	Enabled            types.Bool   `tfsdk:"enabled"`
	EnableLayer4       types.Bool   `tfsdk:"enable_layer4"`
	HTTPPort           types.Int64  `tfsdk:"http_port"`
	HTTPSPort          types.Int64  `tfsdk:"https_port"`
	ACMEEmail          types.String `tfsdk:"acme_email"`
	AutoHTTPS          types.String `tfsdk:"auto_https"`
	RunAsUser          types.String `tfsdk:"run_as_user"`
	GracePeriod        types.Int64  `tfsdk:"grace_period"`
	HTTPVersions       types.Set    `tfsdk:"http_versions"`
	LogLevel           types.String `tfsdk:"log_level"`
	PlainAccessLog     types.Bool   `tfsdk:"plain_access_log"`
	PlainAccessLogKeep types.Int64  `tfsdk:"plain_access_log_keep"`
}

func settingsResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manages Caddy global settings. This singleton must be imported with `caddy_settings` before use. Destroy removes it from Terraform state without changing OPNsense.", Version: 1,
		Attributes: map[string]schema.Attribute{
			"id":                    schema.StringAttribute{Computed: true, MarkdownDescription: "Always `caddy_settings`.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"enabled":               schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), MarkdownDescription: "Whether to enable Caddy. Defaults to `false`."},
			"enable_layer4":         schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), MarkdownDescription: "Whether to enable Layer 4 routing. Defaults to `false`."},
			"http_port":             schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(80), MarkdownDescription: "Local HTTP listen port. Defaults to `80`.", Validators: []validator.Int64{int64validator.Between(1, 65535)}},
			"https_port":            schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(443), MarkdownDescription: "Local HTTPS listen port. Defaults to `443`.", Validators: []validator.Int64{int64validator.Between(1, 65535)}},
			"acme_email":            schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), MarkdownDescription: "Email address used by public ACME issuers such as Let's Encrypt and ZeroSSL. Defaults to `\"\"`."},
			"auto_https":            schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), MarkdownDescription: "Global automatic HTTPS mode: empty for enabled, `off`, `disable_redirects`, `disable_certs`, or `ignore_loaded_certs`. Defaults to enabled.", Validators: []validator.String{stringvalidator.OneOf("", "off", "disable_redirects", "disable_certs", "ignore_loaded_certs")}},
			"run_as_user":           schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("root"), MarkdownDescription: "Operating-system account used to run Caddy: `root` or `www`. Defaults to `root`.", Validators: []validator.String{stringvalidator.OneOf("root", "www")}},
			"grace_period":          schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(10), MarkdownDescription: "Grace period in seconds for closing connections during reload. Defaults to `10`.", Validators: []validator.Int64{int64validator.Between(1, 20)}},
			"http_versions":         schema.SetAttribute{Optional: true, Computed: true, ElementType: types.StringType, Default: setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{types.StringValue("h1"), types.StringValue("h2")})), MarkdownDescription: "Frontend HTTP versions: `h1`, `h2`, and optionally `h3`. Defaults to `h1` and `h2`."},
			"log_level":             schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), MarkdownDescription: "Global log level. Empty means INFO; alternatives are DEBUG, WARN, ERROR, PANIC, or FATAL. Defaults to INFO.", Validators: []validator.String{stringvalidator.OneOf("", "DEBUG", "WARN", "ERROR", "PANIC", "FATAL")}},
			"plain_access_log":      schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), MarkdownDescription: "Whether to write HTTP access logs to plain JSON files. Defaults to `false`."},
			"plain_access_log_keep": schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(10), MarkdownDescription: "Number of rotated plain access logs to retain. Defaults to `10`.", Validators: []validator.Int64{int64validator.AtLeast(1)}},
		},
	}
}

func settingsDataSourceSchema() dschema.Schema {
	return dschema.Schema{MarkdownDescription: "Reads Caddy global settings.", Attributes: map[string]dschema.Attribute{
		"id": dschema.StringAttribute{Computed: true}, "enabled": dschema.BoolAttribute{Computed: true}, "enable_layer4": dschema.BoolAttribute{Computed: true},
		"http_port": dschema.Int64Attribute{Computed: true}, "https_port": dschema.Int64Attribute{Computed: true}, "acme_email": dschema.StringAttribute{Computed: true},
		"auto_https": dschema.StringAttribute{Computed: true}, "run_as_user": dschema.StringAttribute{Computed: true}, "grace_period": dschema.Int64Attribute{Computed: true},
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
	return &settingsResourceModel{ID: types.StringValue("caddy_settings"), Enabled: types.BoolValue(tools.StringToBool(g.Enabled)), EnableLayer4: types.BoolValue(tools.StringToBool(g.EnableLayer4)), HTTPPort: types.Int64Value(httpPort), HTTPSPort: types.Int64Value(httpsPort), ACMEEmail: types.StringValue(g.ACMEEmail), AutoHTTPS: types.StringValue(g.AutoHTTPS.String()), RunAsUser: types.StringValue(runAs), GracePeriod: types.Int64Value(tools.StringToInt64(g.GracePeriod)), HTTPVersions: tools.StringSliceToSet([]string(g.HTTPVersions)), LogLevel: types.StringValue(g.LogLevel.String()), PlainAccessLog: types.BoolValue(tools.StringToBool(g.PlainAccessLog)), PlainAccessLogKeep: types.Int64Value(tools.StringToInt64(g.PlainAccessLogKeep))}, nil
}

func applySettingsModel(g *apicaddy.GeneralSettings, d *settingsResourceModel) {
	g.Enabled = tools.BoolToString(d.Enabled.ValueBool())
	g.EnableLayer4 = tools.BoolToString(d.EnableLayer4.ValueBool())
	g.HTTPPort = tools.Int64ToString(d.HTTPPort.ValueInt64())
	g.HTTPSPort = tools.Int64ToString(d.HTTPSPort.ValueInt64())
	g.ACMEEmail = d.ACMEEmail.ValueString()
	g.AutoHTTPS = api.SelectedMap(d.AutoHTTPS.ValueString())
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
