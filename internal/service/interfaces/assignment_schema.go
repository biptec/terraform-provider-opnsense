package interfaces

import (
	"context"
	"fmt"
	"regexp"

	"github.com/biptec/opnsense-go/pkg/api"
	apiinterfaces "github.com/biptec/opnsense-go/pkg/interfaces"
	"github.com/biptec/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var assignmentAutomaticIdentifierPattern = regexp.MustCompile(`^opt[0-9]+$`)

type assignmentIdentifierValidator struct{}

func (assignmentIdentifierValidator) Description(_ context.Context) string {
	return "identifier must not use OPNsense-reserved lan, wan, or optN names"
}

func (v assignmentIdentifierValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v assignmentIdentifierValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() || request.ConfigValue.ValueString() == "" {
		return
	}
	value := request.ConfigValue.ValueString()
	if isReservedAssignmentIdentifier(value) {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(request.Path, v.Description(ctx), value))
	}
}

type assignmentIPv4Model struct {
	Mode         types.String `tfsdk:"mode"`
	Address      types.String `tfsdk:"address"`
	Prefix       types.Int64  `tfsdk:"prefix"`
	Gateway      types.String `tfsdk:"gateway"`
	DHCPHostname types.String `tfsdk:"dhcp_hostname"`
	AliasAddress types.String `tfsdk:"alias_address"`
	AliasPrefix  types.Int64  `tfsdk:"alias_prefix"`
	RejectFrom   types.String `tfsdk:"reject_from"`
}

type assignmentIPv6Model struct {
	Mode              types.String `tfsdk:"mode"`
	Address           types.String `tfsdk:"address"`
	Prefix            types.Int64  `tfsdk:"prefix"`
	Gateway           types.String `tfsdk:"gateway"`
	IAPDLength        types.Int64  `tfsdk:"ia_pd_length"`
	IAPDSendHint      types.Bool   `tfsdk:"ia_pd_send_hint"`
	PrefixOnly        types.Bool   `tfsdk:"prefix_only"`
	UseIPv4Interface  types.Bool   `tfsdk:"use_ipv4_interface"`
	VLANPriority      types.Int64  `tfsdk:"vlan_priority"`
	TrackInterface    types.String `tfsdk:"track_interface"`
	TrackPrefixID     types.Int64  `tfsdk:"track_prefix_id"`
	TrackAssociatedPD types.Int64  `tfsdk:"track_associated_pd"`
}

type assignmentResourceModel struct {
	Identifier       types.String         `tfsdk:"identifier"`
	Description      types.String         `tfsdk:"description"`
	Device           types.String         `tfsdk:"device"`
	Locked           types.Bool           `tfsdk:"locked"`
	Enabled          types.Bool           `tfsdk:"enabled"`
	BlockPrivate     types.Bool           `tfsdk:"block_private"`
	BlockBogons      types.Bool           `tfsdk:"block_bogons"`
	GatewayInterface types.Bool           `tfsdk:"gateway_interface"`
	Promiscuous      types.Bool           `tfsdk:"promiscuous"`
	SpoofMAC         types.String         `tfsdk:"spoof_mac"`
	MTU              types.Int64          `tfsdk:"mtu"`
	MSS              types.Int64          `tfsdk:"mss"`
	IPv4             *assignmentIPv4Model `tfsdk:"ipv4"`
	IPv6             *assignmentIPv6Model `tfsdk:"ipv6"`
	AllowReaddress   types.Bool           `tfsdk:"allow_readdress"`
	Name             types.String         `tfsdk:"name"`
	Id               types.String         `tfsdk:"id"`
}

type assignmentDataSourceModel struct {
	Identifier       types.String         `tfsdk:"identifier"`
	Description      types.String         `tfsdk:"description"`
	Device           types.String         `tfsdk:"device"`
	Locked           types.Bool           `tfsdk:"locked"`
	Enabled          types.Bool           `tfsdk:"enabled"`
	BlockPrivate     types.Bool           `tfsdk:"block_private"`
	BlockBogons      types.Bool           `tfsdk:"block_bogons"`
	GatewayInterface types.Bool           `tfsdk:"gateway_interface"`
	Promiscuous      types.Bool           `tfsdk:"promiscuous"`
	SpoofMAC         types.String         `tfsdk:"spoof_mac"`
	MTU              types.Int64          `tfsdk:"mtu"`
	MSS              types.Int64          `tfsdk:"mss"`
	IPv4             *assignmentIPv4Model `tfsdk:"ipv4"`
	IPv6             *assignmentIPv6Model `tfsdk:"ipv6"`
	Name             types.String         `tfsdk:"name"`
	Id               types.String         `tfsdk:"id"`
}

func assignmentResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Assigns a physical or virtual device to an OPNsense logical interface and manages its basic IPv4/IPv6 configuration.",
		Attributes: map[string]schema.Attribute{
			"identifier": schema.StringAttribute{
				Optional: true, Computed: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 32),
					stringvalidator.RegexMatches(regexp.MustCompile(`^[a-z][a-z0-9_]*$`), "Identifier must start with a lowercase letter and contain only lowercase letters, digits, and underscores."),
					assignmentIdentifierValidator{},
				},
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown(), stringplanmodifier.RequiresReplaceIfConfigured()},
				MarkdownDescription: "Stable OPNsense logical interface identifier. Leave unset to use automatic `optN` allocation. Legacy identifiers `lan`, `wan`, and `optN` are reserved by OPNsense.",
			},
			"description":       schema.StringAttribute{Optional: true, MarkdownDescription: "Interface description."},
			"device":            schema.StringAttribute{Required: true, MarkdownDescription: "Physical or virtual device, for example `vtnet1`, `vlan01`, or `vxlan0`."},
			"locked":            schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), MarkdownDescription: "Prevent deletion of the assignment in OPNsense."},
			"enabled":           schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true), MarkdownDescription: "Enable the interface."},
			"block_private":     schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), MarkdownDescription: "Block private networks on this interface."},
			"block_bogons":      schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), MarkdownDescription: "Block bogon networks on this interface."},
			"gateway_interface": schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), MarkdownDescription: "Mark the interface as a gateway interface."},
			"promiscuous":       schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), MarkdownDescription: "Enable promiscuous mode."},
			"spoof_mac":         schema.StringAttribute{Optional: true, MarkdownDescription: "Optional spoofed MAC address."},
			"mtu":               schema.Int64Attribute{Optional: true, Validators: []validator.Int64{int64validator.Between(576, 65535)}, MarkdownDescription: "Interface MTU."},
			"mss":               schema.Int64Attribute{Optional: true, Validators: []validator.Int64{int64validator.Between(576, 65535)}, MarkdownDescription: "TCP MSS adjustment."},
			"ipv4":              assignmentIPv4ResourceAttribute(),
			"ipv6":              assignmentIPv6ResourceAttribute(),
			"allow_readdress":   schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), MarkdownDescription: "Explicitly allow changes to the assigned device or address configuration. Keep disabled for the management interface unless an alternate management path is available."},
			"name":              schema.StringAttribute{Computed: true, MarkdownDescription: "Logical OPNsense interface name, for example `lan` or `opt1`."},
			"id":                schema.StringAttribute{Computed: true, MarkdownDescription: "Logical OPNsense interface name used as the import identifier.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		},
	}
}

func assignmentIPv4ResourceAttribute() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Required:            true,
		MarkdownDescription: "IPv4 address mode and mode-specific settings.",
		Attributes: map[string]schema.Attribute{
			"mode":          schema.StringAttribute{Required: true, Validators: []validator.String{stringvalidator.OneOf("none", "static", "dhcp")}, MarkdownDescription: "IPv4 mode: `none`, `static`, or `dhcp`."},
			"address":       schema.StringAttribute{Optional: true, MarkdownDescription: "Static IPv4 address."},
			"prefix":        schema.Int64Attribute{Optional: true, Validators: []validator.Int64{int64validator.Between(0, 32)}, MarkdownDescription: "Static IPv4 prefix length."},
			"gateway":       schema.StringAttribute{Optional: true, MarkdownDescription: "IPv4 gateway reference."},
			"dhcp_hostname": schema.StringAttribute{Optional: true, MarkdownDescription: "DHCP client hostname."},
			"alias_address": schema.StringAttribute{Optional: true, MarkdownDescription: "DHCP alias IPv4 address."},
			"alias_prefix":  schema.Int64Attribute{Optional: true, Validators: []validator.Int64{int64validator.Between(1, 32)}, MarkdownDescription: "DHCP alias prefix length."},
			"reject_from":   schema.StringAttribute{Optional: true, MarkdownDescription: "Comma-separated DHCP servers to reject."},
		},
	}
}

