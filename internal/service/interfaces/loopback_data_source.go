package interfaces

import (
	"context"
	apiinterfaces "github.com/biptec/opnsense-go/pkg/interfaces"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var _ datasource.DataSource = &loopbackDataSource{}
var _ datasource.DataSourceWithConfigure = &loopbackDataSource{}

type loopbackDataSource struct{ interfaceDataSourceClient }

func newLoopbackDataSource() datasource.DataSource { return &loopbackDataSource{} }
func (d *loopbackDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_interfaces_loopback"
}
func (d *loopbackDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = loopbackDataSourceSchema()
}
func (d *loopbackDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	readInterfaceDataSource[loopbackResourceModel, apiinterfaces.Loopback](ctx, req, resp, "Loopback", d.client.Interfaces().GetLoopback, convertLoopbackStructToSchema, func(m *loopbackResourceModel) string { return m.Id.ValueString() }, func(m *loopbackResourceModel, id string) { m.Id = typesString(id) })
}
