package bind

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &viewDataSource{}
var _ datasource.DataSourceWithConfigure = &viewDataSource{}
var _ datasource.DataSourceWithConfigValidators = &viewDataSource{}

type viewDataSource struct{ dataSourceClient }

func newViewDataSource() datasource.DataSource { return &viewDataSource{} }
func (d *viewDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bind_view"
}
func (d *viewDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = viewDataSourceSchema()
}
func (d *viewDataSource) ConfigValidators(context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{datasourcevalidator.ExactlyOneOf(path.MatchRoot("id"), path.MatchRoot("name"))}
}
func (d *viewDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var id types.String
	var name types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("id"), &id)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("name"), &name)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resolvedID := id.ValueString()
	if id.IsNull() {
		result, err := d.client.Bind().SearchView(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Unable to Search BIND Views", err.Error())
			return
		}
		resolved, err := selectBindViewByName(result.Rows, name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Unable to Resolve BIND View", err.Error())
			return
		}
		resolvedID = resolved.ID
	}

	remote, err := d.client.Bind().GetView(ctx, resolvedID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read BIND View", err.Error())
		return
	}
	state, err := viewAPIToModel(remote)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Decode BIND View", err.Error())
		return
	}
	state.ID = types.StringValue(resolvedID)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
