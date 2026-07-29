package routing

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &gatewayDataSource{}
var _ datasource.DataSourceWithConfigure = &gatewayDataSource{}

type gatewayDataSource struct{ routingDataSourceClient }

func newGatewayDataSource() datasource.DataSource { return &gatewayDataSource{} }
func (d *gatewayDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_routing_gateway"
}
func (d *gatewayDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = gatewayDataSourceSchema()
}
func (d *gatewayDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	readRoutingDataSource(ctx, req, resp, "Routing Gateway", d.client.Routing().GetGateway,
		convertGatewayStructToSchema,
		func(model *gatewayResourceModel) string { return model.Id.ValueString() },
		func(model *gatewayResourceModel, id string) { model.Id = types.StringValue(id) },
	)
}
