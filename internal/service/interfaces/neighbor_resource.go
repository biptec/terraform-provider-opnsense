package interfaces

import (
	"context"
	apiinterfaces "github.com/biptec/opnsense-go/pkg/interfaces"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ resource.Resource = &neighborResource{}
var _ resource.ResourceWithConfigure = &neighborResource{}
var _ resource.ResourceWithImportState = &neighborResource{}

type neighborResource struct{ interfaceResourceClient }

func newNeighborResource() resource.Resource { return &neighborResource{} }
func (r *neighborResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_interfaces_neighbor"
}
func (r *neighborResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = neighborResourceSchema()
}
func (r *neighborResource) operations() interfaceCRUDOperations[neighborResourceModel, apiinterfaces.Neighbor] {
	c := r.client.Interfaces()
	return interfaceCRUDOperations[neighborResourceModel, apiinterfaces.Neighbor]{Name: "Neighbor", Convert: convertNeighborSchemaToStruct, Expand: convertNeighborStructToSchema, Add: c.AddNeighbor, Get: c.GetNeighbor, Update: c.UpdateNeighbor, Delete: c.DeleteNeighbor, GetID: func(d *neighborResourceModel) string { return d.Id.ValueString() }, SetID: func(d *neighborResourceModel, id string) { d.Id = typesString(id) }}
}
func (r *neighborResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	createInterfaceResource(ctx, req, resp, r.operations())
}
func (r *neighborResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	readInterfaceResource(ctx, req, resp, r.operations())
}
func (r *neighborResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	updateInterfaceResource(ctx, req, resp, r.operations())
}
func (r *neighborResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	deleteInterfaceResource(ctx, req, resp, r.operations())
}
func (r *neighborResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
