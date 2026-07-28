package routing

import (
	"context"
	"fmt"

	"github.com/biptec/opnsense-go/pkg/api"
	"github.com/biptec/opnsense-go/pkg/opnsense"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &gatewayStatusDataSource{}
var _ datasource.DataSourceWithConfigure = &gatewayStatusDataSource{}

type gatewayStatusDataSource struct{ client opnsense.Client }

type gatewayStatusDataSourceModel struct {
	Status types.String `tfsdk:"status"`
	Items  types.List   `tfsdk:"items"`
}

type gatewayStatusItemModel struct {
	Name              types.String `tfsdk:"name"`
	Address           types.String `tfsdk:"address"`
	Monitor           types.String `tfsdk:"monitor"`
	Delay             types.String `tfsdk:"delay"`
	StandardDeviation types.String `tfsdk:"standard_deviation"`
	Loss              types.String `tfsdk:"loss"`
	Status            types.String `tfsdk:"status"`
	StatusTranslated  types.String `tfsdk:"status_translated"`
}

var gatewayStatusItemAttrTypes = map[string]attr.Type{
	"name":               types.StringType,
	"address":            types.StringType,
	"monitor":            types.StringType,
	"delay":              types.StringType,
	"standard_deviation": types.StringType,
	"loss":               types.StringType,
	"status":             types.StringType,
	"status_translated":  types.StringType,
}

func newGatewayStatusDataSource() datasource.DataSource { return &gatewayStatusDataSource{} }
func (d *gatewayStatusDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_routing_gateway_status"
}
func (d *gatewayStatusDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads live status for all OPNsense gateways.",
		Attributes: map[string]schema.Attribute{
			"status": schema.StringAttribute{Computed: true},
			"items": schema.ListNestedAttribute{Computed: true, NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"name":               schema.StringAttribute{Computed: true},
					"address":            schema.StringAttribute{Computed: true},
					"monitor":            schema.StringAttribute{Computed: true},
					"delay":              schema.StringAttribute{Computed: true},
					"standard_deviation": schema.StringAttribute{Computed: true},
					"loss":               schema.StringAttribute{Computed: true},
					"status":             schema.StringAttribute{Computed: true},
					"status_translated":  schema.StringAttribute{Computed: true},
				},
			}},
		},
	}
}

func (d *gatewayStatusDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	apiClient, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected *api.Client, got %T.", req.ProviderData))
		return
	}
	d.client = opnsense.NewClient(apiClient)
}

func (d *gatewayStatusDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	result, err := d.client.Routing().GatewayStatusGet(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Gateway Status", err.Error())
		return
	}
	items := make([]gatewayStatusItemModel, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, gatewayStatusItemModel{
			Name:              types.StringValue(item.Name),
			Address:           types.StringValue(item.Address),
			Monitor:           types.StringValue(item.Monitor),
			Delay:             types.StringValue(item.Delay),
			StandardDeviation: types.StringValue(item.StandardDeviation),
			Loss:              types.StringValue(item.Loss),
			Status:            types.StringValue(item.Status),
			StatusTranslated:  types.StringValue(item.StatusTranslated),
		})
	}
	list, diagnostics := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: gatewayStatusItemAttrTypes}, items)
	resp.Diagnostics.Append(diagnostics...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &gatewayStatusDataSourceModel{
		Status: types.StringValue(result.Status),
		Items:  list,
	})...)
}
