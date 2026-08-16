package haproxy

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &serverDataSource{}
var _ datasource.DataSourceWithConfigure = &serverDataSource{}

type serverDataSource struct{ dataSourceClient }

func newServerDataSource() datasource.DataSource { return &serverDataSource{} }
func (d *serverDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_haproxy_server"
}
func (d *serverDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = serverDataSourceSchema()
}
func (d *serverDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	readDataSource(ctx, req, resp, "HAProxy Server", d.client.Haproxy().GetServer, serverAPIToModel, func(m *serverModel) string { return m.ID.ValueString() }, func(m *serverModel, id string) { m.ID = types.StringValue(id) })
}
