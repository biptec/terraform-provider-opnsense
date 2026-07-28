package interfaces

import (
	"context"
	apiinterfaces "github.com/biptec/opnsense-go/pkg/interfaces"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ resource.Resource = &gifResource{}
var _ resource.ResourceWithConfigure = &gifResource{}
var _ resource.ResourceWithImportState = &gifResource{}

type gifResource struct{ interfaceResourceClient }

func newGifResource() resource.Resource { return &gifResource{} }
func (r *gifResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_interfaces_gif"
}
func (r *gifResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = gifResourceSchema()
}
func (r *gifResource) operations() interfaceCRUDOperations[gifResourceModel, apiinterfaces.Gif] {
	c := r.client.Interfaces()
	return interfaceCRUDOperations[gifResourceModel, apiinterfaces.Gif]{Name: "GIF", Convert: convertGifSchemaToStruct, Expand: convertGifStructToSchema, Add: c.AddGif, Get: c.GetGif, Update: c.UpdateGif, Delete: c.DeleteGif, GetID: func(d *gifResourceModel) string { return d.Id.ValueString() }, SetID: func(d *gifResourceModel, id string) { d.Id = typesString(id) }}
}
func (r *gifResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	createInterfaceResource(ctx, req, resp, r.operations())
}
func (r *gifResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	readInterfaceResource(ctx, req, resp, r.operations())
}
func (r *gifResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	updateInterfaceResource(ctx, req, resp, r.operations())
}
func (r *gifResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	deleteInterfaceResource(ctx, req, resp, r.operations())
}
func (r *gifResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
