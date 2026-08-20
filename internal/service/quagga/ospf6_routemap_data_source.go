package quagga

import (
	"context"
	"github.com/biptec/opnsense-go/pkg/opnsense"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var _ datasource.DataSource = &ospf6RouteMapDataSource{}
var _ datasource.DataSourceWithConfigure = &ospf6RouteMapDataSource{}

type ospf6RouteMapDataSource struct{ client opnsense.Client }

func newOspf6RouteMapDataSource() datasource.DataSource { return &ospf6RouteMapDataSource{} }
func (d *ospf6RouteMapDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quagga_ospf6_routemap"
}
func (d *ospf6RouteMapDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = ospf6RouteMapDataSourceSchema()
}
func (d *ospf6RouteMapDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureQuaggaDataSource(req, resp)
}
func (d *ospf6RouteMapDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ospf6RouteMapModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	remote, err := d.client.Quagga().GetOSPF6RouteMap(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read OSPFv3 Route Map", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, ospf6RouteMapFromAPI(remote, data.ID.ValueString()))...)
}
