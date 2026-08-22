package dns

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type serviceCutoverResourceModel struct {
	Target               types.String `tfsdk:"target"`
	AllowCutover         types.Bool   `tfsdk:"allow_cutover"`
	VerifyTimeoutSeconds types.Int64  `tfsdk:"verify_timeout_seconds"`
	ActiveService        types.String `tfsdk:"active_service"`
	ID                   types.String `tfsdk:"id"`
}

func serviceCutoverResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Coordinates DNS port ownership between Unbound and BIND. Initial resource creation may select the configured target without a separate cutover approval; later owner changes remain guarded. When adopting an existing appliance without selecting a service owner, import this resource before planning changes. It is the exclusive Terraform owner of BIND enabled state, Unbound enabled state, and the dnsmasq DNS port. The resource preflights BIND while it is disabled, verifies the selected service runtime, and restores the previous service state when activation fails.",
		Attributes: map[string]schema.Attribute{
			"target": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Desired DNS service owner: `unbound` or `bind`. The transition sets the dnsmasq DNS port to zero without changing whether dnsmasq itself is enabled.",
				Validators: []validator.String{
					stringvalidator.OneOf("unbound", "bind"),
				},
			},
			"allow_cutover": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Explicitly permit changing the DNS owner after this resource is already managed. Initial resource creation may select target without this flag. Keep false during normal steady-state operation.",
			},
			"verify_timeout_seconds": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(60),
				MarkdownDescription: "Maximum time to wait for the selected DNS service state to converge.",
				Validators: []validator.Int64{
					int64validator.Between(5, 600),
				},
			},
			"active_service": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Observed DNS owner. Values other than `unbound` or `bind` indicate an inconsistent external state.",
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Always `dns_service_cutover`.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}
