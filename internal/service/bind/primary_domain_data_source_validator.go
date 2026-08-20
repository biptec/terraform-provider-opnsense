package bind

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type primaryDomainLookupConfigValidator struct{}

func (primaryDomainLookupConfigValidator) Description(context.Context) string {
	return "select a BIND primary domain by id, domain_name + view_name, or domain_name + view_id"
}

func (v primaryDomainLookupConfigValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func lookupStringConfigured(value types.String) bool {
	return !value.IsNull() && !value.IsUnknown()
}

func validPrimaryDomainLookupSelectors(id, domainName, viewName, viewID types.String) bool {
	if id.IsUnknown() || domainName.IsUnknown() || viewName.IsUnknown() || viewID.IsUnknown() {
		return true
	}
	idSet := lookupStringConfigured(id)
	domainSet := lookupStringConfigured(domainName)
	viewNameSet := lookupStringConfigured(viewName)
	viewIDSet := lookupStringConfigured(viewID)

	if idSet {
		return !domainSet && !viewNameSet && !viewIDSet
	}
	return domainSet && (viewNameSet != viewIDSet)
}

func (primaryDomainLookupConfigValidator) ValidateDataSource(
	ctx context.Context,
	req datasource.ValidateConfigRequest,
	resp *datasource.ValidateConfigResponse,
) {
	var id, domainName, viewName, viewID types.String
	for attribute, target := range map[string]*types.String{
		"id": &id, "domain_name": &domainName, "view_name": &viewName, "view_id": &viewID,
	} {
		resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root(attribute), target)...)
	}
	if resp.Diagnostics.HasError() || validPrimaryDomainLookupSelectors(id, domainName, viewName, viewID) {
		return
	}
	resp.Diagnostics.AddError(
		"Invalid BIND Primary Domain Lookup",
		"Configure exactly one lookup mode: id; domain_name + view_name; or domain_name + view_id. Do not combine id with semantic selectors, and do not configure both view_name and view_id.",
	)
}
