package caddy

import (
	"context"
	apicaddy "github.com/biptec/opnsense-go/pkg/caddy"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &accessListResource{}
var _ resource.ResourceWithConfigure = &accessListResource{}
var _ resource.ResourceWithImportState = &accessListResource{}

type accessListResource struct{ resourceClient }

func newAccessListResource() resource.Resource { return &accessListResource{} }
func (r *accessListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_caddy_access_list"
}
func (r *accessListResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = accessListResourceSchema()
}
func (r *accessListResource) operations() crudOperations[accessListResourceModel, apicaddy.AccessList] {
	c := r.client.Caddy()
	return crudOperations[accessListResourceModel, apicaddy.AccessList]{Name: "Caddy Access List", Convert: convertAccessListSchemaToStruct, Expand: convertAccessListStructToSchema, Add: c.AddAccessList, Get: c.GetAccessList, Update: c.UpdateAccessList, Delete: c.DeleteAccessList, GetID: func(d *accessListResourceModel) string { return d.ID.ValueString() }, SetID: func(d *accessListResourceModel, id string) { d.ID = types.StringValue(id) }}
}
func (r *accessListResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	createResource(ctx, req, resp, r.operations())
}
func (r *accessListResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	readResource(ctx, req, resp, r.operations())
}
func (r *accessListResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	updateResource(ctx, req, resp, r.operations())
}
func (r *accessListResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	deleteResource(ctx, req, resp, r.operations())
}
func (r *accessListResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
