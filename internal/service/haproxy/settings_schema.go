package haproxy

import (
	apihaproxy "github.com/biptec/opnsense-go/pkg/haproxy"
	"github.com/biptec/terraform-provider-opnsense/internal/tools"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const settingsID = "haproxy_settings"

type settingsModel struct {
	Enabled         types.Bool   `tfsdk:"enabled"`
	GracefulStop    types.Bool   `tfsdk:"graceful_stop"`
	HardStopAfter   types.String `tfsdk:"hard_stop_after"`
	CloseSpreadTime types.String `tfsdk:"close_spread_time"`
	SeamlessReload  types.Bool   `tfsdk:"seamless_reload"`
	ID              types.String `tfsdk:"id"`
}

func settingsResourceSchema() schema.Schema {
	return schema.Schema{Version: 1, MarkdownDescription: "Manages global OPNsense HAProxy settings. This singleton must be imported before use.", Attributes: map[string]schema.Attribute{
		"enabled": schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Whether the HAProxy service is enabled."}, "graceful_stop": schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Whether HAProxy uses graceful stop behavior."},
		"hard_stop_after": schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Maximum graceful shutdown duration, such as `60s`."}, "close_spread_time": schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Optional connection close spread time."},
		"seamless_reload": schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Whether seamless reload is enabled."},
		"id":              schema.StringAttribute{Computed: true, MarkdownDescription: "Fixed singleton identifier `haproxy_settings`.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
	}}
}
func settingsDataSourceSchema() dschema.Schema {
	return dschema.Schema{MarkdownDescription: "Reads global OPNsense HAProxy settings.", Attributes: map[string]dschema.Attribute{
		"enabled": dschema.BoolAttribute{Computed: true}, "graceful_stop": dschema.BoolAttribute{Computed: true}, "hard_stop_after": dschema.StringAttribute{Computed: true}, "close_spread_time": dschema.StringAttribute{Computed: true}, "seamless_reload": dschema.BoolAttribute{Computed: true}, "id": dschema.StringAttribute{Computed: true},
	}}
}
func settingsAPIToModel(d *apihaproxy.SettingsResponse) *settingsModel {
	g := d.HAProxy.General
	return &settingsModel{Enabled: types.BoolValue(tools.StringToBool(g.Enabled)), GracefulStop: types.BoolValue(tools.StringToBool(g.GracefulStop)), HardStopAfter: tools.StringOrNull(g.HardStopAfter), CloseSpreadTime: tools.StringOrNull(g.CloseSpreadTime), SeamlessReload: types.BoolValue(tools.StringToBool(g.SeamlessReload)), ID: types.StringValue(settingsID)}
}
func applySettingsModel(g *apihaproxy.GeneralSettings, d *settingsModel) {
	g.Enabled = tools.BoolToString(d.Enabled.ValueBool())
	g.GracefulStop = tools.BoolToString(d.GracefulStop.ValueBool())
	g.HardStopAfter = d.HardStopAfter.ValueString()
	g.CloseSpreadTime = d.CloseSpreadTime.ValueString()
	g.SeamlessReload = tools.BoolToString(d.SeamlessReload.ValueBool())
}
