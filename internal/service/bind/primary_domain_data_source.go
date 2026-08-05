package bind

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &primaryDomainDataSource{}
var _ datasource.DataSourceWithConfigure = &primaryDomainDataSource{}

type primaryDomainDataSource struct{ dataSourceClient }

func newPrimaryDomainDataSource() datasource.DataSource { return &primaryDomainDataSource{} }
func (d *primaryDomainDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bind_primary_domain"
}
func (d *primaryDomainDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = primaryDomainDataSourceSchema()
}
func (d *primaryDomainDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	readDataSource(ctx, req, resp, "BIND Primary Domain", d.client.Bind().GetPrimaryDomain, primaryDomainAPIToModel, func(m *primaryDomainResourceModel) string { return m.ID.ValueString() }, func(m *primaryDomainResourceModel, id string) { m.ID = types.StringValue(id) })
}
