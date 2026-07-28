package interfaces

import (
	"context"

	apiinterfaces "github.com/biptec/opnsense-go/pkg/interfaces"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var _ datasource.DataSource = &vxlanDataSource{}
var _ datasource.DataSourceWithConfigure = &vxlanDataSource{}

type vxlanDataSource struct{ interfaceDataSourceClient }

func newVxlanDataSource() datasource.DataSource { return &vxlanDataSource{} }
func (d *vxlanDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_interfaces_vxlan"
}
func (d *vxlanDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = vxlanDataSourceSchema()
}
func (d *vxlanDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	readInterfaceDataSource[vxlanResourceModel, apiinterfaces.Vxlan](ctx, req, resp, "VXLAN", d.client.Interfaces().GetVxlan, convertVxlanStructToSchema, func(m *vxlanResourceModel) string { return m.Id.ValueString() }, func(m *vxlanResourceModel, id string) { m.Id = typesString(id) })
}