func assignmentIPv6ResourceAttribute() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Required:            true,
		MarkdownDescription: "IPv6 address mode and mode-specific settings.",
		Attributes: map[string]schema.Attribute{
			"mode":                schema.StringAttribute{Required: true, Validators: []validator.String{stringvalidator.OneOf("none", "static", "linklocal", "slaac", "dhcp6", "track6", "idassoc6")}, MarkdownDescription: "IPv6 mode."},
			"address":             schema.StringAttribute{Optional: true, MarkdownDescription: "Static IPv6 address."},
			"prefix":              schema.Int64Attribute{Optional: true, Validators: []validator.Int64{int64validator.Between(0, 128)}, MarkdownDescription: "Static IPv6 prefix length."},
			"gateway":             schema.StringAttribute{Optional: true, MarkdownDescription: "IPv6 gateway reference."},
			"ia_pd_length":        schema.Int64Attribute{Optional: true, Validators: []validator.Int64{int64validator.Between(48, 64)}, MarkdownDescription: "DHCPv6 IA-PD prefix length."},
			"ia_pd_send_hint":     schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), MarkdownDescription: "Send the configured IA-PD length as a hint."},
			"prefix_only":         schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), MarkdownDescription: "Request only a delegated prefix."},
			"use_ipv4_interface":  schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), MarkdownDescription: "Use the IPv4 interface for DHCPv6."},
			"vlan_priority":       schema.Int64Attribute{Optional: true, Validators: []validator.Int64{int64validator.Between(0, 7)}, MarkdownDescription: "DHCPv6 VLAN priority."},
			"track_interface":     schema.StringAttribute{Optional: true, MarkdownDescription: "Interface tracked by `track6` or `idassoc6`."},
			"track_prefix_id":     schema.Int64Attribute{Optional: true, Validators: []validator.Int64{int64validator.Between(0, 65535)}, MarkdownDescription: "Tracked prefix ID."},
			"track_associated_pd": schema.Int64Attribute{Optional: true, Validators: []validator.Int64{int64validator.AtLeast(0)}, MarkdownDescription: "Associated delegated-prefix index."},
		},
	}
}

func assignmentDataSourceSchema() dschema.Schema {
	return dschema.Schema{
		MarkdownDescription: "Reads an OPNsense logical interface assignment.",
		Attributes: map[string]dschema.Attribute{
			"id":                dschema.StringAttribute{Required: true, MarkdownDescription: "Logical interface name, for example `lan`, `opt1`, or a semantic identifier."},
			"identifier":        dschema.StringAttribute{Computed: true, MarkdownDescription: "Stable logical interface identifier."},
			"name":              dschema.StringAttribute{Computed: true},
			"description":       dschema.StringAttribute{Computed: true},
			"device":            dschema.StringAttribute{Computed: true},
			"locked":            dschema.BoolAttribute{Computed: true},
			"enabled":           dschema.BoolAttribute{Computed: true},
			"block_private":     dschema.BoolAttribute{Computed: true},
			"block_bogons":      dschema.BoolAttribute{Computed: true},
			"gateway_interface": dschema.BoolAttribute{Computed: true},
			"promiscuous":       dschema.BoolAttribute{Computed: true},
			"spoof_mac":         dschema.StringAttribute{Computed: true},
			"mtu":               dschema.Int64Attribute{Computed: true},
			"mss":               dschema.Int64Attribute{Computed: true},
			"ipv4":              assignmentIPv4DataSourceAttribute(),
			"ipv6":              assignmentIPv6DataSourceAttribute(),
		},
	}
}

func assignmentIPv4DataSourceAttribute() dschema.SingleNestedAttribute {
	return dschema.SingleNestedAttribute{Computed: true, Attributes: map[string]dschema.Attribute{
		"mode": dschema.StringAttribute{Computed: true}, "address": dschema.StringAttribute{Computed: true},
		"prefix": dschema.Int64Attribute{Computed: true}, "gateway": dschema.StringAttribute{Computed: true},
		"dhcp_hostname": dschema.StringAttribute{Computed: true}, "alias_address": dschema.StringAttribute{Computed: true},
		"alias_prefix": dschema.Int64Attribute{Computed: true}, "reject_from": dschema.StringAttribute{Computed: true},
	}}
}

func assignmentIPv6DataSourceAttribute() dschema.SingleNestedAttribute {
	return dschema.SingleNestedAttribute{Computed: true, Attributes: map[string]dschema.Attribute{
		"mode": dschema.StringAttribute{Computed: true}, "address": dschema.StringAttribute{Computed: true},
		"prefix": dschema.Int64Attribute{Computed: true}, "gateway": dschema.StringAttribute{Computed: true},
		"ia_pd_length": dschema.Int64Attribute{Computed: true}, "ia_pd_send_hint": dschema.BoolAttribute{Computed: true},
		"prefix_only": dschema.BoolAttribute{Computed: true}, "use_ipv4_interface": dschema.BoolAttribute{Computed: true},
		"vlan_priority": dschema.Int64Attribute{Computed: true}, "track_interface": dschema.StringAttribute{Computed: true},
		"track_prefix_id": dschema.Int64Attribute{Computed: true}, "track_associated_pd": dschema.Int64Attribute{Computed: true},
	}}
}

