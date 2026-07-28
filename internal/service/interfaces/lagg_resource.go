package interfaces

import (
	"context"

	apiinterfaces "github.com/biptec/opnsense-go/pkg/interfaces"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ resource.Resource = &laggResource{}
var _ resource.ResourceWithConfigure = &laggResource{}
var _ resource.ResourceWithImportState = &laggResource{}

type laggResource struct{ interfaceResourceClient }

func newLaggResource() resource.Resource { return &laggResource{} }
func (r *laggResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_interfaces_lagg"
}
func (r *laggResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = laggResourceSchema()
}
func (r *laggResource) operations() interfaceCRUDOperations[laggResourceModel, apiinterfaces.Lagg] {
	controller := r.client.Interfaces()
	return interfaceCRUDOperations[laggResourceModel, apiinterfaces.Lagg]{Name: "LAGG", Convert: convertLaggSchemaToStruct, Expand: convertLaggStructToSchema, Add: controller.AddLagg, Get: controller.GetLagg, Update: controller.UpdateLagg, Delete: controller.DeleteLagg, GetID: func(d *laggResourceModel) string { return d.Id.ValueString() }, SetID: func(d *laggResourceModel, id string) { d.Id = typesString(id) }}
}
func (r *laggResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	createInterfaceResource(ctx, req, resp, r.operations())
}
func (r *laggResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	readInterfaceResource(ctx, req, resp, r.operations())
}
func (r *laggResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	updateInterfaceResource(ctx, req, resp, r.operations())
}
func (r *laggResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	deleteInterfaceResource(ctx, req, resp, r.operations())
}
func (r *laggResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
