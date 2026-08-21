package bind

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type secondaryDomainTransferConfigValidator struct{}

func (secondaryDomainTransferConfigValidator) Description(context.Context) string {
	return "shared transfer_key_id and legacy inline secondary transfer credentials are mutually exclusive"
}

func (v secondaryDomainTransferConfigValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (secondaryDomainTransferConfigValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var keyID, algorithm, name, secret types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("transfer_key_id"), &keyID)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("transfer_key_algorithm"), &algorithm)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("transfer_key_name"), &name)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("transfer_key"), &secret)...)
	if resp.Diagnostics.HasError() || keyID.IsUnknown() || algorithm.IsUnknown() || name.IsUnknown() || secret.IsUnknown() {
		return
	}
	if keyID.IsNull() || keyID.ValueString() == "" {
		return
	}
	legacyConfigured := (!algorithm.IsNull() && algorithm.ValueString() != "") ||
		(!name.IsNull() && name.ValueString() != "") || (!secret.IsNull() && secret.ValueString() != "")
	if legacyConfigured {
		resp.Diagnostics.AddAttributeError(
			path.Root("transfer_key_id"),
			"Conflicting BIND Secondary Transfer Key",
			"transfer_key_id is mutually exclusive with transfer_key_algorithm, transfer_key_name, and transfer_key.",
		)
	}
}
