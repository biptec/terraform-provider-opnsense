package interfaces

import (
	"context"
	apiinterfaces "github.com/biptec/opnsense-go/pkg/interfaces"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var _ datasource.DataSource = &neighborDataSource{}
var _ datasource.DataSourceWithConfigure = &neighborDataSource{}

type neighborDataSource struct{ interfaceDataSourceClient }

func newNeighborDataSource() datasource.DataSource { return &neighborDataSource{} }
func (d *neighborDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_interfaces_neighbor"
}
func (d *neighborDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = neighborDataSourceSchema()
}
func (d *neighborDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	readInterfaceDataSource[neighborResourceModel, apiinterfaces.Neighbor](ctx, req, resp, "Neighbor", d.client.Interfaces().GetNeighbor, convertNeighborStructToSchema, func(m *neighborResourceModel) string { return m.Id.ValueString() }, func(m *neighborResourceModel, id string) { m.Id = typesString(id) })
}
