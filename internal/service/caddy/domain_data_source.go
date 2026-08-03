package caddy

import (
	"context"
	caddy "github.com/biptec/opnsense-go/pkg/caddy"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &domainDataSource{}
var _ datasource.DataSourceWithConfigure = &domainDataSource{}

type domainDataSource struct{ dataSourceClient }

func newDomainDataSource() datasource.DataSource { return &domainDataSource{} }
func (d *domainDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_caddy_domain"
}
func (d *domainDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = domainDataSourceSchema()
}
func (d *domainDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	readDataSource(ctx, req, resp, "Caddy Domain", d.client.Caddy().GetDomain, func(v *caddy.Domain) (*domainResourceModel, error) { return domainStructToSchema(v, nil) }, func(m *domainResourceModel) string { return m.ID.ValueString() }, func(m *domainResourceModel, id string) { m.ID = types.StringValue(id) })
}
