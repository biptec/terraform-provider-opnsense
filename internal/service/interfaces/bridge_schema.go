package interfaces

import (
	"fmt"

	"github.com/biptec/opnsense-go/pkg/api"
	apiinterfaces "github.com/biptec/opnsense-go/pkg/interfaces"
	"github.com/biptec/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type bridgeResourceModel struct {
	Device           types.String `tfsdk:"device"`
	Members          types.Set    `tfsdk:"members"`
	LinkLocal        types.Bool   `tfsdk:"link_local"`
	EnableSTP        types.Bool   `tfsdk:"enable_stp"`
	Protocol         types.String `tfsdk:"protocol"`
	STPMembers       types.Set    `tfsdk:"stp_members"`
	MaxAge           types.Int64  `tfsdk:"max_age"`
	ForwardDelay     types.Int64  `tfsdk:"forward_delay"`
	HoldCount        types.Int64  `tfsdk:"hold_count"`
	MaxAddresses     types.Int64  `tfsdk:"max_addresses"`
	Timeout          types.Int64  `tfsdk:"timeout"`
	Span             types.String `tfsdk:"span"`
	Edge             types.Set    `tfsdk:"edge"`
	AutoEdge         types.Set    `tfsdk:"auto_edge"`
	PointToPoint     types.Set    `tfsdk:"point_to_point"`
	AutoPointToPoint types.Set    `tfsdk:"auto_point_to_point"`
	Static           types.Set    `tfsdk:"sticky"`
	Private          types.Set    `tfsdk:"private"`
	Description      types.String `tfsdk:"description"`
	Id               types.String `tfsdk:"id"`
}

func bridgeResourceSchema() schema.Schema {
	emptySet := setdefault.StaticValue(tools.EmptySetValue(types.StringType))
	return schema.Schema{
		MarkdownDescription: "Manages an OPNsense bridge interface and its spanning-tree options.",
		Attributes: map[string]schema.Attribute{
			"device":              schema.StringAttribute{Computed: true, MarkdownDescription: "Automatically allocated bridge device name."},
			"members":             schema.SetAttribute{Required: true, ElementType: types.StringType, Validators: []validator.Set{setvalidator.SizeAtLeast(1)}, MarkdownDescription: "Interfaces participating in the bridge."},
			"link_local":          schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), MarkdownDescription: "Enable a link-local address on the bridge."},
			"enable_stp":          schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), MarkdownDescription: "Enable spanning-tree processing."},
			"protocol":            schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("rstp"), Validators: []validator.String{stringvalidator.OneOf("rstp", "stp")}, MarkdownDescription: "Spanning-tree protocol."},
			"stp_members":         optionalInterfaceSet("Interfaces on which STP is enabled.", emptySet),
			"max_age":             schema.Int64Attribute{Optional: true, Validators: []validator.Int64{int64validator.Between(6, 40)}, MarkdownDescription: "STP configuration validity time."},
			"forward_delay":       schema.Int64Attribute{Optional: true, Validators: []validator.Int64{int64validator.Between(4, 30)}, MarkdownDescription: "STP forwarding delay."},
			"hold_count":          schema.Int64Attribute{Optional: true, Validators: []validator.Int64{int64validator.Between(1, 10)}, MarkdownDescription: "STP transmit hold count."},
			"max_addresses":       schema.Int64Attribute{Optional: true, Validators: []validator.Int64{int64validator.AtLeast(1)}, MarkdownDescription: "Bridge address-cache size."},
			"timeout":             schema.Int64Attribute{Optional: true, Validators: []validator.Int64{int64validator.AtLeast(0)}, MarkdownDescription: "Bridge address-cache timeout in seconds."},
			"span":                schema.StringAttribute{Optional: true, MarkdownDescription: "Optional span port. It cannot also be a bridge member."},
			"edge":                optionalInterfaceSet("Edge ports.", emptySet),
			"auto_edge":           optionalInterfaceSet("Ports for which automatic edge detection is disabled.", emptySet),
			"point_to_point":      optionalInterfaceSet("Point-to-point ports.", emptySet),
			"auto_point_to_point": optionalInterfaceSet("Ports for which automatic point-to-point detection is disabled.", emptySet),
			"sticky":              optionalInterfaceSet("Sticky ports.", emptySet),
			"private":             optionalInterfaceSet("Private ports.", emptySet),
			"description":         schema.StringAttribute{Optional: true, MarkdownDescription: "Optional description."},
			"id":                  schema.StringAttribute{Computed: true, MarkdownDescription: "UUID of the bridge configuration.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		},
	}
}

func optionalInterfaceSet(description string, defaultValue defaults.Set) schema.SetAttribute {
	return schema.SetAttribute{Optional: true, Computed: true, Default: defaultValue, ElementType: types.StringType, MarkdownDescription: description}
}

