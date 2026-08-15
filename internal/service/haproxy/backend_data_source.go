package haproxy

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &backendDataSource{}
var _ datasource.DataSourceWithConfigure = &backendDataSource{}

type backendDataSource struct{ dataSourceClient }

func newBackendDataSource() datasource.DataSource { return &backendDataSource{} }
func (d *backendDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_haproxy_backend"
}
func (d *backendDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = backendDataSourceSchema()
}
func (d *backendDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	readDataSource(ctx, req, resp, "HAProxy Backend", d.client.Haproxy().GetBackend, backendAPIToModel, func(m *backendModel) string { return m.ID.ValueString() }, func(m *backendModel, id string) { m.ID = types.StringValue(id) })
}
