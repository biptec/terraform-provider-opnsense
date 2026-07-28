package firewall

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &groupDataSource{}
var _ datasource.DataSourceWithConfigure = &groupDataSource{}

type groupDataSource struct{ firewallCoreDataSourceClient }

func newGroupDataSource() datasource.DataSource { return &groupDataSource{} }
func (d *groupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_group"
}
func (d *groupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = groupDataSourceSchema()
}
func (d *groupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	readFirewallCoreDataSource(ctx, req, resp, "Firewall Group", d.client.Firewall().GetGroup,
		convertGroupStructToSchema,
		func(model *groupResourceModel) string { return model.Id.ValueString() },
		func(model *groupResourceModel, id string) { model.Id = types.StringValue(id) },
	)
}
