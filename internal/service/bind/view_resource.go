package bind

import (
	"context"
	apibind "github.com/biptec/opnsense-go/pkg/bind"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &viewResource{}
var _ resource.ResourceWithConfigure = &viewResource{}
var _ resource.ResourceWithImportState = &viewResource{}
var _ resource.ResourceWithConfigValidators = &viewResource{}

type viewResource struct{ resourceClient }

func newViewResource() resource.Resource { return &viewResource{} }
func (r *viewResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bind_view"
}
func (r *viewResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = viewResourceSchema()
}
func (r *viewResource) ConfigValidators(context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{viewTSIGMatchConfigValidator{}}
}
func (r *viewResource) operations() crudOperations[viewResourceModel, apibind.View] {
	c := r.client.Bind()
	return crudOperations[viewResourceModel, apibind.View]{Name: "BIND View", Convert: viewModelToAPI, Expand: viewAPIToModel, Add: c.AddView, Get: c.GetView, Update: c.UpdateView, Delete: c.DeleteView, GetID: func(d *viewResourceModel) string { return d.ID.ValueString() }, SetID: func(d *viewResourceModel, id string) { d.ID = types.StringValue(id) }}
}
func (r *viewResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	createResource(ctx, req, resp, r.operations())
}
func (r *viewResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	readResource(ctx, req, resp, r.operations())
}
func (r *viewResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	updateResource(ctx, req, resp, r.operations())
}
func (r *viewResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	deleteResource(ctx, req, resp, r.operations())
}
func (r *viewResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
