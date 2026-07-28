package interfaces

import (
	"context"

	apiinterfaces "github.com/biptec/opnsense-go/pkg/interfaces"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ resource.Resource = &vxlanResource{}
var _ resource.ResourceWithConfigure = &vxlanResource{}
var _ resource.ResourceWithImportState = &vxlanResource{}

type vxlanResource struct{ interfaceResourceClient }

func newVxlanResource() resource.Resource { return &vxlanResource{} }
func (r *vxlanResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_interfaces_vxlan"
}
func (r *vxlanResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = vxlanResourceSchema()
}
func (r *vxlanResource) operations() interfaceCRUDOperations[vxlanResourceModel, apiinterfaces.Vxlan] {
	controller := r.client.Interfaces()
	return interfaceCRUDOperations[vxlanResourceModel, apiinterfaces.Vxlan]{Name: "VXLAN", Convert: convertVxlanSchemaToStruct, Expand: convertVxlanStructToSchema, Add: controller.AddVxlan, Get: controller.GetVxlan, Update: controller.UpdateVxlan, Delete: controller.DeleteVxlan, GetID: func(d *vxlanResourceModel) string { return d.Id.ValueString() }, SetID: func(d *vxlanResourceModel, id string) { d.Id = typesString(id) }}
}
func (r *vxlanResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	createInterfaceResource(ctx, req, resp, r.operations())
}
func (r *vxlanResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	readInterfaceResource(ctx, req, resp, r.operations())
}
func (r *vxlanResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	updateInterfaceResource(ctx, req, resp, r.operations())
}
func (r *vxlanResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	deleteInterfaceResource(ctx, req, resp, r.operations())
}
func (r *vxlanResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
