package bind

import (
	"context"
	apibind "github.com/biptec/opnsense-go/pkg/bind"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &tsigKeyResource{}
var _ resource.ResourceWithConfigure = &tsigKeyResource{}
var _ resource.ResourceWithImportState = &tsigKeyResource{}

type tsigKeyResource struct{ resourceClient }

func newTsigKeyResource() resource.Resource { return &tsigKeyResource{} }
func (r *tsigKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bind_tsig_key"
}
func (r *tsigKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = tsigKeyResourceSchema()
}
func (r *tsigKeyResource) operations() crudOperations[tsigKeyResourceModel, apibind.TsigKey] {
	c := r.client.Bind()
	return crudOperations[tsigKeyResourceModel, apibind.TsigKey]{Name: "BIND TSIG Key", Convert: tsigKeyModelToAPI, Expand: tsigKeyAPIToModel, Add: c.AddTsigKey, Get: c.GetTsigKey, Update: c.UpdateTsigKey, Delete: c.DeleteTsigKey, GetID: func(d *tsigKeyResourceModel) string { return d.ID.ValueString() }, SetID: func(d *tsigKeyResourceModel, id string) { d.ID = types.StringValue(id) }}
}
func (r *tsigKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	createResource(ctx, req, resp, r.operations())
}
func (r *tsigKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	readResource(ctx, req, resp, r.operations())
}
func (r *tsigKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	updateResource(ctx, req, resp, r.operations())
}
func (r *tsigKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	deleteResource(ctx, req, resp, r.operations())
}
func (r *tsigKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
