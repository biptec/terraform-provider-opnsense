package interfaces

import (
	"github.com/biptec/opnsense-go/pkg/api"
	apiinterfaces "github.com/biptec/opnsense-go/pkg/interfaces"
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

type interfaceSettingsResourceModel struct {
	DisableChecksumOffloading     types.Bool   `tfsdk:"disable_checksum_offloading"`
	DisableSegmentationOffloading types.Bool   `tfsdk:"disable_segmentation_offloading"`
	DisableLargeReceiveOffloading types.Bool   `tfsdk:"disable_large_receive_offloading"`
	VLANHardwareFiltering         types.String `tfsdk:"vlan_hardware_filtering"`
	DisableIPv6                   types.Bool   `tfsdk:"disable_ipv6"`
	DHCP6NoRelease                types.Bool   `tfsdk:"dhcp6_no_release"`
	DHCP6Debug                    types.Bool   `tfsdk:"dhcp6_debug"`
	DHCP6DUID                     types.String `tfsdk:"dhcp6_duid"`
	DHCP6RATimeout                types.Int64  `tfsdk:"dhcp6_ra_timeout"`
	Id                            types.String `tfsdk:"id"`
}

func interfaceSettingsResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manages global OPNsense interface settings. This is a singleton resource.",
		Attributes: map[string]schema.Attribute{
			"disable_checksum_offloading":      schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true), MarkdownDescription: "Disable hardware checksum offloading."},
			"disable_segmentation_offloading":  schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true), MarkdownDescription: "Disable hardware TCP segmentation offloading."},
			"disable_large_receive_offloading": schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true), MarkdownDescription: "Disable hardware large receive offloading."},
			"vlan_hardware_filtering":          schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("2"), Validators: []validator.String{stringvalidator.OneOf("0", "1", "2")}, MarkdownDescription: "VLAN hardware filtering mode: `0`, `1`, or `2` (default)."},
			"disable_ipv6":                     schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), MarkdownDescription: "Disable IPv6 globally."},
			"dhcp6_no_release":                 schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), MarkdownDescription: "Do not send DHCPv6 release messages."},
			"dhcp6_debug":                      schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), MarkdownDescription: "Enable DHCPv6 client debug logging."},
			"dhcp6_duid":                       schema.StringAttribute{Optional: true, MarkdownDescription: "Custom DHCPv6 DUID. Leave unset to use the current/default DUID."},
			"dhcp6_ra_timeout":                 schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(10), Validators: []validator.Int64{int64validator.AtLeast(0)}, MarkdownDescription: "Router-advertisement timeout in seconds."},
			"id":                               schema.StringAttribute{Computed: true, MarkdownDescription: "Fixed singleton identifier `interfaces_settings`.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		},
	}
}

func interfaceSettingsDataSourceSchema() dschema.Schema {
	return dschema.Schema{
		MarkdownDescription: "Reads global OPNsense interface settings.",
		Attributes: map[string]dschema.Attribute{
			"disable_checksum_offloading":      dschema.BoolAttribute{Computed: true},
			"disable_segmentation_offloading":  dschema.BoolAttribute{Computed: true},
			"disable_large_receive_offloading": dschema.BoolAttribute{Computed: true},
			"vlan_hardware_filtering":          dschema.StringAttribute{Computed: true},
			"disable_ipv6":                     dschema.BoolAttribute{Computed: true},
			"dhcp6_no_release":                 dschema.BoolAttribute{Computed: true},
			"dhcp6_debug":                      dschema.BoolAttribute{Computed: true},
			"dhcp6_duid":                       dschema.StringAttribute{Computed: true},
			"dhcp6_ra_timeout":                 dschema.Int64Attribute{Computed: true},
			"id":                               dschema.StringAttribute{Computed: true},
		},
	}
}

func convertInterfaceSettingsSchemaToStruct(d *interfaceSettingsResourceModel) *apiinterfaces.InterfaceSettings {
	return &apiinterfaces.InterfaceSettings{
		DisableChecksumOffloading:     tools.BoolToString(d.DisableChecksumOffloading.ValueBool()),
		DisableSegmentationOffloading: tools.BoolToString(d.DisableSegmentationOffloading.ValueBool()),
		DisableLargeReceiveOffloading: tools.BoolToString(d.DisableLargeReceiveOffloading.ValueBool()),
		VLANHardwareFiltering:         api.SelectedMap(d.VLANHardwareFiltering.ValueString()),
		DisableIPv6:                   tools.BoolToString(d.DisableIPv6.ValueBool()),
		DHCP6NoRelease:                tools.BoolToString(d.DHCP6NoRelease.ValueBool()),
		DHCP6Debug:                    tools.BoolToString(d.DHCP6Debug.ValueBool()),
		DHCP6DUID:                     d.DHCP6DUID.ValueString(),
		DHCP6RATimeout:                tools.Int64ToString(d.DHCP6RATimeout.ValueInt64()),
	}
}

func convertInterfaceSettingsStructToSchema(d *apiinterfaces.InterfaceSettings) *interfaceSettingsResourceModel {
	raTimeout := tools.StringToInt64(d.DHCP6RATimeout)
	if raTimeout < 0 {
		raTimeout = 10
	}
	vlanHardwareFiltering := d.VLANHardwareFiltering.String()
	if vlanHardwareFiltering == "" {
		vlanHardwareFiltering = "2"
	}
	return &interfaceSettingsResourceModel{
		DisableChecksumOffloading:     types.BoolValue(tools.StringToBool(d.DisableChecksumOffloading)),
		DisableSegmentationOffloading: types.BoolValue(tools.StringToBool(d.DisableSegmentationOffloading)),
		DisableLargeReceiveOffloading: types.BoolValue(tools.StringToBool(d.DisableLargeReceiveOffloading)),
		VLANHardwareFiltering:         types.StringValue(vlanHardwareFiltering),
		DisableIPv6:                   types.BoolValue(tools.StringToBool(d.DisableIPv6)),
		DHCP6NoRelease:                types.BoolValue(tools.StringToBool(d.DHCP6NoRelease)),
		DHCP6Debug:                    types.BoolValue(tools.StringToBool(d.DHCP6Debug)),
		DHCP6DUID:                     tools.StringOrNull(d.DHCP6DUID),
		DHCP6RATimeout:                types.Int64Value(raTimeout),
		Id:                            types.StringValue("interfaces_settings"),
	}
}
