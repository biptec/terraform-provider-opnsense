package bind

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &viewDataSource{}
var _ datasource.DataSourceWithConfigure = &viewDataSource{}

type viewDataSource struct{ dataSourceClient }

func newViewDataSource() datasource.DataSource { return &viewDataSource{} }
func (d *viewDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bind_view"
}
func (d *viewDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = viewDataSourceSchema()
}
func (d *viewDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	readDataSource(ctx, req, resp, "BIND View", d.client.Bind().GetView, viewAPIToModel, func(m *viewResourceModel) string { return m.ID.ValueString() }, func(m *viewResourceModel, id string) { m.ID = types.StringValue(id) })
}
