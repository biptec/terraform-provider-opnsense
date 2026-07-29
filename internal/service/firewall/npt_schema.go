package firewall

import (
	"regexp"

	"github.com/biptec/opnsense-go/pkg/api"
	apifirewall "github.com/biptec/opnsense-go/pkg/firewall"
	"github.com/biptec/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var ipv6PrefixValidator = stringvalidator.RegexMatches(
	regexp.MustCompile(`^[0-9A-Fa-f:]+(?:/(?:[0-9]|[1-9][0-9]|1[01][0-9]|12[0-8]))?$`),
	"must be an IPv6 address or prefix, for example fd00:1::/48",
)

type nptResourceModel struct {
	Enabled        types.Bool   `tfsdk:"enabled"`
	Log            types.Bool   `tfsdk:"log"`
	Sequence       types.Int64  `tfsdk:"sequence"`
	Categories     types.Set    `tfsdk:"categories"`
	Description    types.String `tfsdk:"description"`
	Interface      types.String `tfsdk:"interface"`
	SourceNet      types.String `tfsdk:"source_net"`
	DestinationNet types.String `tfsdk:"destination_net"`
	TrackInterface types.String `tfsdk:"track_interface"`
	Id             types.String `tfsdk:"id"`
}

func nptResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Configures an OPNsense IPv6 Network Prefix Translation rule.",
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{MarkdownDescription: "Enable this NPT rule.", Optional: true, Computed: true, Default: booldefault.StaticBool(true)},
			"log":     schema.BoolAttribute{MarkdownDescription: "Log packets handled by this rule.", Optional: true, Computed: true, Default: booldefault.StaticBool(false)},
			"sequence": schema.Int64Attribute{
				MarkdownDescription: "Rule sequence.", Optional: true, Computed: true, Default: int64default.StaticInt64(1),
				Validators: []validator.Int64{int64validator.Between(1, 999999)},
			},
			"categories": schema.SetAttribute{
				MarkdownDescription: "Firewall category UUIDs.", Optional: true, Computed: true,
				ElementType: types.StringType, Default: setdefault.StaticValue(tools.EmptySetValue(types.StringType)),
			},
			"description": schema.StringAttribute{MarkdownDescription: "Optional description.", Optional: true},
			"interface": schema.StringAttribute{
				MarkdownDescription: "Logical interface where the NPT rule applies.",
				Optional:            true, Computed: true, Default: stringdefault.StaticString("lan"),
			},
			"source_net": schema.StringAttribute{
				MarkdownDescription: "Internal IPv6 prefix.", Required: true,
				Validators: []validator.String{ipv6PrefixValidator},
			},
			"destination_net": schema.StringAttribute{
				MarkdownDescription: "External IPv6 prefix. Leave unset when using `track_interface`.", Optional: true,
				Validators: []validator.String{ipv6PrefixValidator},
			},
			"track_interface": schema.StringAttribute{
				MarkdownDescription: "Interface whose delegated prefix is used when `destination_net` is unset.", Optional: true,
			},
			"id": schema.StringAttribute{Computed: true, MarkdownDescription: "UUID of the NPT rule.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		},
	}
}

func nptDataSourceSchema() dschema.Schema {
	return dschema.Schema{
		MarkdownDescription: "Reads an OPNsense NPTv6 rule by UUID.",
		Attributes: map[string]dschema.Attribute{
			"id":              dschema.StringAttribute{Required: true},
			"enabled":         dschema.BoolAttribute{Computed: true},
			"log":             dschema.BoolAttribute{Computed: true},
			"sequence":        dschema.Int64Attribute{Computed: true},
			"categories":      dschema.SetAttribute{Computed: true, ElementType: types.StringType},
			"description":     dschema.StringAttribute{Computed: true},
			"interface":       dschema.StringAttribute{Computed: true},
			"source_net":      dschema.StringAttribute{Computed: true},
			"destination_net": dschema.StringAttribute{Computed: true},
			"track_interface": dschema.StringAttribute{Computed: true},
		},
	}
}

func convertNptSchemaToStruct(d *nptResourceModel) (*apifirewall.Npt, error) {
	return &apifirewall.Npt{
		Enabled:        tools.BoolToString(d.Enabled.ValueBool()),
		Log:            tools.BoolToString(d.Log.ValueBool()),
		Sequence:       tools.Int64ToString(d.Sequence.ValueInt64()),
		Categories:     api.SelectedMapList(tools.SetToStringSlice(d.Categories)),
		Description:    d.Description.ValueString(),
		Interface:      api.SelectedMap(d.Interface.ValueString()),
		SourceNet:      d.SourceNet.ValueString(),
		DestinationNet: d.DestinationNet.ValueString(),
		TrackInterface: api.SelectedMap(d.TrackInterface.ValueString()),
	}, nil
}

func convertNptStructToSchema(d *apifirewall.Npt) (*nptResourceModel, error) {
	return &nptResourceModel{
		Enabled:        types.BoolValue(tools.StringToBool(d.Enabled)),
		Log:            types.BoolValue(tools.StringToBool(d.Log)),
		Sequence:       tools.StringToInt64Null(d.Sequence),
		Categories:     tools.StringSliceToSet([]string(d.Categories)),
		Description:    tools.StringOrNull(d.Description),
		Interface:      types.StringValue(d.Interface.String()),
		SourceNet:      types.StringValue(d.SourceNet),
		DestinationNet: tools.StringOrNull(d.DestinationNet),
		TrackInterface: tools.StringOrNull(d.TrackInterface.String()),
	}, nil
}
