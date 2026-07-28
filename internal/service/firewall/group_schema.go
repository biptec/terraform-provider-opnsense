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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type groupResourceModel struct {
	Name        types.String `tfsdk:"name"`
	Members     types.Set    `tfsdk:"members"`
	NoGroup     types.Bool   `tfsdk:"no_gui_group"`
	Sequence    types.Int64  `tfsdk:"sequence"`
	Description types.String `tfsdk:"description"`
	Id          types.String `tfsdk:"id"`
}

func groupResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Configures an OPNsense firewall interface group.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Group name. Maximum 15 characters; letters, digits, and underscores only; must not start or end with a digit.",
				Required:            true,
				Validators:          []validator.String{stringvalidator.RegexMatches(regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]{0,13}[A-Za-z_]$|^[A-Za-z_]$`), "must be 1-15 characters, use only letters, digits and underscores, and not start or end with a digit")},
			},
			"members": schema.SetAttribute{
				MarkdownDescription: "Logical interface identifiers that belong to the group.",
				Required:            true, ElementType: types.StringType,
			},
			"no_gui_group": schema.BoolAttribute{
				MarkdownDescription: "Do not group these members in the Interfaces menu.",
				Optional:            true, Computed: true, Default: booldefault.StaticBool(false),
			},
			"sequence": schema.Int64Attribute{
				MarkdownDescription: "Priority used when sorting groups.",
				Optional:            true, Computed: true, Default: int64default.StaticInt64(0),
				Validators: []validator.Int64{int64validator.Between(0, 9999)},
			},
			"description": schema.StringAttribute{MarkdownDescription: "Optional description.", Optional: true},
			"id":          schema.StringAttribute{Computed: true, MarkdownDescription: "UUID of the group.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		},
	}
}

func groupDataSourceSchema() dschema.Schema {
	return dschema.Schema{
		MarkdownDescription: "Reads an OPNsense firewall interface group by UUID.",
		Attributes: map[string]dschema.Attribute{
			"id":           dschema.StringAttribute{Required: true},
			"name":         dschema.StringAttribute{Computed: true},
			"members":      dschema.SetAttribute{Computed: true, ElementType: types.StringType},
			"no_gui_group": dschema.BoolAttribute{Computed: true},
			"sequence":     dschema.Int64Attribute{Computed: true},
			"description":  dschema.StringAttribute{Computed: true},
		},
	}
}

func convertGroupSchemaToStruct(d *groupResourceModel) (*apifirewall.Group, error) {
	return &apifirewall.Group{
		Name:        d.Name.ValueString(),
		Members:     api.SelectedMapList(tools.SetToStringSlice(d.Members)),
		NoGroup:     tools.BoolToString(d.NoGroup.ValueBool()),
		Sequence:    tools.Int64ToString(d.Sequence.ValueInt64()),
		Description: d.Description.ValueString(),
	}, nil
}

func convertGroupStructToSchema(d *apifirewall.Group) (*groupResourceModel, error) {
	return &groupResourceModel{
		Name:        types.StringValue(d.Name),
		Members:     tools.StringSliceToSet([]string(d.Members)),
		NoGroup:     types.BoolValue(tools.StringToBool(d.NoGroup)),
		Sequence:    tools.StringToInt64Null(d.Sequence),
		Description: tools.StringOrNull(d.Description),
	}, nil
}
