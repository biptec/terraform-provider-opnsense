package bind

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &forwardDomainDataSource{}
var _ datasource.DataSourceWithConfigure = &forwardDomainDataSource{}

type forwardDomainDataSource struct{ dataSourceClient }

func newForwardDomainDataSource() datasource.DataSource { return &forwardDomainDataSource{} }
func (d *forwardDomainDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bind_forward_domain"
}
func (d *forwardDomainDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = forwardDomainDataSourceSchema()
}
func (d *forwardDomainDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	readDataSource(ctx, req, resp, "BIND ForwardDomain", d.client.Bind().GetForwardDomain, forwardDomainAPIToModel, func(m *forwardDomainResourceModel) string { return m.ID.ValueString() }, func(m *forwardDomainResourceModel, id string) { m.ID = types.StringValue(id) })
}
