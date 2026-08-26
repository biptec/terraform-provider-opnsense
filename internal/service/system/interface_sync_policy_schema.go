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

type interfaceSyncPolicyResourceModel struct {
	PolicyID    types.String `tfsdk:"policy_id"`
	Description types.String `tfsdk:"description"`
	Synchronize types.Bool   `tfsdk:"synchronize"`
	ID          types.String `tfsdk:"id"`
}

func interfaceSyncPolicyResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Defines a first-class interface synchronization policy in the api-extensions model. The router never infers this policy from an interface name, creator, VLAN range, or service type.",
		Attributes: map[string]schema.Attribute{
			"policy_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Stable policy identifier shared by WebUI, API and Terraform.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-z][a-z0-9_-]{0,31}$`),
						"must be 1-32 lowercase letters, digits, underscores or dashes and start with a letter",
					),
				},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"description": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Operator-visible policy description.",
			},
			"synchronize": schema.BoolAttribute{
				Required:            true,
				MarkdownDescription: "Whether interfaces assigned to this policy are eligible for HA VLAN/interface synchronization when the corresponding HA service is enabled.",
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "OPNsense MVC object UUID.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}
