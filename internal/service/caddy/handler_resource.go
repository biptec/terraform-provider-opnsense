package caddy

import (
	"context"
	apicaddy "github.com/biptec/opnsense-go/pkg/caddy"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &handlerResource{}
var _ resource.ResourceWithConfigure = &handlerResource{}
var _ resource.ResourceWithImportState = &handlerResource{}

type handlerResource struct{ resourceClient }

func newHandlerResource() resource.Resource { return &handlerResource{} }
func (r *handlerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_caddy_handler"
}
func (r *handlerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = handlerResourceSchema()
}
func (r *handlerResource) operations() crudOperations[handlerResourceModel, apicaddy.Handler] {
	c := r.client.Caddy()
	return crudOperations[handlerResourceModel, apicaddy.Handler]{Name: "Caddy Handler", Convert: convertHandlerSchemaToStruct, Expand: convertHandlerStructToSchema, Add: c.AddHandler, Get: c.GetHandler, Update: c.UpdateHandler, Delete: c.DeleteHandler, GetID: func(d *handlerResourceModel) string { return d.ID.ValueString() }, SetID: func(d *handlerResourceModel, id string) { d.ID = types.StringValue(id) }}
}
func (r *handlerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	createResource(ctx, req, resp, r.operations())
}
func (r *handlerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	readResource(ctx, req, resp, r.operations())
}
func (r *handlerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	updateResource(ctx, req, resp, r.operations())
}
func (r *handlerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	deleteResource(ctx, req, resp, r.operations())
}
func (r *handlerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
