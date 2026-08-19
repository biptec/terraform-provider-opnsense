package quagga

import (
	"github.com/biptec/opnsense-go/pkg/api"
	"github.com/biptec/opnsense-go/pkg/quagga"
	"github.com/biptec/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
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

type ospfPrefixListModel struct {
	Enabled        types.Bool   `tfsdk:"enabled"`
	Name           types.String `tfsdk:"name"`
	SequenceNumber types.Int64  `tfsdk:"sequence_number"`
	Action         types.String `tfsdk:"action"`
	Network        types.String `tfsdk:"network"`
	ID             types.String `tfsdk:"id"`
}

func ospfPrefixListResourceSchema() rschema.Schema {
	return rschema.Schema{Version: 1, MarkdownDescription: "Manages one OPNsense OSPFv2 prefix-list entry.", Attributes: map[string]rschema.Attribute{
		"enabled":         rschema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true), MarkdownDescription: "Whether to enable this prefix-list entry. Defaults to `true`."},
		"name":            rschema.StringAttribute{Required: true, MarkdownDescription: "Name of the prefix list."},
		"sequence_number": rschema.Int64Attribute{Required: true, Validators: []validator.Int64{int64validator.Between(10, 99)}, MarkdownDescription: "Sequence number used to order this prefix-list entry (10-99)."},
		"action":          rschema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("permit"), Validators: []validator.String{stringvalidator.OneOf("permit", "deny")}, MarkdownDescription: "Whether matching routes are permitted or denied. Defaults to `permit`."},
		"network":         rschema.StringAttribute{Required: true, MarkdownDescription: "Network prefix matched by this entry."},
		"id":              rschema.StringAttribute{Computed: true, MarkdownDescription: "UUID of the prefix-list entry.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
	}}
}

func ospfPrefixListDataSourceSchema() dsschema.Schema {
	return dsschema.Schema{MarkdownDescription: "Reads one OPNsense OSPFv2 prefix-list entry by UUID.", Attributes: map[string]dsschema.Attribute{
		"enabled":         dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether this prefix-list entry is enabled."},
		"name":            dsschema.StringAttribute{Computed: true, MarkdownDescription: "Name of the prefix list."},
		"sequence_number": dsschema.Int64Attribute{Computed: true, MarkdownDescription: "Sequence number used to order this prefix-list entry."},
		"action":          dsschema.StringAttribute{Computed: true, MarkdownDescription: "Whether matching routes are permitted or denied."},
		"network":         dsschema.StringAttribute{Computed: true, MarkdownDescription: "Network prefix matched by this entry."},
		"id":              dsschema.StringAttribute{Required: true, MarkdownDescription: "UUID of the prefix-list entry."},
	}}
}

func ospfPrefixListToAPI(d *ospfPrefixListModel) *quagga.OSPFPrefixList {
	return &quagga.OSPFPrefixList{Enabled: tools.BoolToString(d.Enabled.ValueBool()), Name: d.Name.ValueString(), SequenceNumber: tools.Int64ToString(d.SequenceNumber.ValueInt64()), Action: api.SelectedMap(d.Action.ValueString()), Network: d.Network.ValueString()}
}
func ospfPrefixListFromAPI(d *quagga.OSPFPrefixList, id string) *ospfPrefixListModel {
	return &ospfPrefixListModel{Enabled: types.BoolValue(tools.StringToBool(d.Enabled)), Name: types.StringValue(d.Name), SequenceNumber: types.Int64Value(tools.StringToInt64(d.SequenceNumber)), Action: types.StringValue(d.Action.String()), Network: types.StringValue(d.Network), ID: types.StringValue(id)}
}
