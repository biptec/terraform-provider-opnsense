package bind

import (
	"context"
	apibind "github.com/biptec/opnsense-go/pkg/bind"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &primaryDomainDataSource{}
var _ datasource.DataSourceWithConfigure = &primaryDomainDataSource{}
var _ datasource.DataSourceWithConfigValidators = &primaryDomainDataSource{}

type primaryDomainDataSource struct{ dataSourceClient }

func newPrimaryDomainDataSource() datasource.DataSource { return &primaryDomainDataSource{} }
func (d *primaryDomainDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bind_primary_domain"
}
func (d *primaryDomainDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = primaryDomainDataSourceSchema()
}
func (d *primaryDomainDataSource) ConfigValidators(context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{primaryDomainLookupConfigValidator{}}
}
func (d *primaryDomainDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var id, domainName, viewName, viewID types.String
	for attribute, target := range map[string]*types.String{
		"id": &id, "domain_name": &domainName, "view_name": &viewName, "view_id": &viewID,
	} {
		resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root(attribute), target)...)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	resolvedID := id.ValueString()
	resolvedViewName := ""
	var remote *apibind.PrimaryDomain
	var err error

	if !id.IsNull() {
		remote, err = d.client.Bind().GetPrimaryDomain(ctx, resolvedID)
		if err != nil {
			resp.Diagnostics.AddError("Unable to Read BIND Primary Domain", err.Error())
			return
		}
		viewID = types.StringValue(remote.View.String())
		view, viewErr := d.client.Bind().GetView(ctx, viewID.ValueString())
		if viewErr != nil {
			resp.Diagnostics.AddError("Unable to Read BIND View", viewErr.Error())
			return
		}
		resolvedViewName = normalizeBindViewName(view.Name)
	} else {
		if !viewName.IsNull() {
			views, searchErr := d.client.Bind().SearchView(ctx)
			if searchErr != nil {
				resp.Diagnostics.AddError("Unable to Search BIND Views", searchErr.Error())
				return
			}
			resolvedView, resolveErr := selectBindViewByName(views.Rows, viewName.ValueString())
			if resolveErr != nil {
				resp.Diagnostics.AddError("Unable to Resolve BIND View", resolveErr.Error())
				return
			}
			viewID = types.StringValue(resolvedView.ID)
			resolvedViewName = resolvedView.Name
		} else {
			view, viewErr := d.client.Bind().GetView(ctx, viewID.ValueString())
			if viewErr != nil {
				resp.Diagnostics.AddError("Unable to Read BIND View", viewErr.Error())
				return
			}
			resolvedViewName = normalizeBindViewName(view.Name)
		}

		zones, searchErr := d.client.Bind().SearchPrimaryDomain(ctx)
		if searchErr != nil {
			resp.Diagnostics.AddError("Unable to Search BIND Primary Domains", searchErr.Error())
			return
		}
		resolved, resolveErr := selectPrimaryDomainInView(
			ctx, zones.Rows, domainName.ValueString(), viewID.ValueString(), resolvedViewName,
			d.client.Bind().GetPrimaryDomain,
		)
		if resolveErr != nil {
			resp.Diagnostics.AddError("Unable to Resolve BIND Primary Domain", resolveErr.Error())
			return
		}
		resolvedID = resolved.ID
		remote = resolved.Domain
	}

	resourceState, err := primaryDomainAPIToModel(remote)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Decode BIND Primary Domain", err.Error())
		return
	}
	state := primaryDomainDataSourceModelFromResource(resourceState, resolvedID, resolvedViewName)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
