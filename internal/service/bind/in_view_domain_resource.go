package bind

import (
	"context"
	apibind "github.com/biptec/opnsense-go/pkg/bind"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &inViewDomainResource{}
var _ resource.ResourceWithConfigure = &inViewDomainResource{}
var _ resource.ResourceWithImportState = &inViewDomainResource{}

type inViewDomainResource struct{ resourceClient }

func newInViewDomainResource() resource.Resource { return &inViewDomainResource{} }
func (r *inViewDomainResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bind_in_view_domain"
}
func (r *inViewDomainResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = inViewDomainResourceSchema()
}
func (r *inViewDomainResource) operations() crudOperations[inViewDomainResourceModel, apibind.InViewDomain] {
	c := r.client.Bind()
	return crudOperations[inViewDomainResourceModel, apibind.InViewDomain]{Name: "BIND InViewDomain", Convert: inViewDomainModelToAPI, Expand: inViewDomainAPIToModel, Add: c.AddInViewDomain, Get: c.GetInViewDomain, Update: c.UpdateInViewDomain, Delete: c.DeleteInViewDomain, GetID: func(d *inViewDomainResourceModel) string { return d.ID.ValueString() }, SetID: func(d *inViewDomainResourceModel, id string) { d.ID = types.StringValue(id) }}
}
func (r *inViewDomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	createResource(ctx, req, resp, r.operations())
}
func (r *inViewDomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	readResource(ctx, req, resp, r.operations())
}
func (r *inViewDomainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	updateResource(ctx, req, resp, r.operations())
}
func (r *inViewDomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	deleteResource(ctx, req, resp, r.operations())
}
func (r *inViewDomainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
