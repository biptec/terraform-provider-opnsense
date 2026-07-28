package routing

import (
	"regexp"

	"github.com/biptec/opnsense-go/pkg/api"
	apirouting "github.com/biptec/opnsense-go/pkg/routing"
	"github.com/biptec/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type gatewayResourceModel struct {
	Enabled                   types.Bool   `tfsdk:"enabled"`
	Name                      types.String `tfsdk:"name"`
	Description               types.String `tfsdk:"description"`
	Interface                 types.String `tfsdk:"interface"`
	IPProtocol                types.String `tfsdk:"ip_protocol"`
	Gateway                   types.String `tfsdk:"gateway"`
	DefaultGateway            types.Bool   `tfsdk:"default_gateway"`
	FarGateway                types.Bool   `tfsdk:"far_gateway"`
	MonitorDisable            types.Bool   `tfsdk:"monitor_disable"`
	MonitorNoRoute            types.Bool   `tfsdk:"monitor_no_route"`
	MonitorKillStates         types.Bool   `tfsdk:"monitor_kill_states"`
	MonitorKillStatesPriority types.Bool   `tfsdk:"monitor_kill_states_priority"`
	Monitor                   types.String `tfsdk:"monitor"`
	ForceDown                 types.Bool   `tfsdk:"force_down"`
	NoSync                    types.Bool   `tfsdk:"no_sync"`
	Priority                  types.Int64  `tfsdk:"priority"`
	Weight                    types.Int64  `tfsdk:"weight"`
	LatencyLow                types.Int64  `tfsdk:"latency_low"`
	LatencyHigh               types.Int64  `tfsdk:"latency_high"`
	LossLow                   types.Int64  `tfsdk:"loss_low"`
	LossHigh                  types.Int64  `tfsdk:"loss_high"`
	Interval                  types.Int64  `tfsdk:"interval"`
	TimePeriod                types.Int64  `tfsdk:"time_period"`
	LossInterval              types.Int64  `tfsdk:"loss_interval"`
	DataLength                types.Int64  `tfsdk:"data_length"`
	Id                        types.String `tfsdk:"id"`
}

func gatewayResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Configures an OPNsense routing gateway.",
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Enable this gateway. Defaults to `true`.",
				Optional:            true, Computed: true, Default: booldefault.StaticBool(true),
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Unique gateway name.", Required: true,
				Validators:    []validator.String{stringvalidator.RegexMatches(regexp.MustCompile(`^[a-zA-Z0-9_-]{1,32}$`), "must contain only letters, digits, underscores, or hyphens and be at most 32 characters")},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"description": schema.StringAttribute{MarkdownDescription: "Optional description.", Optional: true},
			"interface":   schema.StringAttribute{MarkdownDescription: "Logical OPNsense interface identifier, for example `wan` or `lan`.", Required: true},
			"ip_protocol": schema.StringAttribute{
				MarkdownDescription: "Address family: `inet` for IPv4 or `inet6` for IPv6.",
				Optional:            true, Computed: true, Default: stringdefault.StaticString("inet"),
				Validators: []validator.String{stringvalidator.OneOf("inet", "inet6")},
			},
			"gateway":                      schema.StringAttribute{MarkdownDescription: "Gateway IP address. Leave unset for a dynamic gateway.", Optional: true},
			"default_gateway":              schema.BoolAttribute{MarkdownDescription: "Prefer this gateway as the default gateway.", Optional: true, Computed: true, Default: booldefault.StaticBool(false)},
			"far_gateway":                  schema.BoolAttribute{MarkdownDescription: "Allow a gateway outside the directly connected subnet.", Optional: true, Computed: true, Default: booldefault.StaticBool(false)},
			"monitor_disable":              schema.BoolAttribute{MarkdownDescription: "Disable gateway monitoring. Defaults to `true`.", Optional: true, Computed: true, Default: booldefault.StaticBool(true)},
			"monitor_no_route":             schema.BoolAttribute{MarkdownDescription: "Do not add a host route for the monitor address.", Optional: true, Computed: true, Default: booldefault.StaticBool(false)},
			"monitor_kill_states":          schema.BoolAttribute{MarkdownDescription: "Kill states when the gateway is down.", Optional: true, Computed: true, Default: booldefault.StaticBool(false)},
			"monitor_kill_states_priority": schema.BoolAttribute{MarkdownDescription: "Kill states according to gateway priority changes.", Optional: true, Computed: true, Default: booldefault.StaticBool(false)},
			"monitor":                      schema.StringAttribute{MarkdownDescription: "Optional monitor IP address.", Optional: true},
			"force_down":                   schema.BoolAttribute{MarkdownDescription: "Administratively force the gateway down.", Optional: true, Computed: true, Default: booldefault.StaticBool(false)},
			"no_sync":                      schema.BoolAttribute{MarkdownDescription: "Do not synchronize this gateway to HA peers.", Optional: true, Computed: true, Default: booldefault.StaticBool(false)},
			"priority": schema.Int64Attribute{
				MarkdownDescription: "Gateway priority. Lower values are preferred. Defaults to `255`.",
				Optional:            true, Computed: true, Default: int64default.StaticInt64(255),
				Validators: []validator.Int64{int64validator.Between(0, 255)},
			},
			"weight": schema.Int64Attribute{
				MarkdownDescription: "Gateway weight for load balancing. Defaults to `1`.",
				Optional:            true, Computed: true, Default: int64default.StaticInt64(1),
				Validators: []validator.Int64{int64validator.Between(1, 10)},
			},
			"latency_low":   schema.Int64Attribute{MarkdownDescription: "Low latency threshold in milliseconds.", Optional: true, Validators: []validator.Int64{int64validator.AtLeast(1)}},
			"latency_high":  schema.Int64Attribute{MarkdownDescription: "High latency threshold in milliseconds.", Optional: true, Validators: []validator.Int64{int64validator.AtLeast(1)}},
			"loss_low":      schema.Int64Attribute{MarkdownDescription: "Low packet-loss threshold percentage.", Optional: true, Validators: []validator.Int64{int64validator.Between(1, 99)}},
			"loss_high":     schema.Int64Attribute{MarkdownDescription: "High packet-loss threshold percentage.", Optional: true, Validators: []validator.Int64{int64validator.Between(1, 100)}},
			"interval":      schema.Int64Attribute{MarkdownDescription: "Probe interval in milliseconds.", Optional: true, Validators: []validator.Int64{int64validator.AtLeast(1)}},
			"time_period":   schema.Int64Attribute{MarkdownDescription: "Monitoring time period in milliseconds.", Optional: true, Validators: []validator.Int64{int64validator.AtLeast(1)}},
			"loss_interval": schema.Int64Attribute{MarkdownDescription: "Packet-loss interval in milliseconds.", Optional: true, Validators: []validator.Int64{int64validator.AtLeast(1)}},
			"data_length":   schema.Int64Attribute{MarkdownDescription: "Probe payload length in bytes.", Optional: true, Validators: []validator.Int64{int64validator.AtLeast(0)}},
			"id":            schema.StringAttribute{Computed: true, MarkdownDescription: "UUID of the gateway.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		},
	}
}

func gatewayDataSourceSchema() dschema.Schema {
	attrs := map[string]dschema.Attribute{
		"id":                           dschema.StringAttribute{MarkdownDescription: "UUID of the gateway.", Required: true},
		"enabled":                      dschema.BoolAttribute{Computed: true},
		"name":                         dschema.StringAttribute{Computed: true},
		"description":                  dschema.StringAttribute{Computed: true},
		"interface":                    dschema.StringAttribute{Computed: true},
		"ip_protocol":                  dschema.StringAttribute{Computed: true},
		"gateway":                      dschema.StringAttribute{Computed: true},
		"default_gateway":              dschema.BoolAttribute{Computed: true},
		"far_gateway":                  dschema.BoolAttribute{Computed: true},
		"monitor_disable":              dschema.BoolAttribute{Computed: true},
		"monitor_no_route":             dschema.BoolAttribute{Computed: true},
		"monitor_kill_states":          dschema.BoolAttribute{Computed: true},
		"monitor_kill_states_priority": dschema.BoolAttribute{Computed: true},
		"monitor":                      dschema.StringAttribute{Computed: true},
		"force_down":                   dschema.BoolAttribute{Computed: true},
		"no_sync":                      dschema.BoolAttribute{Computed: true},
		"priority":                     dschema.Int64Attribute{Computed: true},
		"weight":                       dschema.Int64Attribute{Computed: true},
		"latency_low":                  dschema.Int64Attribute{Computed: true},
		"latency_high":                 dschema.Int64Attribute{Computed: true},
		"loss_low":                     dschema.Int64Attribute{Computed: true},
		"loss_high":                    dschema.Int64Attribute{Computed: true},
		"interval":                     dschema.Int64Attribute{Computed: true},
		"time_period":                  dschema.Int64Attribute{Computed: true},
		"loss_interval":                dschema.Int64Attribute{Computed: true},
		"data_length":                  dschema.Int64Attribute{Computed: true},
	}
	return dschema.Schema{MarkdownDescription: "Reads an OPNsense routing gateway by UUID.", Attributes: attrs}
}

