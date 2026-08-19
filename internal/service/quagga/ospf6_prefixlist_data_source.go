package quagga

import (
	"context"
	"github.com/biptec/opnsense-go/pkg/opnsense"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var _ datasource.DataSource = &ospf6PrefixListDataSource{}
var _ datasource.DataSourceWithConfigure = &ospf6PrefixListDataSource{}

type ospf6PrefixListDataSource struct{ client opnsense.Client }

func newOspf6PrefixListDataSource() datasource.DataSource { return &ospf6PrefixListDataSource{} }
func (d *ospf6PrefixListDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quagga_ospf6_prefixlist"
}
func (d *ospf6PrefixListDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = ospf6PrefixListDataSourceSchema()
}
func (d *ospf6PrefixListDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureQuaggaDataSource(req, resp)
}
func (d *ospf6PrefixListDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ospf6PrefixListModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	remote, err := d.client.Quagga().GetOSPF6PrefixList(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read OSPFv3 Prefix List", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, ospf6PrefixListFromAPI(remote, data.ID.ValueString()))...)
}
