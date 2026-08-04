package caddy

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &headerDataSource{}
var _ datasource.DataSourceWithConfigure = &headerDataSource{}

type headerDataSource struct{ dataSourceClient }

func newHeaderDataSource() datasource.DataSource { return &headerDataSource{} }

func (d *headerDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_caddy_header"
}

func (d *headerDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = headerDataSourceSchema()
}

func (d *headerDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	readDataSource(
		ctx,
		req,
		resp,
		"Caddy Header",
		d.client.Caddy().GetHeader,
		convertHeaderStructToSchema,
		func(m *headerResourceModel) string { return m.ID.ValueString() },
		func(m *headerResourceModel, id string) { m.ID = types.StringValue(id) },
	)
}
