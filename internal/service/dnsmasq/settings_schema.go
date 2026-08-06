package dnsmasq

import (
	"github.com/biptec/opnsense-go/pkg/api"
	apidnsmasq "github.com/biptec/opnsense-go/pkg/dnsmasq"
	"github.com/biptec/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type settingsResourceModel struct {
	ID                     types.String `tfsdk:"id"`
	Enabled                types.Bool   `tfsdk:"enabled"`
	Interfaces             types.Set    `tfsdk:"interfaces"`
	StrictInterfaceBinding types.Bool   `tfsdk:"strict_interface_binding"`
	DNSPort                types.Int64  `tfsdk:"dns_port"`
}

func settingsResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manages dnsmasq general listener settings. Import the singleton before use. Omitted settings retain their current OPNsense values.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Always `dnsmasq_settings`.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"enabled": schema.BoolAttribute{
				Optional: true, Computed: true,
				MarkdownDescription: "Enable the dnsmasq service.",
			},
			"interfaces": schema.SetAttribute{
				Optional: true, Computed: true, ElementType: types.StringType,
				MarkdownDescription: "Logical interfaces on which dnsmasq is allowed to listen.",
			},
			"strict_interface_binding": schema.BoolAttribute{
				Optional: true, Computed: true,
				MarkdownDescription: "Bind only to the selected interfaces instead of wildcard sockets.",
			},
			"dns_port": schema.Int64Attribute{
				Optional: true, Computed: true,
				Validators:          []validator.Int64{int64validator.Between(0, 65535)},
				MarkdownDescription: "DNS listener port. Set to `0` to disable only the DNS function while keeping dnsmasq available for DHCP or router advertisements.",
			},
		},
	}
}

func settingsDataSourceSchema() dschema.Schema {
	return dschema.Schema{
		MarkdownDescription: "Reads dnsmasq general listener settings.",
		Attributes: map[string]dschema.Attribute{
			"id":                       dschema.StringAttribute{Computed: true},
			"enabled":                  dschema.BoolAttribute{Computed: true},
			"interfaces":               dschema.SetAttribute{Computed: true, ElementType: types.StringType},
			"strict_interface_binding": dschema.BoolAttribute{Computed: true},
			"dns_port":                 dschema.Int64Attribute{Computed: true},
		},
	}
}
func settingsAPIToModel(settings *apidnsmasq.GeneralSettingsWrapper) *settingsResourceModel {
	general := settings.Dnsmasq
	return &settingsResourceModel{
		ID:                     types.StringValue("dnsmasq_settings"),
		Enabled:                types.BoolValue(tools.StringToBool(general.IsEnabled)),
		Interfaces:             tools.StringSliceToSet([]string(general.Interface)),
		StrictInterfaceBinding: types.BoolValue(tools.StringToBool(general.StrictInterfaceBinding)),
		DNSPort:                types.Int64Value(tools.StringToInt64(general.DNS_Port)),
	}
}

func applySettingsModel(general *apidnsmasq.GeneralSettings, model *settingsResourceModel) {
	if !model.Enabled.IsNull() && !model.Enabled.IsUnknown() {
		general.IsEnabled = tools.BoolToString(model.Enabled.ValueBool())
	}
	if !model.Interfaces.IsNull() && !model.Interfaces.IsUnknown() {
		general.Interface = api.SelectedMapList(tools.SetToStringSlice(model.Interfaces))
	}
	if !model.StrictInterfaceBinding.IsNull() && !model.StrictInterfaceBinding.IsUnknown() {
		general.StrictInterfaceBinding = tools.BoolToString(model.StrictInterfaceBinding.ValueBool())
	}
	if !model.DNSPort.IsNull() && !model.DNSPort.IsUnknown() {
		general.DNS_Port = tools.Int64ToString(model.DNSPort.ValueInt64())
	}
}
