package interfaces

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/biptec/opnsense-go/pkg/api"
	"github.com/biptec/opnsense-go/pkg/opnsense"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &overviewDetailsDataSource{}
var _ datasource.DataSourceWithConfigure = &overviewDetailsDataSource{}

type overviewDetailsDataSource struct{ client opnsense.Client }
type overviewDetailsModel struct {
	Interface   types.String `tfsdk:"interface"`
	DetailsJSON types.String `tfsdk:"details_json"`
}

func newOverviewDetailsDataSource() datasource.DataSource { return &overviewDetailsDataSource{} }
func (d *overviewDetailsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_interfaces_details"
}
func (d *overviewDetailsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{MarkdownDescription: "Reads the complete OPNsense detail payload for one operating-system interface.", Attributes: map[string]schema.Attribute{
		"interface":    schema.StringAttribute{Required: true, MarkdownDescription: "Operating-system interface device accepted by the OPNsense overview API, for example `vtnet0`."},
		"details_json": schema.StringAttribute{Computed: true, MarkdownDescription: "Complete API detail payload encoded as JSON. The payload varies by interface type and OPNsense version."},
	}}
}
func (d *overviewDetailsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *overviewDetailsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data overviewDetailsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := d.client.Interfaces().OverviewGetInterface(ctx, data.Interface.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Interface Details", err.Error())
		return
	}
	if status, ok := result.Message.(string); ok && status == "failed" {
		resp.Diagnostics.AddError(
			"Unable to Read Interface Details",
			fmt.Sprintf("OPNsense did not find operating-system interface %q.", data.Interface.ValueString()),
		)
		return
	}
	encoded, err := json.Marshal(result.Message)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Encode Interface Details", err.Error())
		return
	}
	data.DetailsJSON = types.StringValue(string(encoded))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
