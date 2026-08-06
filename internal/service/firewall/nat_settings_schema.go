package firewall

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const natSettingsID = "firewall_nat_settings"

type natSettingsResourceModel struct {
	ID   types.String `tfsdk:"id"`
	Mode types.String `tfsdk:"mode"`
}

func natSettingsResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manages the global OPNsense outbound Source NAT generation mode. This is a singleton resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Fixed singleton identifier `firewall_nat_settings`.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"mode": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf("automatic", "hybrid", "advanced", "disabled"),
				},
				MarkdownDescription: "Outbound Source NAT generation mode: `automatic`, `hybrid`, `advanced` (manual), or `disabled`. Use `hybrid` when Terraform manages explicit NO-NAT rules while retaining automatic rules for other networks.",
			},
		},
	}
}
