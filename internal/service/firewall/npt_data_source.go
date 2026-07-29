package firewall

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &nptDataSource{}
var _ datasource.DataSourceWithConfigure = &nptDataSource{}

type nptDataSource struct{ firewallCoreDataSourceClient }

func newNptDataSource() datasource.DataSource { return &nptDataSource{} }
func (d *nptDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_npt"
}
func (d *nptDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = nptDataSourceSchema()
}
func (d *nptDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	readFirewallCoreDataSource(ctx, req, resp, "Firewall NPT", d.client.Firewall().GetNpt,
		convertNptStructToSchema,
		func(model *nptResourceModel) string { return model.Id.ValueString() },
		func(model *nptResourceModel, id string) { model.Id = types.StringValue(id) },
	)
}
