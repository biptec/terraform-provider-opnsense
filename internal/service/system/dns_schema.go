package system

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type dnsSettingsResourceModel struct {
	Servers         types.List   `tfsdk:"servers"`
	AllowOverride   types.Bool   `tfsdk:"allow_override"`
	UseLocalService types.Bool   `tfsdk:"use_local_service"`
	ID              types.String `tfsdk:"id"`
}

func dnsSettingsResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manages OPNsense system DNS resolver settings through `os-api-extensions`. This is a singleton resource. `os-api-extensions` version 0.11 or newer must be installed first.",
		Attributes: map[string]schema.Attribute{
			"servers": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Validators: []validator.List{
					listvalidator.SizeBetween(1, 8),
				},
				MarkdownDescription: "Ordered IPv4/IPv6 DNS server addresses used by the OPNsense host resolver.",
			},
			"allow_override": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Allow dynamic interfaces to add DNS servers. Defaults to `false`.",
			},
			"use_local_service": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Prepend localhost when a local DNS service is active. Defaults to `true`.",
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Fixed singleton identifier `system_dns`.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}
