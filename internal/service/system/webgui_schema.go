package system

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type webguiResourceModel struct {
	Protocol              types.String `tfsdk:"protocol"`
	Port                  types.Int64  `tfsdk:"port"`
	Interfaces            types.Set    `tfsdk:"interfaces"`
	CertificateRef        types.String `tfsdk:"certificate_ref"`
	SessionTimeoutMinutes types.Int64  `tfsdk:"session_timeout_minutes"`
	HSTS                  types.Bool   `tfsdk:"hsts"`
	DisableHTTPRedirect   types.Bool   `tfsdk:"disable_http_redirect"`
	AlternateHostnames    types.Set    `tfsdk:"alternate_hostnames"`
	AllowReaddress        types.Bool   `tfsdk:"allow_readdress"`
	ID                    types.String `tfsdk:"id"`
}

func webguiResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manages OPNsense Web GUI listener settings through `os-api-extensions`. This is a singleton resource. The `os-api-extensions` package must be installed first.",
		Attributes: map[string]schema.Attribute{
			"protocol": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.OneOf("http", "https"),
				},
				MarkdownDescription: "Web GUI protocol. When omitted, the current protocol is preserved.",
			},
			"port": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Validators: []validator.Int64{
					int64validator.Between(1, 65535),
				},
				MarkdownDescription: "Web GUI TCP port. When omitted, the current port is preserved.",
			},
			"interfaces": schema.SetAttribute{
				Required:            true,
				ElementType:         types.StringType,
				Validators:          []validator.Set{setvalidator.SizeAtLeast(1)},
				MarkdownDescription: "Logical OPNsense interfaces on which the Web GUI and API listen. An explicit non-empty set is required.",
			},
			"certificate_ref": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "OPNsense certificate reference used by HTTPS. When omitted, the current certificate is preserved.",
			},
			"session_timeout_minutes": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Validators: []validator.Int64{
					int64validator.Between(1, 86400),
				},
				MarkdownDescription: "Session timeout in minutes. When omitted, the current setting is preserved.",
			},
			"hsts": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Enable HTTP Strict Transport Security. When omitted, the current setting is preserved.",
			},
			"disable_http_redirect": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Disable automatic HTTP-to-HTTPS redirection. When omitted, the current setting is preserved.",
			},
			"alternate_hostnames": schema.SetAttribute{
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Additional valid Web GUI hostnames. When omitted, the current set is preserved.",
			},
			"allow_readdress": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
				MarkdownDescription: "Explicitly allow listener interface, protocol, port, or certificate changes that can disconnect the provider. " +
					"Keep disabled unless the target management path and provider URI are known to remain reachable.",
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Fixed singleton identifier `system_webgui`.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}
