package bind

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &recordDataSource{}
var _ datasource.DataSourceWithConfigure = &recordDataSource{}

type recordDataSource struct{ dataSourceClient }

func newRecordDataSource() datasource.DataSource { return &recordDataSource{} }
func (d *recordDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bind_record"
}
func (d *recordDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = recordDataSourceSchema()
}
func (d *recordDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	readDataSource(ctx, req, resp, "BIND Record", d.client.Bind().GetRecord, recordAPIToModel, func(m *recordResourceModel) string { return m.ID.ValueString() }, func(m *recordResourceModel, id string) { m.ID = types.StringValue(id) })
}
