package bind

import (
	"context"
	apibind "github.com/biptec/opnsense-go/pkg/bind"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &forwardDomainResource{}
var _ resource.ResourceWithConfigure = &forwardDomainResource{}
var _ resource.ResourceWithImportState = &forwardDomainResource{}

type forwardDomainResource struct{ resourceClient }

func newForwardDomainResource() resource.Resource { return &forwardDomainResource{} }
func (r *forwardDomainResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bind_forward_domain"
}
func (r *forwardDomainResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = forwardDomainResourceSchema()
}
func (r *forwardDomainResource) operations() crudOperations[forwardDomainResourceModel, apibind.ForwardDomain] {
	c := r.client.Bind()
	return crudOperations[forwardDomainResourceModel, apibind.ForwardDomain]{Name: "BIND ForwardDomain", Convert: forwardDomainModelToAPI, Expand: forwardDomainAPIToModel, Add: c.AddForwardDomain, Get: c.GetForwardDomain, Update: c.UpdateForwardDomain, Delete: c.DeleteForwardDomain, GetID: func(d *forwardDomainResourceModel) string { return d.ID.ValueString() }, SetID: func(d *forwardDomainResourceModel, id string) { d.ID = types.StringValue(id) }}
}
func (r *forwardDomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	createResource(ctx, req, resp, r.operations())
}
func (r *forwardDomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	readResource(ctx, req, resp, r.operations())
}
func (r *forwardDomainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	updateResource(ctx, req, resp, r.operations())
}
func (r *forwardDomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	deleteResource(ctx, req, resp, r.operations())
}
func (r *forwardDomainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
