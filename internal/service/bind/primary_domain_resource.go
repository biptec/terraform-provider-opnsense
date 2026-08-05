package bind

import (
	"context"
	apibind "github.com/biptec/opnsense-go/pkg/bind"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &primaryDomainResource{}
var _ resource.ResourceWithConfigure = &primaryDomainResource{}
var _ resource.ResourceWithImportState = &primaryDomainResource{}

type primaryDomainResource struct{ resourceClient }

func newPrimaryDomainResource() resource.Resource { return &primaryDomainResource{} }
func (r *primaryDomainResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bind_primary_domain"
}
func (r *primaryDomainResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = primaryDomainResourceSchema()
}
func (r *primaryDomainResource) operations() crudOperations[primaryDomainResourceModel, apibind.PrimaryDomain] {
	c := r.client.Bind()
	return crudOperations[primaryDomainResourceModel, apibind.PrimaryDomain]{Name: "BIND Primary Domain", Convert: primaryDomainModelToAPI, Expand: primaryDomainAPIToModel, Add: c.AddPrimaryDomain, Get: c.GetPrimaryDomain, Update: c.UpdatePrimaryDomain, Delete: c.DeletePrimaryDomain, GetID: func(d *primaryDomainResourceModel) string { return d.ID.ValueString() }, SetID: func(d *primaryDomainResourceModel, id string) { d.ID = types.StringValue(id) }}
}
func (r *primaryDomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	createResource(ctx, req, resp, r.operations())
}
func (r *primaryDomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	readResource(ctx, req, resp, r.operations())
}
func (r *primaryDomainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	updateResource(ctx, req, resp, r.operations())
}
func (r *primaryDomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	deleteResource(ctx, req, resp, r.operations())
}
func (r *primaryDomainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
