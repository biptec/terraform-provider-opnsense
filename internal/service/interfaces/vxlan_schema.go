package interfaces

import (
	"fmt"
	"net/netip"

	"github.com/biptec/opnsense-go/pkg/api"
	apiinterfaces "github.com/biptec/opnsense-go/pkg/interfaces"
	"github.com/biptec/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type vxlanResourceModel struct {
	DeviceID           types.Int64  `tfsdk:"device_id"`
	VNI                types.Int64  `tfsdk:"vni"`
	SourceAddress      types.String `tfsdk:"source_address"`
	SourcePort         types.Int64  `tfsdk:"source_port"`
	RemoteAddress      types.String `tfsdk:"remote_address"`
	RemotePort         types.Int64  `tfsdk:"remote_port"`
	MulticastGroup     types.String `tfsdk:"multicast_group"`
	MulticastInterface types.String `tfsdk:"multicast_interface"`
	Id                 types.String `tfsdk:"id"`
}

func vxlanResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manages an OPNsense VXLAN interface in unicast or multicast mode.",
		Attributes: map[string]schema.Attribute{
			"device_id":           schema.Int64Attribute{Computed: true, MarkdownDescription: "Automatically allocated VXLAN device number."},
			"vni":                 schema.Int64Attribute{Required: true, Validators: []validator.Int64{int64validator.Between(0, 16777215)}, MarkdownDescription: "VXLAN Network Identifier."},
			"source_address":      schema.StringAttribute{Required: true, MarkdownDescription: "Local source IPv4 or IPv6 address used for encapsulation."},
			"source_port":         schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(4789), Validators: []validator.Int64{int64validator.Between(1, 65535)}, MarkdownDescription: "Local UDP source port."},
			"remote_address":      schema.StringAttribute{Optional: true, MarkdownDescription: "Remote tunnel endpoint for unicast mode. Mutually exclusive with multicast_group."},
			"remote_port":         schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(4789), Validators: []validator.Int64{int64validator.Between(1, 65535)}, MarkdownDescription: "Remote UDP destination port."},
			"multicast_group":     schema.StringAttribute{Optional: true, MarkdownDescription: "Multicast group for multicast mode. Mutually exclusive with remote_address."},
			"multicast_interface": schema.StringAttribute{Optional: true, MarkdownDescription: "Physical interface used to transmit multicast packets. Required with multicast_group and forbidden with remote_address."},
			"id":                  schema.StringAttribute{Computed: true, MarkdownDescription: "UUID of the VXLAN configuration.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		},
	}
}

func vxlanDataSourceSchema() dschema.Schema {
	return dschema.Schema{
		MarkdownDescription: "Reads an OPNsense VXLAN interface.",
		Attributes: map[string]dschema.Attribute{
			"id":                  dschema.StringAttribute{Required: true, MarkdownDescription: "UUID of the VXLAN configuration."},
			"device_id":           dschema.Int64Attribute{Computed: true},
			"vni":                 dschema.Int64Attribute{Computed: true},
			"source_address":      dschema.StringAttribute{Computed: true},
			"source_port":         dschema.Int64Attribute{Computed: true},
			"remote_address":      dschema.StringAttribute{Computed: true},
			"remote_port":         dschema.Int64Attribute{Computed: true},
			"multicast_group":     dschema.StringAttribute{Computed: true},
			"multicast_interface": dschema.StringAttribute{Computed: true},
		},
	}
}

func convertVxlanSchemaToStruct(d *vxlanResourceModel) (*apiinterfaces.Vxlan, error) {
	local, err := netip.ParseAddr(d.SourceAddress.ValueString())
	if err != nil {
		return nil, fmt.Errorf("invalid source_address: %w", err)
	}
	remoteSet := !d.RemoteAddress.IsNull() && !d.RemoteAddress.IsUnknown() && d.RemoteAddress.ValueString() != ""
	groupSet := !d.MulticastGroup.IsNull() && !d.MulticastGroup.IsUnknown() && d.MulticastGroup.ValueString() != ""
	deviceSet := !d.MulticastInterface.IsNull() && !d.MulticastInterface.IsUnknown() && d.MulticastInterface.ValueString() != ""
	if remoteSet == groupSet {
		return nil, fmt.Errorf("exactly one of remote_address or multicast_group must be configured")
	}
	if remoteSet && deviceSet {
		return nil, fmt.Errorf("multicast_interface must not be configured with remote_address")
	}
	if groupSet && !deviceSet {
		return nil, fmt.Errorf("multicast_interface is required with multicast_group")
	}
	if remoteSet {
		remote, parseErr := netip.ParseAddr(d.RemoteAddress.ValueString())
		if parseErr != nil {
			return nil, fmt.Errorf("invalid remote_address: %w", parseErr)
		}
		if local.BitLen() != remote.BitLen() {
			return nil, fmt.Errorf("source_address and remote_address must use the same address family")
		}
	}
	if groupSet {
		group, parseErr := netip.ParseAddr(d.MulticastGroup.ValueString())
		if parseErr != nil {
			return nil, fmt.Errorf("invalid multicast_group: %w", parseErr)
		}
		if !group.IsMulticast() {
			return nil, fmt.Errorf("multicast_group must be a multicast address")
		}
		if local.BitLen() != group.BitLen() {
			return nil, fmt.Errorf("source_address and multicast_group must use the same address family")
		}
	}
	return &apiinterfaces.Vxlan{
		VNI: tools.Int64ToString(d.VNI.ValueInt64()), LocalAddress: d.SourceAddress.ValueString(),
		LocalPort: tools.Int64ToString(d.SourcePort.ValueInt64()), RemoteAddress: d.RemoteAddress.ValueString(),
		RemotePort: tools.Int64ToString(d.RemotePort.ValueInt64()), MulticastGroup: d.MulticastGroup.ValueString(),
		PhysicalInterface: api.SelectedMap(d.MulticastInterface.ValueString()),
	}, nil
}

func convertVxlanStructToSchema(d *apiinterfaces.Vxlan) (*vxlanResourceModel, error) {
	sourcePort := tools.StringToInt64(d.LocalPort)
	if sourcePort < 1 {
		sourcePort = 4789
	}
	remotePort := tools.StringToInt64(d.RemotePort)
	if remotePort < 1 {
		remotePort = 4789
	}
	return &vxlanResourceModel{
		DeviceID: tools.StringToInt64Null(d.DeviceID), VNI: tools.StringToInt64Null(d.VNI),
		SourceAddress: types.StringValue(d.LocalAddress), SourcePort: types.Int64Value(sourcePort),
		RemoteAddress: tools.StringOrNull(d.RemoteAddress), RemotePort: types.Int64Value(remotePort),
		MulticastGroup: tools.StringOrNull(d.MulticastGroup), MulticastInterface: tools.StringOrNull(d.PhysicalInterface.String()),
	}, nil
}
