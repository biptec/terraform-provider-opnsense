package interfaces

import (
	"fmt"

	"github.com/biptec/opnsense-go/pkg/api"
	apiinterfaces "github.com/biptec/opnsense-go/pkg/interfaces"
	"github.com/biptec/terraform-provider-opnsense/internal/tools"
	"github.com/biptec/terraform-provider-opnsense/internal/validators"
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

type vipResourceModel struct {
	Mode              types.String `tfsdk:"mode"`
	Interface         types.String `tfsdk:"interface"`
	Network           types.String `tfsdk:"network"`
	Gateway           types.String `tfsdk:"gateway"`
	NoExpand          types.Bool   `tfsdk:"no_expand"`
	NoBind            types.Bool   `tfsdk:"no_bind"`
	Password          types.String `tfsdk:"password"`
	VHID              types.Int64  `tfsdk:"vhid"`
	AdvertisementBase types.Int64  `tfsdk:"advertisement_base"`
	AdvertisementSkew types.Int64  `tfsdk:"advertisement_skew"`
	PeerIPv4          types.String `tfsdk:"peer_ipv4"`
	PeerIPv6          types.String `tfsdk:"peer_ipv6"`
	NoSync            types.Bool   `tfsdk:"no_sync"`
	Address           types.String `tfsdk:"address"`
	VHIDText          types.String `tfsdk:"vhid_text"`
	Description       types.String `tfsdk:"description"`
	Id                types.String `tfsdk:"id"`
}

func vipResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manages an OPNsense virtual IP, including IP Alias, CARP, and Proxy ARP modes.",
		Attributes: map[string]schema.Attribute{
			"mode": schema.StringAttribute{
				Optional: true, Computed: true, Default: stringdefault.StaticString("ipalias"),
				Validators:          []validator.String{stringvalidator.OneOf("ipalias", "carp", "proxyarp")},
				MarkdownDescription: "Virtual IP mode: `ipalias`, `carp`, or `proxyarp`.",
			},
			"interface": schema.StringAttribute{
				Optional: true, Computed: true, Default: stringdefault.StaticString("wan"),
				MarkdownDescription: "Logical interface to which the virtual IP applies.",
			},
			"network": schema.StringAttribute{
				Required: true, Validators: []validator.String{validators.CIDR()},
				MarkdownDescription: "Virtual IP address and prefix, for example `192.0.2.10/24`.",
			},
			"gateway": schema.StringAttribute{
				Optional: true, Validators: []validator.String{validators.IpOrCIDR()},
				MarkdownDescription: "Gateway used by IP Alias on PPP, PPPoE, or tunnel interfaces.",
			},
			"no_expand":          schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), MarkdownDescription: "Do not expand the virtual IP into automatic firewall rules."},
			"no_bind":            schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), MarkdownDescription: "Do not bind services to the virtual IP."},
			"password":           schema.StringAttribute{Optional: true, Sensitive: true, MarkdownDescription: "CARP password."},
			"vhid":               schema.Int64Attribute{Optional: true, Validators: []validator.Int64{int64validator.Between(1, 255)}, MarkdownDescription: "CARP VHID."},
			"advertisement_base": schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(1), Validators: []validator.Int64{int64validator.AtLeast(1)}, MarkdownDescription: "CARP advertisement base interval."},
			"advertisement_skew": schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(0), Validators: []validator.Int64{int64validator.Between(0, 254)}, MarkdownDescription: "CARP advertisement skew."},
			"peer_ipv4":          schema.StringAttribute{Optional: true, Validators: []validator.String{validators.IpOrCIDR()}, MarkdownDescription: "Optional CARP IPv4 peer."},
			"peer_ipv6":          schema.StringAttribute{Optional: true, Validators: []validator.String{validators.IpOrCIDR()}, MarkdownDescription: "Optional CARP IPv6 peer."},
			"no_sync":            schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), MarkdownDescription: "Exclude this virtual IP from XMLRPC synchronization."},
			"address":            schema.StringAttribute{Computed: true, MarkdownDescription: "Address rendered by OPNsense."},
			"vhid_text":          schema.StringAttribute{Computed: true, MarkdownDescription: "VHID label rendered by OPNsense."},
			"description":        schema.StringAttribute{Optional: true, MarkdownDescription: "Optional description."},
			"id":                 schema.StringAttribute{Computed: true, MarkdownDescription: "UUID of the virtual IP.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		},
	}
}

