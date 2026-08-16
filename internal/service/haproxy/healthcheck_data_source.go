package haproxy

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &healthcheckDataSource{}
var _ datasource.DataSourceWithConfigure = &healthcheckDataSource{}

type healthcheckDataSource struct{ dataSourceClient }

func newHealthcheckDataSource() datasource.DataSource { return &healthcheckDataSource{} }
func (d *healthcheckDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_haproxy_healthcheck"
}
func (d *healthcheckDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = healthcheckDataSourceSchema()
}
func (d *healthcheckDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	readDataSource(ctx, req, resp, "HAProxy Healthcheck", d.client.Haproxy().GetHealthcheck, healthcheckAPIToModel, func(m *healthcheckModel) string { return m.ID.ValueString() }, func(m *healthcheckModel, id string) { m.ID = types.StringValue(id) })
}
