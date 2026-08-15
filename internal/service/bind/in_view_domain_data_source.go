package bind

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &inViewDomainDataSource{}
var _ datasource.DataSourceWithConfigure = &inViewDomainDataSource{}

type inViewDomainDataSource struct{ dataSourceClient }

func newInViewDomainDataSource() datasource.DataSource { return &inViewDomainDataSource{} }
func (d *inViewDomainDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bind_in_view_domain"
}
func (d *inViewDomainDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = inViewDomainDataSourceSchema()
}
func (d *inViewDomainDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	readDataSource(ctx, req, resp, "BIND InViewDomain", d.client.Bind().GetInViewDomain, inViewDomainAPIToModel, func(m *inViewDomainResourceModel) string { return m.ID.ValueString() }, func(m *inViewDomainResourceModel, id string) { m.ID = types.StringValue(id) })
}
