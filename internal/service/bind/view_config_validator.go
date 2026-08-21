package bind

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type viewTSIGMatchConfigValidator struct{}

func (viewTSIGMatchConfigValidator) Description(context.Context) string {
	return "a BIND view cannot both include and exclude the same client TSIG key"
}

func (v viewTSIGMatchConfigValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (viewTSIGMatchConfigValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var included, excluded types.Set
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("match_client_tsig_key_ids"), &included)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("exclude_match_client_tsig_key_ids"), &excluded)...)
	if resp.Diagnostics.HasError() || included.IsNull() || included.IsUnknown() || excluded.IsNull() || excluded.IsUnknown() {
		return
	}

	if validateViewTSIGSelectors(included, excluded) == nil {
		return
	}
	resp.Diagnostics.AddAttributeError(
		path.Root("exclude_match_client_tsig_key_ids"),
		"Conflicting BIND View TSIG Selector",
		"The same TSIG key UUID cannot appear in both match_client_tsig_key_ids and exclude_match_client_tsig_key_ids.",
	)
}

func validateViewTSIGSelectors(included, excluded types.Set) error {
	if included.IsNull() || included.IsUnknown() || excluded.IsNull() || excluded.IsUnknown() {
		return nil
	}
	excludedIDs := make(map[string]struct{}, len(excluded.Elements()))
	for _, value := range excluded.Elements() {
		id, ok := value.(types.String)
		if ok && !id.IsNull() && !id.IsUnknown() {
			excludedIDs[id.ValueString()] = struct{}{}
		}
	}
	for _, value := range included.Elements() {
		id, ok := value.(types.String)
		if !ok || id.IsNull() || id.IsUnknown() {
			continue
		}
		if _, found := excludedIDs[id.ValueString()]; found {
			return errors.New("same TSIG key appears in include and exclude selectors")
		}
	}
	return nil
}
