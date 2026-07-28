package routing

import (
	"regexp"

	"github.com/biptec/opnsense-go/pkg/api"
	apirouting "github.com/biptec/opnsense-go/pkg/routing"
	"github.com/biptec/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type gatewayGroupResourceModel struct {
	Name        types.String `tfsdk:"name"`
	Tier1       types.Set    `tfsdk:"tier1"`
	Tier2       types.Set    `tfsdk:"tier2"`
	Tier3       types.Set    `tfsdk:"tier3"`
	Tier4       types.Set    `tfsdk:"tier4"`
	Tier5       types.Set    `tfsdk:"tier5"`
	Trigger     types.String `tfsdk:"trigger"`
	PoolOptions types.String `tfsdk:"pool_options"`
	Description types.String `tfsdk:"description"`
	Id          types.String `tfsdk:"id"`
}

func gatewayGroupResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Configures an OPNsense gateway group for failover or load balancing.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Unique gateway-group name.", Required: true,
				Validators:    []validator.String{stringvalidator.RegexMatches(regexp.MustCompile(`^[a-zA-Z0-9_-]{1,32}$`), "must contain only letters, digits, underscores, or hyphens and be at most 32 characters")},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"tier1": schema.SetAttribute{MarkdownDescription: "Gateway names in tier 1. At least one gateway is required.", Required: true, ElementType: types.StringType},
			"tier2": schema.SetAttribute{MarkdownDescription: "Gateway names in tier 2.", Optional: true, Computed: true, Default: setdefault.StaticValue(tools.EmptySetValue(types.StringType)), ElementType: types.StringType},
			"tier3": schema.SetAttribute{MarkdownDescription: "Gateway names in tier 3.", Optional: true, Computed: true, Default: setdefault.StaticValue(tools.EmptySetValue(types.StringType)), ElementType: types.StringType},
			"tier4": schema.SetAttribute{MarkdownDescription: "Gateway names in tier 4.", Optional: true, Computed: true, Default: setdefault.StaticValue(tools.EmptySetValue(types.StringType)), ElementType: types.StringType},
			"tier5": schema.SetAttribute{MarkdownDescription: "Gateway names in tier 5.", Optional: true, Computed: true, Default: setdefault.StaticValue(tools.EmptySetValue(types.StringType)), ElementType: types.StringType},
			"trigger": schema.StringAttribute{
				MarkdownDescription: "Failure trigger: `down`, `downloss`, `downlatency`, or `downlosslatency`.",
				Optional:            true, Computed: true, Default: stringdefault.StaticString("down"),
				Validators: []validator.String{stringvalidator.OneOf("down", "downloss", "downlatency", "downlosslatency")},
			},
			"pool_options": schema.StringAttribute{
				MarkdownDescription: "Pool behaviour: empty for default, `round-robin`, or `round-robin sticky-address`.",
				Optional:            true, Computed: true, Default: stringdefault.StaticString(""),
				Validators: []validator.String{stringvalidator.OneOf("", "round-robin", "round-robin sticky-address")},
			},
			"description": schema.StringAttribute{MarkdownDescription: "Optional description.", Optional: true},
			"id":          schema.StringAttribute{Computed: true, MarkdownDescription: "UUID of the gateway group.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		},
	}
}

func gatewayGroupDataSourceSchema() dschema.Schema {
	return dschema.Schema{
		MarkdownDescription: "Reads an OPNsense gateway group by UUID.",
		Attributes: map[string]dschema.Attribute{
			"id":           dschema.StringAttribute{Required: true},
			"name":         dschema.StringAttribute{Computed: true},
			"tier1":        dschema.SetAttribute{Computed: true, ElementType: types.StringType},
			"tier2":        dschema.SetAttribute{Computed: true, ElementType: types.StringType},
			"tier3":        dschema.SetAttribute{Computed: true, ElementType: types.StringType},
			"tier4":        dschema.SetAttribute{Computed: true, ElementType: types.StringType},
			"tier5":        dschema.SetAttribute{Computed: true, ElementType: types.StringType},
			"trigger":      dschema.StringAttribute{Computed: true},
			"pool_options": dschema.StringAttribute{Computed: true},
			"description":  dschema.StringAttribute{Computed: true},
		},
	}
}

func convertGatewayGroupSchemaToStruct(d *gatewayGroupResourceModel) (*apirouting.GatewayGroup, error) {
	return &apirouting.GatewayGroup{
		Name:        d.Name.ValueString(),
		Tier1:       api.SelectedMapList(tools.SetToStringSlice(d.Tier1)),
		Tier2:       api.SelectedMapList(tools.SetToStringSlice(d.Tier2)),
		Tier3:       api.SelectedMapList(tools.SetToStringSlice(d.Tier3)),
		Tier4:       api.SelectedMapList(tools.SetToStringSlice(d.Tier4)),
		Tier5:       api.SelectedMapList(tools.SetToStringSlice(d.Tier5)),
		Trigger:     api.SelectedMap(d.Trigger.ValueString()),
		PoolOptions: api.SelectedMap(d.PoolOptions.ValueString()),
		Description: d.Description.ValueString(),
	}, nil
}

func convertGatewayGroupStructToSchema(d *apirouting.GatewayGroup) (*gatewayGroupResourceModel, error) {
	return &gatewayGroupResourceModel{
		Name:        types.StringValue(d.Name),
		Tier1:       tools.StringSliceToSet([]string(d.Tier1)),
		Tier2:       tools.StringSliceToSet([]string(d.Tier2)),
		Tier3:       tools.StringSliceToSet([]string(d.Tier3)),
		Tier4:       tools.StringSliceToSet([]string(d.Tier4)),
		Tier5:       tools.StringSliceToSet([]string(d.Tier5)),
		Trigger:     types.StringValue(d.Trigger.String()),
		PoolOptions: types.StringValue(d.PoolOptions.String()),
		Description: tools.StringOrNull(d.Description),
	}, nil
}
