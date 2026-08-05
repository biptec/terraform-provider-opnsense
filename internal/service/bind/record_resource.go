package bind

import (
	"context"
	apibind "github.com/biptec/opnsense-go/pkg/bind"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &recordResource{}
var _ resource.ResourceWithConfigure = &recordResource{}
var _ resource.ResourceWithImportState = &recordResource{}

type recordResource struct{ resourceClient }

func newRecordResource() resource.Resource { return &recordResource{} }
func (r *recordResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bind_record"
}
func (r *recordResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = recordResourceSchema()
}
func (r *recordResource) operations() crudOperations[recordResourceModel, apibind.Record] {
	c := r.client.Bind()
	return crudOperations[recordResourceModel, apibind.Record]{Name: "BIND Record", Convert: recordModelToAPI, Expand: recordAPIToModel, Add: c.AddRecord, Get: c.GetRecord, Update: c.UpdateRecord, Delete: c.DeleteRecord, GetID: func(d *recordResourceModel) string { return d.ID.ValueString() }, SetID: func(d *recordResourceModel, id string) { d.ID = types.StringValue(id) }}
}
func (r *recordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	createResource(ctx, req, resp, r.operations())
}
func (r *recordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	readResource(ctx, req, resp, r.operations())
}
func (r *recordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	updateResource(ctx, req, resp, r.operations())
}
func (r *recordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	deleteResource(ctx, req, resp, r.operations())
}
func (r *recordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
