package interfaces

import (
	"context"
	apiinterfaces "github.com/biptec/opnsense-go/pkg/interfaces"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ resource.Resource = &loopbackResource{}
var _ resource.ResourceWithConfigure = &loopbackResource{}
var _ resource.ResourceWithImportState = &loopbackResource{}

type loopbackResource struct{ interfaceResourceClient }

func newLoopbackResource() resource.Resource { return &loopbackResource{} }
func (r *loopbackResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_interfaces_loopback"
}
func (r *loopbackResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = loopbackResourceSchema()
}
func (r *loopbackResource) operations() interfaceCRUDOperations[loopbackResourceModel, apiinterfaces.Loopback] {
	c := r.client.Interfaces()
	return interfaceCRUDOperations[loopbackResourceModel, apiinterfaces.Loopback]{Name: "Loopback", Convert: convertLoopbackSchemaToStruct, Expand: convertLoopbackStructToSchema, Add: c.AddLoopback, Get: c.GetLoopback, Update: c.UpdateLoopback, Delete: c.DeleteLoopback, GetID: func(d *loopbackResourceModel) string { return d.Id.ValueString() }, SetID: func(d *loopbackResourceModel, id string) { d.Id = typesString(id) }}
}
func (r *loopbackResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	createInterfaceResource(ctx, req, resp, r.operations())
}
func (r *loopbackResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	readInterfaceResource(ctx, req, resp, r.operations())
}
func (r *loopbackResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	updateInterfaceResource(ctx, req, resp, r.operations())
}
func (r *loopbackResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	deleteInterfaceResource(ctx, req, resp, r.operations())
}
func (r *loopbackResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
