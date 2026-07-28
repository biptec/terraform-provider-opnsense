package interfaces

import (
	"context"
	apiinterfaces "github.com/biptec/opnsense-go/pkg/interfaces"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var _ datasource.DataSource = &greDataSource{}
var _ datasource.DataSourceWithConfigure = &greDataSource{}

type greDataSource struct{ interfaceDataSourceClient }

func newGreDataSource() datasource.DataSource { return &greDataSource{} }
func (d *greDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_interfaces_gre"
}
func (d *greDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = greDataSourceSchema()
}
func (d *greDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	readInterfaceDataSource[greResourceModel, apiinterfaces.Gre](ctx, req, resp, "GRE", d.client.Interfaces().GetGre, convertGreStructToSchema, func(m *greResourceModel) string { return m.Id.ValueString() }, func(m *greResourceModel, id string) { m.Id = typesString(id) })
}
