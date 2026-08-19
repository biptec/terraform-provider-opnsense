package quagga

import (
	"context"
	"github.com/biptec/opnsense-go/pkg/opnsense"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var _ datasource.DataSource = &ospfPrefixListDataSource{}
var _ datasource.DataSourceWithConfigure = &ospfPrefixListDataSource{}

type ospfPrefixListDataSource struct{ client opnsense.Client }

func newOspfPrefixListDataSource() datasource.DataSource { return &ospfPrefixListDataSource{} }
func (d *ospfPrefixListDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quagga_ospf_prefixlist"
}
func (d *ospfPrefixListDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = ospfPrefixListDataSourceSchema()
}
func (d *ospfPrefixListDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureQuaggaDataSource(req, resp)
}
func (d *ospfPrefixListDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ospfPrefixListModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	remote, err := d.client.Quagga().GetOSPFPrefixList(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read OSPFv2 Prefix List", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, ospfPrefixListFromAPI(remote, data.ID.ValueString()))...)
}
