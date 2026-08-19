package quagga

import (
	"context"
	"github.com/biptec/opnsense-go/pkg/api"
	"github.com/biptec/opnsense-go/pkg/quagga"
	"github.com/biptec/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ospfRouteMapModel struct {
	Enabled     types.Bool   `tfsdk:"enabled"`
	Name        types.String `tfsdk:"name"`
	Action      types.String `tfsdk:"action"`
	RouteMapID  types.Int64  `tfsdk:"route_map_id"`
	PrefixLists types.Set    `tfsdk:"prefix_lists"`
	Set         types.String `tfsdk:"set"`
	ID          types.String `tfsdk:"id"`
}

func ospfRouteMapResourceSchema() rschema.Schema {
	return rschema.Schema{Version: 1, MarkdownDescription: "Manages one OPNsense OSPFv2 route-map entry.", Attributes: map[string]rschema.Attribute{
		"enabled":      rschema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true), MarkdownDescription: "Whether to enable this route-map entry. Defaults to `true`."},
		"name":         rschema.StringAttribute{Required: true, MarkdownDescription: "Name of the route map."},
		"action":       rschema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("permit"), Validators: []validator.String{stringvalidator.OneOf("permit", "deny")}, MarkdownDescription: "Whether this route-map entry permits or denies matching routes. Defaults to `permit`."},
		"route_map_id": rschema.Int64Attribute{Required: true, Validators: []validator.Int64{int64validator.Between(10, 99)}, MarkdownDescription: "Route-map sequence ID (10-99)."},
		"prefix_lists": rschema.SetAttribute{Optional: true, Computed: true, Default: setdefault.StaticValue(tools.EmptySetValue(types.StringType)), ElementType: types.StringType, MarkdownDescription: "UUIDs of prefix-list entries matched by this route-map entry. Defaults to `[]`."},
		"set":          rschema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), MarkdownDescription: "Optional FRR route-map set expression. Defaults to an empty string."},
		"id":           rschema.StringAttribute{Computed: true, MarkdownDescription: "UUID of the route-map entry.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
	}}
}
func ospfRouteMapDataSourceSchema() dsschema.Schema {
	return dsschema.Schema{MarkdownDescription: "Reads one OPNsense OSPFv2 route-map entry by UUID.", Attributes: map[string]dsschema.Attribute{"enabled": dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether this route-map entry is enabled."}, "name": dsschema.StringAttribute{Computed: true, MarkdownDescription: "Name of the route map."}, "action": dsschema.StringAttribute{Computed: true, MarkdownDescription: "Whether this route-map entry permits or denies matching routes."}, "route_map_id": dsschema.Int64Attribute{Computed: true, MarkdownDescription: "Route-map sequence ID."}, "prefix_lists": dsschema.SetAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "UUIDs of prefix-list entries matched by this route-map entry."}, "set": dsschema.StringAttribute{Computed: true, MarkdownDescription: "FRR route-map set expression."}, "id": dsschema.StringAttribute{Required: true, MarkdownDescription: "UUID of the route-map entry."}}}
}
func ospfRouteMapToAPI(d *ospfRouteMapModel) *quagga.OSPFRouteMap {
	var prefix []string
	_ = d.PrefixLists.ElementsAs(context.Background(), &prefix, false)
	return &quagga.OSPFRouteMap{Enabled: tools.BoolToString(d.Enabled.ValueBool()), Name: d.Name.ValueString(), Action: api.SelectedMap(d.Action.ValueString()), RouteMapID: tools.Int64ToString(d.RouteMapID.ValueInt64()), PrefixList: api.SelectedMapList(prefix), Set: d.Set.ValueString()}
}
func ospfRouteMapFromAPI(d *quagga.OSPFRouteMap, id string) *ospfRouteMapModel {
	return &ospfRouteMapModel{Enabled: types.BoolValue(tools.StringToBool(d.Enabled)), Name: types.StringValue(d.Name), Action: types.StringValue(d.Action.String()), RouteMapID: types.Int64Value(tools.StringToInt64(d.RouteMapID)), PrefixLists: tools.StringSliceToSet([]string(d.PrefixList)), Set: types.StringValue(d.Set), ID: types.StringValue(id)}
}
