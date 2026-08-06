package unbound

import (
	"context"
	"fmt"

	"github.com/biptec/opnsense-go/pkg/api"
	"github.com/biptec/opnsense-go/pkg/opnsense"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var _ datasource.DataSource = &serviceDataSource{}
var _ datasource.DataSourceWithConfigure = &serviceDataSource{}

type serviceDataSource struct {
	client opnsense.Client
}

func newServiceDataSource() datasource.DataSource {
	return &serviceDataSource{}
}

func (d *serviceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_unbound_service"
}
func (d *serviceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = serviceDataSourceSchema()
}

func (d *serviceDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *serviceDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	remote, err := d.client.Unbound().SettingsGet(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Unbound Service", err.Error())
		return
	}
	if remote == nil {
		resp.Diagnostics.AddError("Unable to Read Unbound Service", "Unbound settings API returned an empty response")
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, serviceAPIToModel(remote.Unbound.General.Enabled))...)
}
