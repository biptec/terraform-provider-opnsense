package routing

import (
	"context"

	apirouting "github.com/biptec/opnsense-go/pkg/routing"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &gatewayResource{}
var _ resource.ResourceWithConfigure = &gatewayResource{}
var _ resource.ResourceWithImportState = &gatewayResource{}

type gatewayResource struct{ routingResourceClient }

func newGatewayResource() resource.Resource { return &gatewayResource{} }
func (r *gatewayResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_routing_gateway"
}
func (r *gatewayResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = gatewayResourceSchema()
}
func (r *gatewayResource) operations() routingCRUDOperations[gatewayResourceModel, apirouting.Gateway] {
	controller := r.client.Routing()
	return routingCRUDOperations[gatewayResourceModel, apirouting.Gateway]{
		Name: "Routing Gateway", Convert: convertGatewaySchemaToStruct, Expand: convertGatewayStructToSchema,
		Add: controller.AddGateway, Get: controller.GetGateway, Update: controller.UpdateGateway, Delete: controller.DeleteGateway,
		GetID: func(d *gatewayResourceModel) string { return d.Id.ValueString() },
		SetID: func(d *gatewayResourceModel, id string) { d.Id = types.StringValue(id) },
	}
}
func (r *gatewayResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	createRoutingResource(ctx, req, resp, r.operations())
}
func (r *gatewayResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	readRoutingResource(ctx, req, resp, r.operations())
}
func (r *gatewayResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	updateRoutingResource(ctx, req, resp, r.operations())
}
func (r *gatewayResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	deleteRoutingResource(ctx, req, resp, r.operations())
}
func (r *gatewayResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
