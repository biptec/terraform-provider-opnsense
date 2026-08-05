package bind

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &secondaryDomainDataSource{}
var _ datasource.DataSourceWithConfigure = &secondaryDomainDataSource{}

type secondaryDomainDataSource struct{ dataSourceClient }

func newSecondaryDomainDataSource() datasource.DataSource { return &secondaryDomainDataSource{} }
func (d *secondaryDomainDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bind_secondary_domain"
}
func (d *secondaryDomainDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = secondaryDomainDataSourceSchema()
}
func (d *secondaryDomainDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	readDataSource(ctx, req, resp, "BIND SecondaryDomain", d.client.Bind().GetSecondaryDomain, secondaryDomainAPIToModel, func(m *secondaryDomainResourceModel) string { return m.ID.ValueString() }, func(m *secondaryDomainResourceModel, id string) { m.ID = types.StringValue(id) })
}
