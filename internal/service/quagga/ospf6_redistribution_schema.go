package quagga

import (
	"github.com/biptec/opnsense-go/pkg/api"
	"github.com/biptec/opnsense-go/pkg/quagga"
	"github.com/biptec/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ospf6RedistributionModel struct {
	Enabled      types.Bool   `tfsdk:"enabled"`
	Description  types.String `tfsdk:"description"`
	Redistribute types.String `tfsdk:"redistribute"`
	RouteMap     types.String `tfsdk:"route_map"`
	ID           types.String `tfsdk:"id"`
}

func ospf6RedistributionResourceSchema() rschema.Schema {
	return rschema.Schema{Version: 1, MarkdownDescription: "Manages one OPNsense OSPFv3 redistribution rule.", Attributes: map[string]rschema.Attribute{"enabled": rschema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true), MarkdownDescription: "Whether to enable this redistribution rule. Defaults to `true`."}, "description": rschema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), MarkdownDescription: "Optional description for this redistribution rule. Defaults to an empty string."}, "redistribute": rschema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("connected"), Validators: []validator.String{stringvalidator.OneOf("bgp", "connected", "kernel", "rip", "static")}, MarkdownDescription: "Routing source redistributed into OSPFv3. Defaults to `connected`."}, "route_map": rschema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), MarkdownDescription: "UUID of the route map applied to redistributed routes. Defaults to an empty string."}, "id": rschema.StringAttribute{Computed: true, MarkdownDescription: "UUID of the redistribution rule.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}}}}
}
func ospf6RedistributionDataSourceSchema() dsschema.Schema {
	return dsschema.Schema{MarkdownDescription: "Reads one OPNsense OSPFv3 redistribution rule by UUID.", Attributes: map[string]dsschema.Attribute{"enabled": dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether this redistribution rule is enabled."}, "description": dsschema.StringAttribute{Computed: true, MarkdownDescription: "Description for this redistribution rule."}, "redistribute": dsschema.StringAttribute{Computed: true, MarkdownDescription: "Routing source redistributed into OSPFv3."}, "route_map": dsschema.StringAttribute{Computed: true, MarkdownDescription: "UUID of the route map applied to redistributed routes."}, "id": dsschema.StringAttribute{Required: true, MarkdownDescription: "UUID of the redistribution rule."}}}
}
func ospf6RedistributionToAPI(d *ospf6RedistributionModel) *quagga.OSPF6Redistribution {
	return &quagga.OSPF6Redistribution{Enabled: tools.BoolToString(d.Enabled.ValueBool()), Description: d.Description.ValueString(), Redistribute: api.SelectedMap(d.Redistribute.ValueString()), RouteMap: api.SelectedMap(d.RouteMap.ValueString())}
}
func ospf6RedistributionFromAPI(d *quagga.OSPF6Redistribution, id string) *ospf6RedistributionModel {
	return &ospf6RedistributionModel{Enabled: types.BoolValue(tools.StringToBool(d.Enabled)), Description: types.StringValue(d.Description), Redistribute: types.StringValue(d.Redistribute.String()), RouteMap: types.StringValue(d.RouteMap.String()), ID: types.StringValue(id)}
}