func vipDataSourceSchema() dschema.Schema {
	return dschema.Schema{
		MarkdownDescription: "Reads an OPNsense virtual IP.",
		Attributes: map[string]dschema.Attribute{
			"id":                 dschema.StringAttribute{Required: true, MarkdownDescription: "UUID of the virtual IP."},
			"mode":               dschema.StringAttribute{Computed: true},
			"interface":          dschema.StringAttribute{Computed: true},
			"network":            dschema.StringAttribute{Computed: true},
			"gateway":            dschema.StringAttribute{Computed: true},
			"no_expand":          dschema.BoolAttribute{Computed: true},
			"no_bind":            dschema.BoolAttribute{Computed: true},
			"password":           dschema.StringAttribute{Computed: true, Sensitive: true},
			"vhid":               dschema.Int64Attribute{Computed: true},
			"advertisement_base": dschema.Int64Attribute{Computed: true},
			"advertisement_skew": dschema.Int64Attribute{Computed: true},
			"peer_ipv4":          dschema.StringAttribute{Computed: true},
			"peer_ipv6":          dschema.StringAttribute{Computed: true},
			"no_sync":            dschema.BoolAttribute{Computed: true},
			"address":            dschema.StringAttribute{Computed: true},
			"vhid_text":          dschema.StringAttribute{Computed: true},
			"description":        dschema.StringAttribute{Computed: true},
		},
	}
}

func convertVipSchemaToStruct(d *vipResourceModel) (*apiinterfaces.Vip, error) {
	mode := d.Mode.ValueString()
	if mode == "carp" {
		if d.Password.IsNull() || d.Password.ValueString() == "" || d.VHID.IsNull() {
			return nil, fmt.Errorf("carp mode requires password and vhid")
		}
	} else if (!d.Password.IsNull() && d.Password.ValueString() != "") || !d.VHID.IsNull() {
		return nil, fmt.Errorf("password and vhid may only be set in carp mode")
	}
	return &apiinterfaces.Vip{
		Mode: api.SelectedMap(mode), Interface: api.SelectedMap(d.Interface.ValueString()),
		Network: d.Network.ValueString(), Gateway: d.Gateway.ValueString(),
		NoExpand: tools.BoolToString(d.NoExpand.ValueBool()), NoBind: tools.BoolToString(d.NoBind.ValueBool()),
		Password: d.Password.ValueString(), VHID: optionalIntString(d.VHID),
		AdvertisementBase: tools.Int64ToString(d.AdvertisementBase.ValueInt64()),
		AdvertisementSkew: tools.Int64ToString(d.AdvertisementSkew.ValueInt64()),
		PeerIPv4:          d.PeerIPv4.ValueString(), PeerIPv6: d.PeerIPv6.ValueString(),
		NoSync: tools.BoolToString(d.NoSync.ValueBool()), Description: d.Description.ValueString(),
	}, nil
}

func convertVipStructToSchema(d *apiinterfaces.Vip) (*vipResourceModel, error) {
	advertisementBase := tools.StringToInt64(d.AdvertisementBase)
	if advertisementBase < 1 {
		advertisementBase = 1
	}
	advertisementSkew := tools.StringToInt64(d.AdvertisementSkew)
	if advertisementSkew < 0 {
		advertisementSkew = 0
	}
	return &vipResourceModel{
		Mode: types.StringValue(d.Mode.String()), Interface: types.StringValue(d.Interface.String()),
		Network: types.StringValue(d.Network), Gateway: tools.StringOrNull(d.Gateway),
		NoExpand: types.BoolValue(tools.StringToBool(d.NoExpand)), NoBind: types.BoolValue(tools.StringToBool(d.NoBind)),
		Password: tools.StringOrNull(d.Password), VHID: tools.StringToInt64Null(d.VHID),
		AdvertisementBase: types.Int64Value(advertisementBase), AdvertisementSkew: types.Int64Value(advertisementSkew),
		PeerIPv4: tools.StringOrNull(d.PeerIPv4), PeerIPv6: tools.StringOrNull(d.PeerIPv6),
		NoSync: types.BoolValue(tools.StringToBool(d.NoSync)), Address: tools.StringOrNull(d.Address),
		VHIDText: tools.StringOrNull(d.VHIDText), Description: tools.StringOrNull(d.Description),
	}, nil
}