func convertAssignmentSchemaToStruct(d *assignmentResourceModel) (*apiinterfaces.Assignment, error) {
	if d.IPv4 == nil || d.IPv6 == nil {
		return nil, fmt.Errorf("both ipv4 and ipv6 blocks are required")
	}
	if err := validateAssignmentAddressMode(d.IPv4.Mode.ValueString(), d.IPv4.Address, d.IPv4.Prefix, 4); err != nil {
		return nil, err
	}
	if err := validateAssignmentAddressMode(d.IPv6.Mode.ValueString(), d.IPv6.Address, d.IPv6.Prefix, 6); err != nil {
		return nil, err
	}
	return &apiinterfaces.Assignment{
		Identifier: d.Identifier.ValueString(), Description: d.Description.ValueString(), Device: api.SelectedMap(d.Device.ValueString()),
		Lock: tools.BoolToString(d.Locked.ValueBool()), Enabled: tools.BoolToString(d.Enabled.ValueBool()),
		BlockPrivate: tools.BoolToString(d.BlockPrivate.ValueBool()), BlockBogons: tools.BoolToString(d.BlockBogons.ValueBool()),
		GatewayInterface: tools.BoolToString(d.GatewayInterface.ValueBool()), Promiscuous: tools.BoolToString(d.Promiscuous.ValueBool()),
		SpoofMAC: d.SpoofMAC.ValueString(), MTU: optionalIntString(d.MTU), MSS: optionalIntString(d.MSS),
		IPv4Mode: api.SelectedMap(d.IPv4.Mode.ValueString()), IPv4Address: d.IPv4.Address.ValueString(),
		IPv4Prefix: optionalIntString(d.IPv4.Prefix), IPv4Gateway: d.IPv4.Gateway.ValueString(),
		DHCPHostname: d.IPv4.DHCPHostname.ValueString(), DHCPAliasAddress: d.IPv4.AliasAddress.ValueString(),
		DHCPAliasPrefix: optionalIntString(d.IPv4.AliasPrefix), DHCPRejectFrom: d.IPv4.RejectFrom.ValueString(),
		IPv6Mode: api.SelectedMap(d.IPv6.Mode.ValueString()), IPv6Address: d.IPv6.Address.ValueString(),
		IPv6Prefix: optionalIntString(d.IPv6.Prefix), IPv6Gateway: d.IPv6.Gateway.ValueString(),
		DHCP6IAPDLength: optionalIntString(d.IPv6.IAPDLength), DHCP6IAPDSendHint: tools.BoolToString(d.IPv6.IAPDSendHint.ValueBool()),
		DHCP6PrefixOnly: tools.BoolToString(d.IPv6.PrefixOnly.ValueBool()), DHCP6UseIPv4Interface: tools.BoolToString(d.IPv6.UseIPv4Interface.ValueBool()),
		DHCP6VLANPriority: optionalIntString(d.IPv6.VLANPriority), Track6Interface: d.IPv6.TrackInterface.ValueString(),
		Track6PrefixID: optionalIntString(d.IPv6.TrackPrefixID), Track6AssociatedPD: optionalIntString(d.IPv6.TrackAssociatedPD),
	}, nil
}

func isReservedAssignmentIdentifier(value string) bool {
	return value == "lan" || value == "wan" || assignmentAutomaticIdentifierPattern.MatchString(value)
}

func validateAssignmentAddressMode(mode string, address types.String, prefix types.Int64, family int) error {
	if mode == "static" {
		if address.IsNull() || address.IsUnknown() || address.ValueString() == "" || prefix.IsNull() || prefix.IsUnknown() {
			return fmt.Errorf("ipv%d static mode requires address and prefix", family)
		}
		return nil
	}
	if (!address.IsNull() && !address.IsUnknown() && address.ValueString() != "") || (!prefix.IsNull() && !prefix.IsUnknown()) {
		return fmt.Errorf("ipv%d address and prefix may only be set in static mode", family)
	}
	return nil
}

func optionalIntString(value types.Int64) string {
	if value.IsNull() || value.IsUnknown() {
		return ""
	}
	return tools.Int64ToString(value.ValueInt64())
}

