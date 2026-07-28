package interfaces

import (
	"context"
	"fmt"

	"github.com/biptec/opnsense-go/pkg/api"
	"github.com/biptec/opnsense-go/pkg/opnsense"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var _ datasource.DataSource = &interfaceSettingsDataSource{}
var _ datasource.DataSourceWithConfigure = &interfaceSettingsDataSource{}

type interfaceSettingsDataSource struct{ client opnsense.Client }

func newInterfaceSettingsDataSource() datasource.DataSource { return &interfaceSettingsDataSource{} }
func (d *interfaceSettingsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_interfaces_settings"
}
func (d *interfaceSettingsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = interfaceSettingsDataSourceSchema()
}
func (d *interfaceSettingsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected *api.Client, got %T.", req.ProviderData))
		return
	}
	d.client = opnsense.NewClient(client)
}
func (d *interfaceSettingsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	result, err := d.client.Interfaces().SettingsGet(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Interface Settings", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, convertInterfaceSettingsStructToSchema(&result.Settings))...)
}
