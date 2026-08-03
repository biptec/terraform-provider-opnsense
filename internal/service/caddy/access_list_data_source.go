package caddy

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &accessListDataSource{}
var _ datasource.DataSourceWithConfigure = &accessListDataSource{}

type accessListDataSource struct{ dataSourceClient }

func newAccessListDataSource() datasource.DataSource { return &accessListDataSource{} }
func (d *accessListDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_caddy_access_list"
}
func (d *accessListDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = accessListDataSourceSchema()
}
func (d *accessListDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	readDataSource(ctx, req, resp, "Caddy Access List", d.client.Caddy().GetAccessList, convertAccessListStructToSchema, func(m *accessListResourceModel) string { return m.ID.ValueString() }, func(m *accessListResourceModel, id string) { m.ID = types.StringValue(id) })
}
