package routing

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &gatewayGroupDataSource{}
var _ datasource.DataSourceWithConfigure = &gatewayGroupDataSource{}

type gatewayGroupDataSource struct{ routingDataSourceClient }

func newGatewayGroupDataSource() datasource.DataSource { return &gatewayGroupDataSource{} }
func (d *gatewayGroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_routing_gateway_group"
}
func (d *gatewayGroupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = gatewayGroupDataSourceSchema()
}
func (d *gatewayGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	readRoutingDataSource(ctx, req, resp, "Routing Gateway Group", d.client.Routing().GetGatewayGroup,
		convertGatewayGroupStructToSchema,
		func(model *gatewayGroupResourceModel) string { return model.Id.ValueString() },
		func(model *gatewayGroupResourceModel, id string) { model.Id = types.StringValue(id) },
	)
}
