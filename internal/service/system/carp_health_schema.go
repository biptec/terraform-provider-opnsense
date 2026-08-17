package system

import (
	"github.com/biptec/terraform-provider-opnsense/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
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
	Enabled   types.Bool   `tfsdk:"enabled"`
	Name      types.String `tfsdk:"name"`
	Interface types.String `tfsdk:"interface"`
	Target    types.String `tfsdk:"target"`
	Scope     types.String `tfsdk:"scope"`
	VHID      types.Int64  `tfsdk:"vhid"`
	ID        types.String `tfsdk:"id"`
}

func carpHealthResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manages CARP L2 health-monitor settings through `os-api-extensions`. Health checks may demote all local CARP VHIDs or only a selected VHID.",
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
	return schema.Schema{
		MarkdownDescription: "Manages one CARP L2 health check through `os-api-extensions`. The ARP probe uses source IPv4 0.0.0.0, so the same check works while the node is CARP MASTER or BACKUP.",
		Attributes: map[string]schema.Attribute{
			"enabled":   schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true), MarkdownDescription: "Enable this health check."},
			"name":      schema.StringAttribute{Required: true, MarkdownDescription: "Stable check name."},
			"interface": schema.StringAttribute{Required: true, MarkdownDescription: "Friendly OPNsense logical interface, for example `opt2`."},
			"target":    schema.StringAttribute{Required: true, Validators: []validator.String{validators.IPv4Address()}, MarkdownDescription: "IPv4 address that must answer ARP on the selected L2 segment."},
			"scope":     schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("global"), Validators: []validator.String{stringvalidator.OneOf("global", "vhid")}, MarkdownDescription: "CARP demotion scope: `global` or `vhid`."},
			"vhid":      schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(0), Validators: []validator.Int64{int64validator.Between(0, 255)}, MarkdownDescription: "CARP VHID when `scope = \"vhid\"`; zero is the inactive value for global scope."},
			"id":        schema.StringAttribute{Computed: true, MarkdownDescription: "UUID of the health check.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		},
	}
}
