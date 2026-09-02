package system

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type pluginResourceModel struct {
	Name               types.String `tfsdk:"name"`
	Locked             types.Bool   `tfsdk:"locked"`
	UninstallOnDestroy types.Bool   `tfsdk:"uninstall_on_destroy"`
	Version            types.String `tfsdk:"version"`
	Repository         types.String `tfsdk:"repository"`
	Installed          types.Bool   `tfsdk:"installed"`
	Provided           types.Bool   `tfsdk:"provided"`
	ID                 types.String `tfsdk:"id"`
}

func pluginResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Installs and manages an OPNsense plugin. Normal Terraform refresh reads local package state through `os-api-extensions` 0.12 or newer and never refreshes remote package repositories. Repository discovery is used only during Create when installation is actually required.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Plugin package name, for example `os-bind` or `os-caddy`.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^os-[A-Za-z0-9][A-Za-z0-9+_.-]*$`),
						"must be an OPNsense plugin package name beginning with os-",
					),
				},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"locked": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Lock the installed package against firmware upgrades. Defaults to `false`.",
			},
			"uninstall_on_destroy": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
				MarkdownDescription: "Uninstall the plugin when the Terraform resource is destroyed. " +
					"Defaults to `false`; without this explicit opt-in, destroy removes only the Terraform state entry.",
			},
			"version": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Installed plugin version.",
			},
			"repository": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Repository that supplied the plugin.",
			},
			"installed": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the plugin is installed.",
			},
			"provided": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the plugin is present in the locally cached package repository metadata. Create may explicitly refresh repository metadata when installation is required.",
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Plugin package name.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}
