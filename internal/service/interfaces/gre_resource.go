package interfaces

import (
	"context"

	apiinterfaces "github.com/biptec/opnsense-go/pkg/interfaces"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ resource.Resource = &greResource{}
var _ resource.ResourceWithConfigure = &greResource{}
var _ resource.ResourceWithImportState = &greResource{}

type greResource struct{ interfaceResourceClient }

func newGreResource() resource.Resource { return &greResource{} }
func (r *greResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_interfaces_gre"
}
func (r *greResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = greResourceSchema()
}
func (r *greResource) operations() interfaceCRUDOperations[greResourceModel, apiinterfaces.Gre] {
	c := r.client.Interfaces()
	return interfaceCRUDOperations[greResourceModel, apiinterfaces.Gre]{Name: "GRE", Convert: convertGreSchemaToStruct, Expand: convertGreStructToSchema, Add: c.AddGre, Get: c.GetGre, Update: c.UpdateGre, Delete: c.DeleteGre, GetID: func(d *greResourceModel) string { return d.Id.ValueString() }, SetID: func(d *greResourceModel, id string) { d.Id = typesString(id) }}
}
func (r *greResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	createInterfaceResource(ctx, req, resp, r.operations())
}
func (r *greResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	readInterfaceResource(ctx, req, resp, r.operations())
}
func (r *greResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	updateInterfaceResource(ctx, req, resp, r.operations())
}
func (r *greResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	deleteInterfaceResource(ctx, req, resp, r.operations())
}
func (r *greResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
