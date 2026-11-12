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

type haproxyTLSIdentityResourceModel struct {
	FrontendName types.String `tfsdk:"frontend_name"`
	DNSName      types.String `tfsdk:"dns_name"`
	PolicyID     types.String `tfsdk:"policy_id"`
	ID           types.String `tfsdk:"id"`
}

func haproxyTLSIdentityResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Registers one consumer TLS identity against an existing Core-owned HAProxy frontend. The Router generates and retains the private key locally, requests the FreeIPA service certificate, maintains the local Trust object and additive frontend certificate membership, and synchronizes only non-secret identity metadata to the HA peer.",
		Attributes: map[string]schema.Attribute{
			"frontend_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Exact semantic name of the existing TLS-terminating HAProxy frontend. The resource references but does not own that frontend.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[A-Za-z0-9._-]{1,255}$`),
						"must be a valid HAProxy frontend semantic name",
					),
				},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"dns_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Consumer DNS identity. The Router derives HAPROXY/<dns_name> FreeIPA principal, exact single-SAN certificate request, local key path and Trust identity from this value.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^(?i:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?(?:\.[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?)+)$`),
						"must be a valid DNS hostname",
					),
				},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"policy_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Stable ID of the HA synchronization policy. Shared Arcturus consumers use a synchronize=true endpoint policy so only non-secret identity metadata is projected to the peer.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-z][a-z0-9_-]{0,31}$`),
						"must be a valid HA synchronization policy ID",
					),
				},
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Receiver-local OPNsense MVC object UUID. It is not part of the cross-peer identity contract.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}
