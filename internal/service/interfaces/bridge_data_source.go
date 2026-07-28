package interfaces

import (
	"context"

	apiinterfaces "github.com/biptec/opnsense-go/pkg/interfaces"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var _ datasource.DataSource = &bridgeDataSource{}
var _ datasource.DataSourceWithConfigure = &bridgeDataSource{}

type bridgeDataSource struct{ interfaceDataSourceClient }

func newBridgeDataSource() datasource.DataSource { return &bridgeDataSource{} }
func (d *bridgeDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_interfaces_bridge"
}
func (d *bridgeDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = bridgeDataSourceSchema()
}
func (d *bridgeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	readInterfaceDataSource[bridgeResourceModel, apiinterfaces.Bridge](ctx, req, resp, "Bridge", d.client.Interfaces().GetBridge, convertBridgeStructToSchema, func(m *bridgeResourceModel) string { return m.Id.ValueString() }, func(m *bridgeResourceModel, id string) { m.Id = typesString(id) })
}
