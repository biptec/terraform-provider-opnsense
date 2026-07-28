package interfaces

import (
	"context"
	"fmt"

	"github.com/biptec/opnsense-go/pkg/api"
	"github.com/biptec/opnsense-go/pkg/opnsense"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var _ datasource.DataSource = &assignmentDataSource{}
var _ datasource.DataSourceWithConfigure = &assignmentDataSource{}

type assignmentDataSource struct{ client opnsense.Client }

func newAssignmentDataSource() datasource.DataSource { return &assignmentDataSource{} }

func (d *assignmentDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_interfaces_assignment"
}

func (d *assignmentDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = assignmentDataSourceSchema()
}

func (d *assignmentDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *assignmentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data assignmentDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := data.Id.ValueString()
	assignment, err := d.client.Interfaces().GetAssignment(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Interface Assignment", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, convertAssignmentStructToDataSourceSchema(assignment, id))...)
}
