package bind

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type primaryDomainTransferConfigValidator struct{}

func (primaryDomainTransferConfigValidator) Description(context.Context) string {
	return "also_notify requires transfer_key_id so BIND NOTIFY messages are authenticated"
}

func (v primaryDomainTransferConfigValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (primaryDomainTransferConfigValidator) ValidateResource(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	var alsoNotify types.Set
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("also_notify"), &alsoNotify)...)
	if resp.Diagnostics.HasError() || alsoNotify.IsNull() || alsoNotify.IsUnknown() || len(alsoNotify.Elements()) == 0 {
		return
	}

	var transferKeyID types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("transfer_key_id"), &transferKeyID)...)
	if resp.Diagnostics.HasError() || transferKeyID.IsUnknown() {
		return
	}
	if err := validatePrimaryDomainTransfer(transferKeyID, alsoNotify); err == nil {
		return
	}

	resp.Diagnostics.AddAttributeError(
		path.Root("transfer_key_id"),
		"Missing Transfer TSIG Key",
		"transfer_key_id must reference an enabled opnsense_bind_tsig_key when also_notify contains secondary nameserver addresses.",
	)
}

func validatePrimaryDomainTransfer(transferKeyID types.String, alsoNotify types.Set) error {
	if alsoNotify.IsNull() || alsoNotify.IsUnknown() || len(alsoNotify.Elements()) == 0 {
		return nil
	}
	if transferKeyID.IsUnknown() {
		return nil
	}
	if transferKeyID.IsNull() || transferKeyID.ValueString() == "" {
		return errors.New("also_notify requires transfer_key_id")
	}
	return nil
}
