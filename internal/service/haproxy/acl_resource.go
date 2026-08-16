package haproxy

import (
	"context"

	apihaproxy "github.com/biptec/opnsense-go/pkg/haproxy"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &aclResource{}
var _ resource.ResourceWithConfigure = &aclResource{}
var _ resource.ResourceWithImportState = &aclResource{}

type aclResource struct{ resourceClient }

func newACLResource() resource.Resource { return &aclResource{} }
func (r *aclResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_haproxy_acl"
}
func (r *aclResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = aclResourceSchema()
}
func (r *aclResource) ops() crudOperations[aclModel, apihaproxy.ACL] {
	return crudOperations[aclModel, apihaproxy.ACL]{Name: "HAProxy ACL", Convert: aclModelToAPI, Expand: aclAPIToModel, Add: r.client.Haproxy().AddACL, Get: r.client.Haproxy().GetACL, Update: r.client.Haproxy().UpdateACL, Delete: r.client.Haproxy().DeleteACL, GetID: func(d *aclModel) string { return d.ID.ValueString() }, SetID: func(d *aclModel, id string) { d.ID = types.StringValue(id) }}
}
func (r *aclResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	createResource(ctx, req, resp, r.ops())
}
func (r *aclResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	readResource(ctx, req, resp, r.ops())
}
func (r *aclResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	updateResource(ctx, req, resp, r.ops())
}
func (r *aclResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	deleteResource(ctx, req, resp, r.ops())
}
func (r *aclResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
