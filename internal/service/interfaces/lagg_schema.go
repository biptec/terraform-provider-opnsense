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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type laggResourceModel struct {
	Device          types.String `tfsdk:"device"`
	Members         types.Set    `tfsdk:"members"`
	PrimaryMember   types.String `tfsdk:"primary_member"`
	Protocol        types.String `tfsdk:"protocol"`
	LACPFastTimeout types.Bool   `tfsdk:"lacp_fast_timeout"`
	UseFlowID       types.String `tfsdk:"use_flow_id"`
	HashLayers      types.Set    `tfsdk:"hash_layers"`
	LACPStrict      types.String `tfsdk:"lacp_strict"`
	MTU             types.Int64  `tfsdk:"mtu"`
	Description     types.String `tfsdk:"description"`
	Id              types.String `tfsdk:"id"`
}

func laggResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manages an OPNsense link aggregation interface.",
		Attributes: map[string]schema.Attribute{
			"device":            schema.StringAttribute{Computed: true, MarkdownDescription: "Automatically allocated LAGG device name."},
			"members":           schema.SetAttribute{Required: true, ElementType: types.StringType, Validators: []validator.Set{setvalidator.SizeAtLeast(1)}, MarkdownDescription: "Member interfaces."},
			"primary_member":    schema.StringAttribute{Optional: true, MarkdownDescription: "Primary member for failover mode. Must also be in members."},
			"protocol":          schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("lacp"), Validators: []validator.String{stringvalidator.OneOf("none", "lacp", "failover", "fec", "loadbalance", "roundrobin")}, MarkdownDescription: "Aggregation protocol."},
			"lacp_fast_timeout": schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), MarkdownDescription: "Enable LACP fast timeout."},
			"use_flow_id":       schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), Validators: []validator.String{stringvalidator.OneOf("", "1", "0")}, MarkdownDescription: "Use RSS flow ID: empty for system default, `1` for yes, or `0` for no."},
			"hash_layers":       schema.SetAttribute{Optional: true, Computed: true, Default: setdefault.StaticValue(tools.EmptySetValue(types.StringType)), ElementType: types.StringType, MarkdownDescription: "Hash layers selected from `l2`, `l3`, and `l4`."},
			"lacp_strict":       schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), Validators: []validator.String{stringvalidator.OneOf("", "1", "0")}, MarkdownDescription: "LACP strict mode: empty for system default, `1` for yes, or `0` for no."},
			"mtu":               schema.Int64Attribute{Optional: true, Validators: []validator.Int64{int64validator.Between(576, 65535)}, MarkdownDescription: "Optional MTU. When unset, OPNsense uses the smallest member MTU."},
			"description":       schema.StringAttribute{Optional: true, MarkdownDescription: "Optional description."},
			"id":                schema.StringAttribute{Computed: true, MarkdownDescription: "UUID of the LAGG configuration.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		},
	}
}

func laggDataSourceSchema() dschema.Schema {
	return dschema.Schema{MarkdownDescription: "Reads an OPNsense link aggregation interface.", Attributes: map[string]dschema.Attribute{
		"id": dschema.StringAttribute{Required: true}, "device": dschema.StringAttribute{Computed: true},
		"members": dschema.SetAttribute{Computed: true, ElementType: types.StringType}, "primary_member": dschema.StringAttribute{Computed: true},
		"protocol": dschema.StringAttribute{Computed: true}, "lacp_fast_timeout": dschema.BoolAttribute{Computed: true},
		"use_flow_id": dschema.StringAttribute{Computed: true}, "hash_layers": dschema.SetAttribute{Computed: true, ElementType: types.StringType},
		"lacp_strict": dschema.StringAttribute{Computed: true}, "mtu": dschema.Int64Attribute{Computed: true}, "description": dschema.StringAttribute{Computed: true},
	}}
}

func convertLaggSchemaToStruct(d *laggResourceModel) (*apiinterfaces.Lagg, error) {
	members := tools.SetToStringSlice(d.Members)
	memberSet := make(map[string]struct{}, len(members))
	for _, member := range members {
		memberSet[member] = struct{}{}
	}
	protocol := d.Protocol.ValueString()
	primary := d.PrimaryMember.ValueString()
	if primary != "" {
		if _, exists := memberSet[primary]; !exists {
			return nil, fmt.Errorf("primary_member %q is not in members", primary)
		}
		if protocol != "failover" {
			return nil, fmt.Errorf("primary_member may only be configured with failover protocol")
		}
	}
	if d.LACPFastTimeout.ValueBool() && protocol != "lacp" {
		return nil, fmt.Errorf("lacp_fast_timeout may only be enabled with lacp protocol")
	}
	if d.LACPStrict.ValueString() != "" && protocol != "lacp" {
		return nil, fmt.Errorf("lacp_strict may only be configured with lacp protocol")
	}
	hashLayers := tools.SetToStringSlice(d.HashLayers)
	for _, layer := range hashLayers {
		if layer != "l2" && layer != "l3" && layer != "l4" {
			return nil, fmt.Errorf("invalid hash layer %q", layer)
		}
	}
	if len(hashLayers) > 0 && protocol != "lacp" && protocol != "loadbalance" {
		return nil, fmt.Errorf("hash_layers require lacp or loadbalance protocol")
	}
	if d.UseFlowID.ValueString() != "" && protocol != "lacp" && protocol != "loadbalance" {
		return nil, fmt.Errorf("use_flow_id requires lacp or loadbalance protocol")
	}
	return &apiinterfaces.Lagg{
		Members: api.SelectedMapList(members), PrimaryMember: api.SelectedMap(primary), Protocol: api.SelectedMap(protocol),
		LACPFastTimeout: tools.BoolToString(d.LACPFastTimeout.ValueBool()), UseFlowID: api.SelectedMap(d.UseFlowID.ValueString()),
		HashLayers: api.SelectedMapList(hashLayers), LACPStrict: api.SelectedMap(d.LACPStrict.ValueString()),
		MTU: optionalIntString(d.MTU), Description: d.Description.ValueString(),
	}, nil
}

func convertLaggStructToSchema(d *apiinterfaces.Lagg) (*laggResourceModel, error) {
	protocol := d.Protocol.String()
	if protocol == "" {
		protocol = "lacp"
	}
	return &laggResourceModel{
		Device: types.StringValue(d.Device), Members: tools.StringSliceToSet([]string(d.Members)), PrimaryMember: tools.StringOrNull(d.PrimaryMember.String()),
		Protocol: types.StringValue(protocol), LACPFastTimeout: types.BoolValue(tools.StringToBool(d.LACPFastTimeout)),
		UseFlowID: types.StringValue(d.UseFlowID.String()), HashLayers: tools.StringSliceToSet([]string(d.HashLayers)),
		LACPStrict: types.StringValue(d.LACPStrict.String()), MTU: tools.StringToInt64Null(d.MTU), Description: tools.StringOrNull(d.Description),
	}, nil
}
