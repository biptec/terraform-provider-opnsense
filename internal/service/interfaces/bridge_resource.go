package interfaces

import (
	"context"

	apiinterfaces "github.com/biptec/opnsense-go/pkg/interfaces"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ resource.Resource = &bridgeResource{}
var _ resource.ResourceWithConfigure = &bridgeResource{}
var _ resource.ResourceWithImportState = &bridgeResource{}

type bridgeResource struct{ interfaceResourceClient }

func newBridgeResource() resource.Resource { return &bridgeResource{} }
func (r *bridgeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_interfaces_bridge"
}
func (r *bridgeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = bridgeResourceSchema()
}
func (r *bridgeResource) operations() interfaceCRUDOperations[bridgeResourceModel, apiinterfaces.Bridge] {
	controller := r.client.Interfaces()
	return interfaceCRUDOperations[bridgeResourceModel, apiinterfaces.Bridge]{Name: "Bridge", Convert: convertBridgeSchemaToStruct, Expand: convertBridgeStructToSchema, Add: controller.AddBridge, Get: controller.GetBridge, Update: controller.UpdateBridge, Delete: controller.DeleteBridge, GetID: func(d *bridgeResourceModel) string { return d.Id.ValueString() }, SetID: func(d *bridgeResourceModel, id string) { d.Id = typesString(id) }}
}
func (r *bridgeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	createInterfaceResource(ctx, req, resp, r.operations())
}
func (r *bridgeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	readInterfaceResource(ctx, req, resp, r.operations())
}
func (r *bridgeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	updateInterfaceResource(ctx, req, resp, r.operations())
}
func (r *bridgeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	deleteInterfaceResource(ctx, req, resp, r.operations())
}
func (r *bridgeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