func optionalIntToString(value types.Int64) string {
	if value.IsNull() || value.IsUnknown() {
		return ""
	}
	return tools.Int64ToString(value.ValueInt64())
}

func convertGatewaySchemaToStruct(d *gatewayResourceModel) (*apirouting.Gateway, error) {
	return &apirouting.Gateway{
		Disabled: tools.BoolToString(!d.Enabled.ValueBool()), Name: d.Name.ValueString(),
		Description: d.Description.ValueString(), Interface: api.SelectedMap(d.Interface.ValueString()),
		IPProtocol: api.SelectedMap(d.IPProtocol.ValueString()), Gateway: d.Gateway.ValueString(),
		DefaultGateway: tools.BoolToString(d.DefaultGateway.ValueBool()), FarGateway: tools.BoolToString(d.FarGateway.ValueBool()),
		MonitorDisable: tools.BoolToString(d.MonitorDisable.ValueBool()), MonitorNoRoute: tools.BoolToString(d.MonitorNoRoute.ValueBool()),
		MonitorKillStates: tools.BoolToString(d.MonitorKillStates.ValueBool()), MonitorKillStatesPriority: tools.BoolToString(d.MonitorKillStatesPriority.ValueBool()),
		Monitor: d.Monitor.ValueString(), ForceDown: tools.BoolToString(d.ForceDown.ValueBool()), NoSync: tools.BoolToString(d.NoSync.ValueBool()),
		Priority: tools.Int64ToString(d.Priority.ValueInt64()), Weight: tools.Int64ToString(d.Weight.ValueInt64()),
		LatencyLow: optionalIntToString(d.LatencyLow), LatencyHigh: optionalIntToString(d.LatencyHigh),
		LossLow: optionalIntToString(d.LossLow), LossHigh: optionalIntToString(d.LossHigh),
		Interval: optionalIntToString(d.Interval), TimePeriod: optionalIntToString(d.TimePeriod),
		LossInterval: optionalIntToString(d.LossInterval), DataLength: optionalIntToString(d.DataLength),
	}, nil
}

func convertGatewayStructToSchema(d *apirouting.Gateway) (*gatewayResourceModel, error) {
	return &gatewayResourceModel{
		Enabled: types.BoolValue(!tools.StringToBool(d.Disabled)),
		Name:    types.StringValue(d.Name), Description: tools.StringOrNull(d.Description),
		Interface: types.StringValue(d.Interface.String()), IPProtocol: types.StringValue(d.IPProtocol.String()),
		Gateway: tools.StringOrNull(d.Gateway), DefaultGateway: types.BoolValue(tools.StringToBool(d.DefaultGateway)),
		FarGateway: types.BoolValue(tools.StringToBool(d.FarGateway)), MonitorDisable: types.BoolValue(tools.StringToBool(d.MonitorDisable)),
		MonitorNoRoute: types.BoolValue(tools.StringToBool(d.MonitorNoRoute)), MonitorKillStates: types.BoolValue(tools.StringToBool(d.MonitorKillStates)),
		MonitorKillStatesPriority: types.BoolValue(tools.StringToBool(d.MonitorKillStatesPriority)), Monitor: tools.StringOrNull(d.Monitor),
		ForceDown: types.BoolValue(tools.StringToBool(d.ForceDown)), NoSync: types.BoolValue(tools.StringToBool(d.NoSync)),
		Priority: types.Int64Value(tools.StringToInt64(d.Priority)), Weight: types.Int64Value(tools.StringToInt64(d.Weight)),
		LatencyLow: tools.StringToInt64Null(d.LatencyLow), LatencyHigh: tools.StringToInt64Null(d.LatencyHigh),
		LossLow: tools.StringToInt64Null(d.LossLow), LossHigh: tools.StringToInt64Null(d.LossHigh),
		Interval: tools.StringToInt64Null(d.Interval), TimePeriod: tools.StringToInt64Null(d.TimePeriod),
		LossInterval: tools.StringToInt64Null(d.LossInterval), DataLength: tools.StringToInt64Null(d.DataLength),
	}, nil
}
