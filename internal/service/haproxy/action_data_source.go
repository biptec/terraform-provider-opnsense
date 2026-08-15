package haproxy

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &actionDataSource{}
var _ datasource.DataSourceWithConfigure = &actionDataSource{}

type actionDataSource struct{ dataSourceClient }

func newActionDataSource() datasource.DataSource { return &actionDataSource{} }
func (d *actionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_haproxy_action"
}
func (d *actionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = actionDataSourceSchema()
}
func (d *actionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	readDataSource(ctx, req, resp, "HAProxy Action", d.client.Haproxy().GetAction, actionAPIToModel, func(m *actionModel) string { return m.ID.ValueString() }, func(m *actionModel, id string) { m.ID = types.StringValue(id) })
}
