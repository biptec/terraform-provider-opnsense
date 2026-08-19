package quagga

import (
	"context"
	"github.com/biptec/opnsense-go/pkg/opnsense"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var _ datasource.DataSource = &ospfRouteMapDataSource{}
var _ datasource.DataSourceWithConfigure = &ospfRouteMapDataSource{}

type ospfRouteMapDataSource struct{ client opnsense.Client }

func newOspfRouteMapDataSource() datasource.DataSource { return &ospfRouteMapDataSource{} }
func (d *ospfRouteMapDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quagga_ospf_routemap"
}
func (d *ospfRouteMapDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = ospfRouteMapDataSourceSchema()
}
func (d *ospfRouteMapDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureQuaggaDataSource(req, resp)
}
func (d *ospfRouteMapDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ospfRouteMapModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	remote, err := d.client.Quagga().GetOSPFRouteMap(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read OSPFv2 Route Map", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, ospfRouteMapFromAPI(remote, data.ID.ValueString()))...)
}
