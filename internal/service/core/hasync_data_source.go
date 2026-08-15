package core

import (
	"context"
	"fmt"

	"github.com/biptec/opnsense-go/pkg/api"
	"github.com/biptec/opnsense-go/pkg/opnsense"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var _ datasource.DataSource = &hasyncDataSource{}
var _ datasource.DataSourceWithConfigure = &hasyncDataSource{}

type hasyncDataSource struct{ client opnsense.Client }

func newHasyncDataSource() datasource.DataSource { return &hasyncDataSource{} }
func (d *hasyncDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_core_hasync"
}
func (d *hasyncDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = hasyncDataSourceSchema()
}
func (d *hasyncDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected *api.Client, got %T.", req.ProviderData))
		return
	}
	d.client = opnsense.NewClient(c)
}
func (d *hasyncDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	remote, err := d.client.Core().HasyncGet(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read OPNsense HA Settings", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, hasyncAPIToModel(&remote.Hasync))...)
}
