package quagga

import (
	"context"
	"github.com/biptec/opnsense-go/pkg/opnsense"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var _ datasource.DataSource = &ospf6RedistributionDataSource{}
var _ datasource.DataSourceWithConfigure = &ospf6RedistributionDataSource{}

type ospf6RedistributionDataSource struct{ client opnsense.Client }

func newOspf6RedistributionDataSource() datasource.DataSource {
	return &ospf6RedistributionDataSource{}
}
func (d *ospf6RedistributionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quagga_ospf6_redistribution"
}
func (d *ospf6RedistributionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = ospf6RedistributionDataSourceSchema()
}
func (d *ospf6RedistributionDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureQuaggaDataSource(req, resp)
}
func (d *ospf6RedistributionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ospf6RedistributionModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	remote, err := d.client.Quagga().GetOSPF6Redistribution(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read OSPFv3 Redistribution", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, ospf6RedistributionFromAPI(remote, data.ID.ValueString()))...)
}
