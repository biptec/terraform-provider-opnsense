package bind

import (
	"context"
	apibind "github.com/biptec/opnsense-go/pkg/bind"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &secondaryDomainResource{}
var _ resource.ResourceWithConfigure = &secondaryDomainResource{}
var _ resource.ResourceWithImportState = &secondaryDomainResource{}

type secondaryDomainResource struct{ resourceClient }

func newSecondaryDomainResource() resource.Resource { return &secondaryDomainResource{} }
func (r *secondaryDomainResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bind_secondary_domain"
}
func (r *secondaryDomainResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = secondaryDomainResourceSchema()
}
func (r *secondaryDomainResource) operations() crudOperations[secondaryDomainResourceModel, apibind.SecondaryDomain] {
	c := r.client.Bind()
	return crudOperations[secondaryDomainResourceModel, apibind.SecondaryDomain]{Name: "BIND SecondaryDomain", Convert: secondaryDomainModelToAPI, Expand: secondaryDomainAPIToModel, Add: c.AddSecondaryDomain, Get: c.GetSecondaryDomain, Update: c.UpdateSecondaryDomain, Delete: c.DeleteSecondaryDomain, GetID: func(d *secondaryDomainResourceModel) string { return d.ID.ValueString() }, SetID: func(d *secondaryDomainResourceModel, id string) { d.ID = types.StringValue(id) }}
}
func (r *secondaryDomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	createResource(ctx, req, resp, r.operations())
}
func (r *secondaryDomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	readResource(ctx, req, resp, r.operations())
}
func (r *secondaryDomainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	updateResource(ctx, req, resp, r.operations())
}
func (r *secondaryDomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	deleteResource(ctx, req, resp, r.operations())
}
func (r *secondaryDomainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
