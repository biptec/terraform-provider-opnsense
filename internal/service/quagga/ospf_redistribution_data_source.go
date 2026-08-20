package quagga

import (
	"context"
	"github.com/biptec/opnsense-go/pkg/opnsense"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var _ datasource.DataSource = &ospfRedistributionDataSource{}
var _ datasource.DataSourceWithConfigure = &ospfRedistributionDataSource{}

type ospfRedistributionDataSource struct{ client opnsense.Client }

func newOspfRedistributionDataSource() datasource.DataSource { return &ospfRedistributionDataSource{} }
func (d *ospfRedistributionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quagga_ospf_redistribution"
}
func (d *ospfRedistributionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = ospfRedistributionDataSourceSchema()
}
func (d *ospfRedistributionDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureQuaggaDataSource(req, resp)
}
func (d *ospfRedistributionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ospfRedistributionModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	remote, err := d.client.Quagga().GetOSPFRedistribution(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read OSPFv2 Redistribution", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, ospfRedistributionFromAPI(remote, data.ID.ValueString()))...)
}
