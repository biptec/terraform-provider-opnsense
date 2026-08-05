package bind

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &tsigKeyDataSource{}
var _ datasource.DataSourceWithConfigure = &tsigKeyDataSource{}

type tsigKeyDataSource struct{ dataSourceClient }

func newTsigKeyDataSource() datasource.DataSource { return &tsigKeyDataSource{} }
func (d *tsigKeyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bind_tsig_key"
}
func (d *tsigKeyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = tsigKeyDataSourceSchema()
}
func (d *tsigKeyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	readDataSource(ctx, req, resp, "BIND TSIG Key", d.client.Bind().GetTsigKey, tsigKeyAPIToModel, func(m *tsigKeyResourceModel) string { return m.ID.ValueString() }, func(m *tsigKeyResourceModel, id string) { m.ID = types.StringValue(id) })
}