func assignmentIPv4FromStruct(d *apiinterfaces.Assignment) *assignmentIPv4Model {
	return &assignmentIPv4Model{Mode: types.StringValue(d.IPv4Mode.String()), Address: tools.StringOrNull(d.IPv4Address), Prefix: tools.StringToInt64Null(d.IPv4Prefix), Gateway: tools.StringOrNull(d.IPv4Gateway), DHCPHostname: tools.StringOrNull(d.DHCPHostname), AliasAddress: tools.StringOrNull(d.DHCPAliasAddress), AliasPrefix: tools.StringToInt64Null(d.DHCPAliasPrefix), RejectFrom: tools.StringOrNull(d.DHCPRejectFrom)}
}

func assignmentIPv6FromStruct(d *apiinterfaces.Assignment) *assignmentIPv6Model {
	return &assignmentIPv6Model{Mode: types.StringValue(d.IPv6Mode.String()), Address: tools.StringOrNull(d.IPv6Address), Prefix: tools.StringToInt64Null(d.IPv6Prefix), Gateway: tools.StringOrNull(d.IPv6Gateway), IAPDLength: tools.StringToInt64Null(d.DHCP6IAPDLength), IAPDSendHint: types.BoolValue(tools.StringToBool(d.DHCP6IAPDSendHint)), PrefixOnly: types.BoolValue(tools.StringToBool(d.DHCP6PrefixOnly)), UseIPv4Interface: types.BoolValue(tools.StringToBool(d.DHCP6UseIPv4Interface)), VLANPriority: tools.StringToInt64Null(d.DHCP6VLANPriority), TrackInterface: tools.StringOrNull(d.Track6Interface), TrackPrefixID: tools.StringToInt64Null(d.Track6PrefixID), TrackAssociatedPD: tools.StringToInt64Null(d.Track6AssociatedPD)}
}

func convertAssignmentStructToResourceSchema(d *apiinterfaces.Assignment, id string, allowReaddress types.Bool) *assignmentResourceModel {
	name := d.Identifier
	if name == "" {
		name = id
	}
	if allowReaddress.IsNull() || allowReaddress.IsUnknown() {
		allowReaddress = types.BoolValue(false)
	}
	return &assignmentResourceModel{Identifier: types.StringValue(name), Description: tools.StringOrNull(d.Description), Device: types.StringValue(d.Device.String()), Locked: types.BoolValue(tools.StringToBool(d.Lock)), Enabled: types.BoolValue(tools.StringToBool(d.Enabled)), BlockPrivate: types.BoolValue(tools.StringToBool(d.BlockPrivate)), BlockBogons: types.BoolValue(tools.StringToBool(d.BlockBogons)), GatewayInterface: types.BoolValue(tools.StringToBool(d.GatewayInterface)), Promiscuous: types.BoolValue(tools.StringToBool(d.Promiscuous)), SpoofMAC: tools.StringOrNull(d.SpoofMAC), MTU: tools.StringToInt64Null(d.MTU), MSS: tools.StringToInt64Null(d.MSS), IPv4: assignmentIPv4FromStruct(d), IPv6: assignmentIPv6FromStruct(d), AllowReaddress: allowReaddress, Name: types.StringValue(name), Id: types.StringValue(id)}
}

func convertAssignmentStructToDataSourceSchema(d *apiinterfaces.Assignment, id string) *assignmentDataSourceModel {
	name := d.Identifier
	if name == "" {
		name = id
	}
	return &assignmentDataSourceModel{Identifier: types.StringValue(name), Description: tools.StringOrNull(d.Description), Device: types.StringValue(d.Device.String()), Locked: types.BoolValue(tools.StringToBool(d.Lock)), Enabled: types.BoolValue(tools.StringToBool(d.Enabled)), BlockPrivate: types.BoolValue(tools.StringToBool(d.BlockPrivate)), BlockBogons: types.BoolValue(tools.StringToBool(d.BlockBogons)), GatewayInterface: types.BoolValue(tools.StringToBool(d.GatewayInterface)), Promiscuous: types.BoolValue(tools.StringToBool(d.Promiscuous)), SpoofMAC: tools.StringOrNull(d.SpoofMAC), MTU: tools.StringToInt64Null(d.MTU), MSS: tools.StringToInt64Null(d.MSS), IPv4: assignmentIPv4FromStruct(d), IPv6: assignmentIPv6FromStruct(d), Name: types.StringValue(name), Id: types.StringValue(id)}
}
