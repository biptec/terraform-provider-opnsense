package caddy

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &handlerDataSource{}
var _ datasource.DataSourceWithConfigure = &handlerDataSource{}

type handlerDataSource struct{ dataSourceClient }

func newHandlerDataSource() datasource.DataSource { return &handlerDataSource{} }
func (d *handlerDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_caddy_handler"
}
func (d *handlerDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = handlerDataSourceSchema()
}
func (d *handlerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	readDataSource(ctx, req, resp, "Caddy Handler", d.client.Caddy().GetHandler, convertHandlerStructToSchema, func(m *handlerResourceModel) string { return m.ID.ValueString() }, func(m *handlerResourceModel, id string) { m.ID = types.StringValue(id) })
}
