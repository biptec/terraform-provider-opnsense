package system

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ntpServerModel struct {
	Host     types.String `tfsdk:"host"`
	NoSelect types.Bool   `tfsdk:"noselect"`
	Prefer   types.Bool   `tfsdk:"prefer"`
	IBurst   types.Bool   `tfsdk:"iburst"`
	Pool     types.Bool   `tfsdk:"pool"`
}

type ntpSettingsResourceModel struct {
	Enabled              types.Bool   `tfsdk:"enabled"`
	Servers              types.Set    `tfsdk:"servers"`
	Interfaces           types.Set    `tfsdk:"interfaces"`
	Orphan               types.Int64  `tfsdk:"orphan"`
	MaxClock             types.Int64  `tfsdk:"max_clock"`
	ClientMode           types.Bool   `tfsdk:"client_mode"`
	KissOfDeath          types.Bool   `tfsdk:"kiss_of_death"`
	RateLimiting         types.Bool   `tfsdk:"rate_limiting"`
	DenyModifications    types.Bool   `tfsdk:"deny_modifications"`
	DisableQueries       types.Bool   `tfsdk:"disable_queries"`
	DisableServing       types.Bool   `tfsdk:"disable_serving"`
	DenyPeerAssociations types.Bool   `tfsdk:"deny_peer_associations"`
	DenyTrapService      types.Bool   `tfsdk:"deny_trap_service"`
	ID                   types.String `tfsdk:"id"`
}

var ntpServerAttributeTypes = map[string]attr.Type{
	"host":     types.StringType,
	"noselect": types.BoolType,
	"prefer":   types.BoolType,
	"iburst":   types.BoolType,
	"pool":     types.BoolType,
}

func ntpSettingsResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manages the built-in OPNsense NTP service through `os-api-extensions`. This is a singleton resource. The `os-api-extensions` package must be installed first.",
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Enable NTP synchronization and serving. Defaults to `true`.",
			},
			"servers": schema.SetNestedAttribute{
				Required:            true,
				Validators:          []validator.Set{setvalidator.SizeAtLeast(1)},
				MarkdownDescription: "Upstream NTP servers. At least one server is always required so re-enabling cannot create an unconfigured service.",
				NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
					"host": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "NTP server hostname or IP address.",
					},
					"noselect": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
						MarkdownDescription: "Collect statistics but never select this source.",
					},
					"prefer": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
						MarkdownDescription: "Prefer this source when candidates are otherwise equivalent.",
					},
					"iburst": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(true),
						MarkdownDescription: "Use an initial burst of requests for faster synchronization.",
					},
					"pool": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
						MarkdownDescription: "Treat the hostname as an NTP pool.",
					},
				}},
			},
			"interfaces": schema.SetAttribute{
				Required:            true,
				ElementType:         types.StringType,
				Validators:          []validator.Set{setvalidator.SizeAtLeast(1)},
				MarkdownDescription: "Logical OPNsense interfaces on which NTP listens and sends queries. An explicit non-empty set prevents wildcard listeners.",
			},
			"orphan": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(12),
				Validators: []validator.Int64{
					int64validator.Between(0, 15),
				},
				MarkdownDescription: "Orphan mode stratum. Defaults to `12`.",
			},
			"max_clock": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(10),
				Validators: []validator.Int64{
					int64validator.Between(2, 99),
				},
				MarkdownDescription: "Maximum retained clock associations. Defaults to `10`.",
			},
			"client_mode": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Synchronize once and exit instead of serving NTP. Defaults to `false`.",
			},
			"kiss_of_death":          secureNtpBoolean("Enable Kiss-o'-Death packets.", true),
			"rate_limiting":          secureNtpBoolean("Enable NTP rate limiting.", true),
			"deny_modifications":     secureNtpBoolean("Deny runtime state modifications from remote clients.", true),
			"disable_queries":        secureNtpBoolean("Disable remote ntpq and ntpdc queries.", true),
			"disable_serving":        secureNtpBoolean("Disable time service while retaining query access.", false),
			"deny_peer_associations": secureNtpBoolean("Deny remote peer association attempts.", true),
			"deny_trap_service":      secureNtpBoolean("Deny mode 6 trap service.", true),
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Fixed singleton identifier `ntp_settings`.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func secureNtpBoolean(description string, defaultValue bool) schema.BoolAttribute {
	return schema.BoolAttribute{
		Optional:            true,
		Computed:            true,
		Default:             booldefault.StaticBool(defaultValue),
		MarkdownDescription: description + " Defaults to `" + boolString(defaultValue) + "`.",
	}
}

func boolString(value bool) string {
	if value {
		return "true"
	}
	return "false"
}
