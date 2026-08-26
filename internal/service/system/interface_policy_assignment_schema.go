package system

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type interfacePolicyAssignmentResourceModel struct {
	Interface types.String `tfsdk:"interface"`
	PolicyID  types.String `tfsdk:"policy_id"`
	ID        types.String `tfsdk:"id"`
}

func interfacePolicyAssignmentResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Assigns exactly one explicit HA interface policy to a logical OPNsense interface. No policy is inferred from the interface identifier.",
		Attributes: map[string]schema.Attribute{
			"interface": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Logical OPNsense interface identifier.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[A-Za-z0-9_.:-]{1,64}$`),
						"must be a valid logical interface identifier",
					),
				},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"policy_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Stable ID of an existing interface synchronization policy.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-z][a-z0-9_-]{0,31}$`),
						"must be a valid interface synchronization policy ID",
					),
				},
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "OPNsense MVC object UUID.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}
