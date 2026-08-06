package system

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type sshResourceModel struct {
	Enabled                types.Bool   `tfsdk:"enabled"`
	Port                   types.Int64  `tfsdk:"port"`
	Interfaces             types.Set    `tfsdk:"interfaces"`
	PasswordAuthentication types.Bool   `tfsdk:"password_authentication"`
	PermitRootLogin        types.Bool   `tfsdk:"permit_root_login"`
	AllowReaddress         types.Bool   `tfsdk:"allow_readdress"`
	ID                     types.String `tfsdk:"id"`
}

func sshResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manages OPNsense SSH listener settings through `os-api-extensions`. This is a singleton resource. The `os-api-extensions` package must be installed first.",
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Enable the OPNsense SSH service. Defaults to `true`.",
			},
			"port": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(22),
				Validators: []validator.Int64{
					int64validator.Between(1, 65535),
				},
				MarkdownDescription: "SSH TCP port. Defaults to `22`.",
			},
			"interfaces": schema.SetAttribute{
				Required:            true,
				ElementType:         types.StringType,
				Validators:          []validator.Set{setvalidator.SizeAtLeast(1)},
				MarkdownDescription: "Logical OPNsense interfaces on which SSH listens. An explicit non-empty set is required.",
			},
			"password_authentication": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Allow SSH password authentication. Defaults to `false`.",
			},
			"permit_root_login": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Permit direct root login over SSH. Defaults to `false`.",
			},
			"allow_readdress": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
				MarkdownDescription: "Explicitly allow changes to SSH enablement, port, or listener interfaces. " +
					"Keep disabled unless an alternate management path or console is available.",
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Fixed singleton identifier `system_ssh`.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}
