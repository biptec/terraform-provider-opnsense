package interfaces

import (
	"fmt"
	"net"
	"net/netip"

	apiinterfaces "github.com/biptec/opnsense-go/pkg/interfaces"
	"github.com/biptec/terraform-provider-opnsense/internal/tools"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type neighborResourceModel struct {
	MACAddress  types.String `tfsdk:"mac_address"`
	IPAddress   types.String `tfsdk:"ip_address"`
	Description types.String `tfsdk:"description"`
	Origin      types.String `tfsdk:"origin"`
	Id          types.String `tfsdk:"id"`
}

func neighborResourceSchema() schema.Schema {
	return schema.Schema{MarkdownDescription: "Manages a static OPNsense ARP or NDP neighbour entry.", Attributes: map[string]schema.Attribute{
		"mac_address": schema.StringAttribute{Required: true, MarkdownDescription: "Ethernet MAC address."},
		"ip_address":  schema.StringAttribute{Required: true, MarkdownDescription: "IPv4 or IPv6 address."},
		"description": schema.StringAttribute{Optional: true, MarkdownDescription: "Optional description."},
		"origin":      schema.StringAttribute{Computed: true, MarkdownDescription: "Origin reported by OPNsense."},
		"id":          schema.StringAttribute{Computed: true, MarkdownDescription: "UUID of the neighbour entry.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
	}}
}
func neighborDataSourceSchema() dschema.Schema {
	return dschema.Schema{MarkdownDescription: "Reads a static OPNsense ARP or NDP neighbour entry.", Attributes: map[string]dschema.Attribute{
		"id": dschema.StringAttribute{Required: true}, "mac_address": dschema.StringAttribute{Computed: true},
		"ip_address": dschema.StringAttribute{Computed: true}, "description": dschema.StringAttribute{Computed: true}, "origin": dschema.StringAttribute{Computed: true},
	}}
}
func convertNeighborSchemaToStruct(d *neighborResourceModel) (*apiinterfaces.Neighbor, error) {
	mac, err := net.ParseMAC(d.MACAddress.ValueString())
	if err != nil {
		return nil, fmt.Errorf("invalid mac_address: %w", err)
	}
	if len(mac) != 6 {
		return nil, fmt.Errorf("mac_address must be a 6-byte Ethernet address")
	}
	if _, err := netip.ParseAddr(d.IPAddress.ValueString()); err != nil {
		return nil, fmt.Errorf("invalid ip_address: %w", err)
	}
	return &apiinterfaces.Neighbor{MACAddress: mac.String(), IPAddress: d.IPAddress.ValueString(), Description: d.Description.ValueString()}, nil
}
func convertNeighborStructToSchema(d *apiinterfaces.Neighbor) (*neighborResourceModel, error) {
	return &neighborResourceModel{MACAddress: types.StringValue(d.MACAddress), IPAddress: types.StringValue(d.IPAddress), Description: tools.StringOrNull(d.Description), Origin: tools.StringOrNull(d.Origin)}, nil
}
