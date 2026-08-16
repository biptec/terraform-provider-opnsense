package haproxy

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &aclDataSource{}
var _ datasource.DataSourceWithConfigure = &aclDataSource{}

type aclDataSource struct{ dataSourceClient }

func newACLDataSource() datasource.DataSource { return &aclDataSource{} }
func (d *aclDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_haproxy_acl"
}
func (d *aclDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = aclDataSourceSchema()
}
func (d *aclDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	readDataSource(ctx, req, resp, "HAProxy ACL", d.client.Haproxy().GetACL, aclAPIToModel, func(m *aclModel) string { return m.ID.ValueString() }, func(m *aclModel, id string) { m.ID = types.StringValue(id) })
}
