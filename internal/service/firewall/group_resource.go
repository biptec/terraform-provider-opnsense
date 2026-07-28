package firewall

import (
	"context"

	apifirewall "github.com/biptec/opnsense-go/pkg/firewall"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &groupResource{}
var _ resource.ResourceWithConfigure = &groupResource{}
var _ resource.ResourceWithImportState = &groupResource{}

type groupResource struct{ firewallCoreResourceClient }

func newGroupResource() resource.Resource { return &groupResource{} }
func (r *groupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_group"
}
func (r *groupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = groupResourceSchema()
}
func (r *groupResource) operations() firewallCoreCRUDOperations[groupResourceModel, apifirewall.Group] {
	controller := r.client.Firewall()
	return firewallCoreCRUDOperations[groupResourceModel, apifirewall.Group]{
		Name: "Firewall Group", Convert: convertGroupSchemaToStruct, Expand: convertGroupStructToSchema,
		Add: controller.AddGroup, Get: controller.GetGroup, Update: controller.UpdateGroup, Delete: controller.DeleteGroup,
		GetID: func(d *groupResourceModel) string { return d.Id.ValueString() },
		SetID: func(d *groupResourceModel, id string) { d.Id = types.StringValue(id) },
	}
}
func (r *groupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	createFirewallCoreResource(ctx, req, resp, r.operations())
}
func (r *groupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	readFirewallCoreResource(ctx, req, resp, r.operations())
}
func (r *groupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	updateFirewallCoreResource(ctx, req, resp, r.operations())
}
func (r *groupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	deleteFirewallCoreResource(ctx, req, resp, r.operations())
}
func (r *groupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
