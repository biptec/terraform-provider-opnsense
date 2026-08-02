package caddy

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var _ datasource.DataSource = &settingsDataSource{}
var _ datasource.DataSourceWithConfigure = &settingsDataSource{}

type settingsDataSource struct{ dataSourceClient }

func newSettingsDataSource() datasource.DataSource { return &settingsDataSource{} }
func (d *settingsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_caddy_settings"
}
func (d *settingsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = settingsDataSourceSchema()
}
func (d *settingsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	remote, err := d.client.Caddy().SettingsGet(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Caddy Settings", err.Error())
		return
	}
	state, err := settingsStructToSchema(remote)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Decode Caddy Settings", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
