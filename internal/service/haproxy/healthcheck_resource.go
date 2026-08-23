package haproxy

import (
	"context"

	apihaproxy "github.com/biptec/opnsense-go/pkg/haproxy"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &healthcheckResource{}
var _ resource.ResourceWithConfigure = &healthcheckResource{}
var _ resource.ResourceWithImportState = &healthcheckResource{}

type healthcheckResource struct{ resourceClient }

func newHealthcheckResource() resource.Resource { return &healthcheckResource{} }
func (r *healthcheckResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_haproxy_healthcheck"
}
func (r *healthcheckResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = healthcheckResourceSchema()
}
func (r *healthcheckResource) ops() crudOperations[healthcheckModel, apihaproxy.Healthcheck] {
	return crudOperations[healthcheckModel, apihaproxy.Healthcheck]{Name: "HAProxy Healthcheck", Convert: healthcheckModelToAPI, Expand: healthcheckAPIToModel, Add: r.client.Haproxy().AddHealthcheck, Get: r.client.Haproxy().GetHealthcheck, Update: r.client.Haproxy().UpdateHealthcheck, Delete: r.client.Haproxy().DeleteHealthcheck, GetID: func(d *healthcheckModel) string { return d.ID.ValueString() }, SetID: func(d *healthcheckModel, id string) { d.ID = types.StringValue(id) }, Apply: r.applyHAProxyConfig}
}
func (r *healthcheckResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	createResource(ctx, req, resp, r.ops())
}
func (r *healthcheckResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	readResource(ctx, req, resp, r.ops())
}
func (r *healthcheckResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	updateResource(ctx, req, resp, r.ops())
}
func (r *healthcheckResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	deleteResource(ctx, req, resp, r.ops())
}
func (r *healthcheckResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
