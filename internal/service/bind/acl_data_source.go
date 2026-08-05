package bind

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &aclDataSource{}
var _ datasource.DataSourceWithConfigure = &aclDataSource{}

type aclDataSource struct{ dataSourceClient }

func newAclDataSource() datasource.DataSource { return &aclDataSource{} }
func (d *aclDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bind_acl"
}
func (d *aclDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = aclDataSourceSchema()
}
func (d *aclDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	readDataSource(ctx, req, resp, "BIND ACL", d.client.Bind().GetAcl, aclAPIToModel, func(m *aclResourceModel) string { return m.ID.ValueString() }, func(m *aclResourceModel, id string) { m.ID = types.StringValue(id) })
}
