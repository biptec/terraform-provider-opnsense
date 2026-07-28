package interfaces

import (
	"context"

	apiinterfaces "github.com/biptec/opnsense-go/pkg/interfaces"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var _ datasource.DataSource = &laggDataSource{}
var _ datasource.DataSourceWithConfigure = &laggDataSource{}

type laggDataSource struct{ interfaceDataSourceClient }

func newLaggDataSource() datasource.DataSource { return &laggDataSource{} }
func (d *laggDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_interfaces_lagg"
}
func (d *laggDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = laggDataSourceSchema()
}
func (d *laggDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	readInterfaceDataSource[laggResourceModel, apiinterfaces.Lagg](ctx, req, resp, "LAGG", d.client.Interfaces().GetLagg, convertLaggStructToSchema, func(m *laggResourceModel) string { return m.Id.ValueString() }, func(m *laggResourceModel, id string) { m.Id = typesString(id) })
}
