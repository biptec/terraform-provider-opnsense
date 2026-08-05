package bind

import (
	"context"

	apibind "github.com/biptec/opnsense-go/pkg/bind"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &aclResource{}
var _ resource.ResourceWithConfigure = &aclResource{}
var _ resource.ResourceWithImportState = &aclResource{}

type aclResource struct{ resourceClient }

func newAclResource() resource.Resource { return &aclResource{} }
func (r *aclResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bind_acl"
}
func (r *aclResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = aclResourceSchema()
}
func (r *aclResource) operations() crudOperations[aclResourceModel, apibind.Acl] {
	c := r.client.Bind()
	return crudOperations[aclResourceModel, apibind.Acl]{Name: "BIND ACL", Convert: aclModelToAPI, Expand: aclAPIToModel, Add: c.AddAcl, Get: c.GetAcl, Update: c.UpdateAcl, Delete: c.DeleteAcl, GetID: func(d *aclResourceModel) string { return d.ID.ValueString() }, SetID: func(d *aclResourceModel, id string) { d.ID = types.StringValue(id) }}
}
func (r *aclResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	createResource(ctx, req, resp, r.operations())
}
func (r *aclResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	readResource(ctx, req, resp, r.operations())
}
func (r *aclResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	updateResource(ctx, req, resp, r.operations())
}
func (r *aclResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	deleteResource(ctx, req, resp, r.operations())
}
func (r *aclResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
