package system

import (
	"regexp"

	"github.com/biptec/terraform-provider-opnsense/internal/tools"
	"github.com/biptec/terraform-provider-opnsense/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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

type carpHealthResourceModel struct {
	Enabled           types.Bool   `tfsdk:"enabled"`
	Interval          types.Int64  `tfsdk:"interval"`
	FailureThreshold  types.Int64  `tfsdk:"failure_threshold"`
	RecoveryThreshold types.Int64  `tfsdk:"recovery_threshold"`
	ID                types.String `tfsdk:"id"`
}

type carpHealthCheckResourceModel struct {
	Enabled                    types.Bool   `tfsdk:"enabled"`
	Name                       types.String `tfsdk:"name"`
	Interface                  types.String `tfsdk:"interface"`
	Target                     types.String `tfsdk:"target"`
	Scope                      types.String `tfsdk:"scope"`
	VHID                       types.Int64  `tfsdk:"vhid"`
	FailureAdvSkew             types.Int64  `tfsdk:"failure_advskew"`
	VHIDTargets                types.Set    `tfsdk:"vhid_targets"`
	FallbackIPv4Target         types.String `tfsdk:"fallback_ipv4_target"`
	FallbackIPv4Gateway        types.String `tfsdk:"fallback_ipv4_gateway"`
	FallbackIPv6Target         types.String `tfsdk:"fallback_ipv6_target"`
	FallbackIPv6Gateway        types.String `tfsdk:"fallback_ipv6_gateway"`
	FallbackIPv4DefaultGateway types.String `tfsdk:"fallback_ipv4_default_gateway"`
	FallbackIPv6DefaultGateway types.String `tfsdk:"fallback_ipv6_default_gateway"`
	ID                         types.String `tfsdk:"id"`
}

func carpHealthResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manages CARP L2 health-monitor settings through `os-api-extensions`. Health checks can automatically discover CARP targets or explicitly select VHIDs.",
		Attributes: map[string]schema.Attribute{
			"enabled":            schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), MarkdownDescription: "Enable CARP L2 health monitoring."},
			"interval":           schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(1), Validators: []validator.Int64{int64validator.Between(1, 60)}, MarkdownDescription: "Probe interval in seconds."},
			"failure_threshold":  schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(2), Validators: []validator.Int64{int64validator.Between(1, 20)}, MarkdownDescription: "Consecutive failed probes required before CARP demotion."},
			"recovery_threshold": schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(2), Validators: []validator.Int64{int64validator.Between(1, 20)}, MarkdownDescription: "Consecutive successful probes required before CARP promotion is allowed again."},
			"id":                 schema.StringAttribute{Computed: true, MarkdownDescription: "Fixed singleton identifier `carp_health`.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		},
	}
}

func carpHealthCheckResourceSchema() schema.Schema {
	emptySet := setdefault.StaticValue(tools.EmptySetValue(types.StringType))
	vhidTargetPattern := regexp.MustCompile(`^[0-9A-Za-z._-]+:(?:[1-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`)
	return schema.Schema{
		MarkdownDescription: "Manages one CARP L2 health check through `os-api-extensions`. Automatic scope discovery is preferred; explicit VHID targeting remains available for exceptions.",
		Attributes: map[string]schema.Attribute{
			"enabled":   schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true), MarkdownDescription: "Enable this health check."},
			"name":      schema.StringAttribute{Required: true, MarkdownDescription: "Stable check name."},
			"interface": schema.StringAttribute{Required: true, MarkdownDescription: "Friendly OPNsense logical interface used for the ARP probe, for example `opt2`."},
			"target":    schema.StringAttribute{Required: true, Validators: []validator.String{validators.IPv4Address()}, MarkdownDescription: "IPv4 address that must answer ARP on the selected L2 segment."},
			"scope": schema.StringAttribute{
				Optional: true, Computed: true, Default: stringdefault.StaticString("interface"),
				Validators:          []validator.String{stringvalidator.OneOf("interface", "all_carp", "vhid", "vhid_group", "global")},
				MarkdownDescription: "CARP target scope. `interface` automatically discovers CARP on the probe interface; `all_carp` discovers all configured CARP; `vhid` and `vhid_group` are explicit overrides; `global` preserves legacy global demotion.",
			},
			"vhid": schema.Int64Attribute{
				Optional: true, Computed: true, Default: int64default.StaticInt64(0), Validators: []validator.Int64{int64validator.Between(0, 255)},
				MarkdownDescription: "Explicit CARP VHID when `scope = \"vhid\"`. Keep zero for every other scope.",
			},
			"failure_advskew": schema.Int64Attribute{
				Optional: true, Computed: true, Default: int64default.StaticInt64(254), Validators: []validator.Int64{int64validator.Between(1, 254)},
				MarkdownDescription: "advskew enforced while this check is unhealthy. Use 254 for a hard local-segment failure; lower values can provide softer preference changes.",
			},
			"vhid_targets": schema.SetAttribute{
				Optional: true, Computed: true, Default: emptySet, ElementType: types.StringType,
				Validators:          []validator.Set{setvalidator.ValueStringsAre(stringvalidator.RegexMatches(vhidTargetPattern, "must use logical-interface:VHID with VHID between 1 and 255"))},
				MarkdownDescription: "Explicit `interface:VHID` targets when `scope = \"vhid_group\"`. Empty for automatic and single-VHID scopes.",
			},
			"fallback_ipv4_target":          schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), Validators: []validator.String{validators.IPv4Address()}, MarkdownDescription: "Optional IPv4 host route destination installed while unhealthy. Configure together with fallback_ipv4_gateway."},
			"fallback_ipv4_gateway":         schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), Validators: []validator.String{validators.IPv4Address()}, MarkdownDescription: "IPv4 peer next hop for the fallback host route. Configure together with fallback_ipv4_target."},
			"fallback_ipv6_target":          schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), Validators: []validator.String{validators.IPv6Address()}, MarkdownDescription: "Optional IPv6 host route destination installed while unhealthy. Configure together with fallback_ipv6_gateway."},
			"fallback_ipv6_gateway":         schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), Validators: []validator.String{validators.IPv6Address()}, MarkdownDescription: "IPv6 peer next hop for the fallback host route. Configure together with fallback_ipv6_target."},
			"fallback_ipv4_default_gateway": schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), Validators: []validator.String{validators.IPv4Address()}, MarkdownDescription: "Optional IPv4 peer next hop for default/Internet traffic after this health check has failed and its CARP target has remained BACKUP for a full monitor cycle."},
			"fallback_ipv6_default_gateway": schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), Validators: []validator.String{validators.IPv6Address()}, MarkdownDescription: "Optional IPv6 peer next hop for default/Internet traffic after this health check has failed and its CARP target has remained BACKUP for a full monitor cycle."},
			"id":                            schema.StringAttribute{Computed: true, MarkdownDescription: "UUID of the health check.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		},
	}
}
