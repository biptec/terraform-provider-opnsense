package routing

import (
	"context"

	apirouting "github.com/biptec/opnsense-go/pkg/routing"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &gatewayGroupResource{}
var _ resource.ResourceWithConfigure = &gatewayGroupResource{}
var _ resource.ResourceWithImportState = &gatewayGroupResource{}

type gatewayGroupResource struct{ routingResourceClient }

func newGatewayGroupResource() resource.Resource { return &gatewayGroupResource{} }
func (r *gatewayGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_routing_gateway_group"
}
func (r *gatewayGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = gatewayGroupResourceSchema()
}
func (r *gatewayGroupResource) operations() routingCRUDOperations[gatewayGroupResourceModel, apirouting.GatewayGroup] {
	controller := r.client.Routing()
	return routingCRUDOperations[gatewayGroupResourceModel, apirouting.GatewayGroup]{
		Name: "Routing Gateway Group", Convert: convertGatewayGroupSchemaToStruct, Expand: convertGatewayGroupStructToSchema,
		Add: func(ctx context.Context, group *apirouting.GatewayGroup) (string, error) {
			return retryGatewayGroupAdd(ctx, controller.AddGatewayGroup, group)
		},
		Get: controller.GetGatewayGroup, Update: controller.UpdateGatewayGroup, Delete: controller.DeleteGatewayGroup,
		GetID: func(d *gatewayGroupResourceModel) string { return d.Id.ValueString() },
		SetID: func(d *gatewayGroupResourceModel, id string) { d.Id = types.StringValue(id) },
	}
}
func (r *gatewayGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	createRoutingResource(ctx, req, resp, r.operations())
}
func (r *gatewayGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	readRoutingResource(ctx, req, resp, r.operations())
}
func (r *gatewayGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	updateRoutingResource(ctx, req, resp, r.operations())
}
func (r *gatewayGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	deleteRoutingResource(ctx, req, resp, r.operations())
}
func (r *gatewayGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
