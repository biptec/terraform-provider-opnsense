package haproxy

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &frontendDataSource{}
var _ datasource.DataSourceWithConfigure = &frontendDataSource{}

type frontendDataSource struct{ dataSourceClient }

func newFrontendDataSource() datasource.DataSource { return &frontendDataSource{} }
func (d *frontendDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_haproxy_frontend"
}
func (d *frontendDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = frontendDataSourceSchema()
}
func (d *frontendDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	readDataSource(ctx, req, resp, "HAProxy Frontend", d.client.Haproxy().GetFrontend, frontendAPIToModel, func(m *frontendModel) string { return m.ID.ValueString() }, func(m *frontendModel, id string) { m.ID = types.StringValue(id) })
}
