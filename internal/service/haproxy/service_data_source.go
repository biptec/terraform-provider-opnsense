package haproxy

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type statusModel struct {
	Status types.String `tfsdk:"status"`
	ID     types.String `tfsdk:"id"`
}
type configtestModel struct {
	Result types.String `tfsdk:"result"`
	ID     types.String `tfsdk:"id"`
}

type statusDataSource struct{ dataSourceClient }

var _ datasource.DataSource = &statusDataSource{}
var _ datasource.DataSourceWithConfigure = &statusDataSource{}

func newStatusDataSource() datasource.DataSource { return &statusDataSource{} }
func (d *statusDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_haproxy_status"
}
func (d *statusDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dschema.Schema{MarkdownDescription: "Reads HAProxy service status.", Attributes: map[string]dschema.Attribute{"status": dschema.StringAttribute{Computed: true}, "id": dschema.StringAttribute{Computed: true}}}
}
func (d *statusDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	result, err := d.client.Haproxy().ServiceStatus(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read HAProxy Status", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &statusModel{Status: types.StringValue(result.Status), ID: types.StringValue("haproxy_status")})...)
}

type configtestDataSource struct{ dataSourceClient }

var _ datasource.DataSource = &configtestDataSource{}
var _ datasource.DataSourceWithConfigure = &configtestDataSource{}

func newConfigtestDataSource() datasource.DataSource { return &configtestDataSource{} }
func (d *configtestDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_haproxy_configtest"
}
func (d *configtestDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dschema.Schema{MarkdownDescription: "Runs the OPNsense HAProxy configuration test and returns its result.", Attributes: map[string]dschema.Attribute{"result": dschema.StringAttribute{Computed: true}, "id": dschema.StringAttribute{Computed: true}}}
}
func (d *configtestDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	result, err := d.client.Haproxy().ServiceConfigtest(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Test HAProxy Configuration", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &configtestModel{Result: types.StringValue(result.Result), ID: types.StringValue("haproxy_configtest")})...)
}
