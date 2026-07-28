package firewall

import (
	"context"

	apifirewall "github.com/biptec/opnsense-go/pkg/firewall"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &nptResource{}
var _ resource.ResourceWithConfigure = &nptResource{}
var _ resource.ResourceWithImportState = &nptResource{}

type nptResource struct{ firewallCoreResourceClient }

func newNptResource() resource.Resource { return &nptResource{} }
func (r *nptResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_npt"
}
func (r *nptResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = nptResourceSchema()
}
func (r *nptResource) operations() firewallCoreCRUDOperations[nptResourceModel, apifirewall.Npt] {
	controller := r.client.Firewall()
	return firewallCoreCRUDOperations[nptResourceModel, apifirewall.Npt]{
		Name: "Firewall NPT", Convert: convertNptSchemaToStruct, Expand: convertNptStructToSchema,
		Add: controller.AddNpt, Get: controller.GetNpt, Update: controller.UpdateNpt, Delete: controller.DeleteNpt,
		GetID: func(d *nptResourceModel) string { return d.Id.ValueString() },
		SetID: func(d *nptResourceModel, id string) { d.Id = types.StringValue(id) },
	}
}
func (r *nptResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	createFirewallCoreResource(ctx, req, resp, r.operations())
}
func (r *nptResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	readFirewallCoreResource(ctx, req, resp, r.operations())
}
func (r *nptResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	updateFirewallCoreResource(ctx, req, resp, r.operations())
}
func (r *nptResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	deleteFirewallCoreResource(ctx, req, resp, r.operations())
}
func (r *nptResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