func bridgeDataSourceSchema() dschema.Schema {
	attrs := map[string]dschema.Attribute{
		"id": dschema.StringAttribute{Required: true}, "device": dschema.StringAttribute{Computed: true},
		"members": dschema.SetAttribute{Computed: true, ElementType: types.StringType}, "link_local": dschema.BoolAttribute{Computed: true},
		"enable_stp": dschema.BoolAttribute{Computed: true}, "protocol": dschema.StringAttribute{Computed: true},
		"stp_members": dschema.SetAttribute{Computed: true, ElementType: types.StringType}, "max_age": dschema.Int64Attribute{Computed: true},
		"forward_delay": dschema.Int64Attribute{Computed: true}, "hold_count": dschema.Int64Attribute{Computed: true},
		"max_addresses": dschema.Int64Attribute{Computed: true}, "timeout": dschema.Int64Attribute{Computed: true},
		"span": dschema.StringAttribute{Computed: true}, "edge": dschema.SetAttribute{Computed: true, ElementType: types.StringType},
		"auto_edge": dschema.SetAttribute{Computed: true, ElementType: types.StringType}, "point_to_point": dschema.SetAttribute{Computed: true, ElementType: types.StringType},
		"auto_point_to_point": dschema.SetAttribute{Computed: true, ElementType: types.StringType}, "sticky": dschema.SetAttribute{Computed: true, ElementType: types.StringType},
		"private": dschema.SetAttribute{Computed: true, ElementType: types.StringType}, "description": dschema.StringAttribute{Computed: true},
	}
	return dschema.Schema{MarkdownDescription: "Reads an OPNsense bridge interface.", Attributes: attrs}
}

func convertBridgeSchemaToStruct(d *bridgeResourceModel) (*apiinterfaces.Bridge, error) {
	members := tools.SetToStringSlice(d.Members)
	memberSet := make(map[string]struct{}, len(members))
	for _, member := range members {
		memberSet[member] = struct{}{}
	}
	if !d.Span.IsNull() && !d.Span.IsUnknown() && d.Span.ValueString() != "" {
		if _, exists := memberSet[d.Span.ValueString()]; exists {
			return nil, fmt.Errorf("span must not also be a bridge member")
		}
	}
	for name, set := range map[string]types.Set{"stp_members": d.STPMembers, "edge": d.Edge, "auto_edge": d.AutoEdge, "point_to_point": d.PointToPoint, "auto_point_to_point": d.AutoPointToPoint, "sticky": d.Static, "private": d.Private} {
		for _, value := range tools.SetToStringSlice(set) {
			if _, exists := memberSet[value]; !exists {
				return nil, fmt.Errorf("%s contains non-member interface %q", name, value)
			}
		}
	}
	return &apiinterfaces.Bridge{
		Members: api.SelectedMapList(members), LinkLocal: tools.BoolToString(d.LinkLocal.ValueBool()), EnableSTP: tools.BoolToString(d.EnableSTP.ValueBool()),
		Protocol: api.SelectedMap(d.Protocol.ValueString()), STPMembers: api.SelectedMapList(tools.SetToStringSlice(d.STPMembers)),
		MaxAge: optionalIntString(d.MaxAge), ForwardDelay: optionalIntString(d.ForwardDelay), HoldCount: optionalIntString(d.HoldCount),
		MaxAddresses: optionalIntString(d.MaxAddresses), Timeout: optionalIntString(d.Timeout), Span: api.SelectedMap(d.Span.ValueString()),
		Edge: api.SelectedMapList(tools.SetToStringSlice(d.Edge)), AutoEdge: api.SelectedMapList(tools.SetToStringSlice(d.AutoEdge)),
		PointToPoint: api.SelectedMapList(tools.SetToStringSlice(d.PointToPoint)), AutoPointToPoint: api.SelectedMapList(tools.SetToStringSlice(d.AutoPointToPoint)),
		Static: api.SelectedMapList(tools.SetToStringSlice(d.Static)), Private: api.SelectedMapList(tools.SetToStringSlice(d.Private)), Description: d.Description.ValueString(),
	}, nil
}

func convertBridgeStructToSchema(d *apiinterfaces.Bridge) (*bridgeResourceModel, error) {
	protocol := d.Protocol.String()
	if protocol == "" {
		protocol = "rstp"
	}
	return &bridgeResourceModel{
		Device: types.StringValue(d.Device), Members: tools.StringSliceToSet([]string(d.Members)), LinkLocal: types.BoolValue(tools.StringToBool(d.LinkLocal)),
		EnableSTP: types.BoolValue(tools.StringToBool(d.EnableSTP)), Protocol: types.StringValue(protocol), STPMembers: tools.StringSliceToSet([]string(d.STPMembers)),
		MaxAge: tools.StringToInt64Null(d.MaxAge), ForwardDelay: tools.StringToInt64Null(d.ForwardDelay), HoldCount: tools.StringToInt64Null(d.HoldCount),
		MaxAddresses: tools.StringToInt64Null(d.MaxAddresses), Timeout: tools.StringToInt64Null(d.Timeout), Span: tools.StringOrNull(d.Span.String()),
		Edge: tools.StringSliceToSet([]string(d.Edge)), AutoEdge: tools.StringSliceToSet([]string(d.AutoEdge)), PointToPoint: tools.StringSliceToSet([]string(d.PointToPoint)),
		AutoPointToPoint: tools.StringSliceToSet([]string(d.AutoPointToPoint)), Static: tools.StringSliceToSet([]string(d.Static)), Private: tools.StringSliceToSet([]string(d.Private)),
		Description: tools.StringOrNull(d.Description),
	}, nil
}
