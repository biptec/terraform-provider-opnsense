package interfaces

import (
	"context"
	apiinterfaces "github.com/biptec/opnsense-go/pkg/interfaces"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var _ datasource.DataSource = &gifDataSource{}
var _ datasource.DataSourceWithConfigure = &gifDataSource{}

type gifDataSource struct{ interfaceDataSourceClient }

func newGifDataSource() datasource.DataSource { return &gifDataSource{} }
func (d *gifDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_interfaces_gif"
}
func (d *gifDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = gifDataSourceSchema()
}
func (d *gifDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	readInterfaceDataSource[gifResourceModel, apiinterfaces.Gif](ctx, req, resp, "GIF", d.client.Interfaces().GetGif, convertGifStructToSchema, func(m *gifResourceModel) string { return m.Id.ValueString() }, func(m *gifResourceModel, id string) { m.Id = typesString(id) })
}
