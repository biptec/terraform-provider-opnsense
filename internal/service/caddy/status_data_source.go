package caddy

import (
	"context"
	"fmt"
	"github.com/biptec/opnsense-go/pkg/api"
	"github.com/biptec/opnsense-go/pkg/opnsense"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type statusDataSourceModel struct {
	ID     types.String `tfsdk:"id"`
	Status types.String `tfsdk:"status"`
}
type statusDataSource struct{ client opnsense.Client }

var _ datasource.DataSource = &statusDataSource{}
var _ datasource.DataSourceWithConfigure = &statusDataSource{}

func newStatusDataSource() datasource.DataSource { return &statusDataSource{} }
func (d *statusDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_caddy_status"
}
func (d *statusDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{MarkdownDescription: "Reads the current Caddy service status.", Attributes: map[string]schema.Attribute{"id": schema.StringAttribute{Computed: true}, "status": schema.StringAttribute{Computed: true, MarkdownDescription: "Caddy service status such as `running`, `stopped`, or `disabled`."}}}
}
func (d *statusDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *statusDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	status, err := d.client.Caddy().ServiceStatus(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Caddy Status", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &statusDataSourceModel{ID: types.StringValue("caddy"), Status: types.StringValue(status.Status)})...)
}
