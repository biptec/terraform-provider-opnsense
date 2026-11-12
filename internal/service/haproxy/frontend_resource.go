package haproxy

import (
	"context"

	"github.com/biptec/opnsense-go/pkg/api"
	apihaproxy "github.com/biptec/opnsense-go/pkg/haproxy"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &frontendResource{}
var _ resource.ResourceWithConfigure = &frontendResource{}
var _ resource.ResourceWithImportState = &frontendResource{}

type frontendResource struct{ resourceClient }

func newFrontendResource() resource.Resource { return &frontendResource{} }
func (r *frontendResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_haproxy_frontend"
}
func (r *frontendResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = frontendResourceSchema()
}
func prepareFrontendCreate(_ context.Context, frontend *apihaproxy.Frontend) error {
	// ssl_certificates may be association-owned after creation. OPNsense still
	// requires the default certificate to be present in the certificate list
	// when a TLS frontend is first created, so seed only that local default
	// when the parent intentionally omits complete-set ownership.
	if frontend.SSLCertificates == nil &&
		frontend.SSLEnabled == "1" &&
		frontend.SSLDefaultCertificate.String() != "" {
		frontend.SSLCertificates = api.SelectedMapList{frontend.SSLDefaultCertificate.String()}
	}
	return nil
}

func (r *frontendResource) ops() crudOperations[frontendModel, apihaproxy.Frontend] {
	return crudOperations[frontendModel, apihaproxy.Frontend]{
		Name:           "HAProxy Frontend",
		Convert:        frontendModelToAPI,
		Expand:         frontendAPIToModel,
		Add:            r.client.Haproxy().AddFrontend,
		Get:            r.client.Haproxy().GetFrontend,
		Update:         r.client.Haproxy().UpdateFrontend,
		Delete:         r.client.Haproxy().DeleteFrontend,
		GetID:          func(d *frontendModel) string { return d.ID.ValueString() },
		SetID:          func(d *frontendModel, id string) { d.ID = types.StringValue(id) },
		ValidateCreate: prepareFrontendCreate,
		Apply:          r.applyHAProxyConfig,
	}
}
func (r *frontendResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	createResource(ctx, req, resp, r.ops())
}
func (r *frontendResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	readResource(ctx, req, resp, r.ops())
}
func (r *frontendResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	updateResource(ctx, req, resp, r.ops())
}
func (r *frontendResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	deleteResource(ctx, req, resp, r.ops())
}
func (r *frontendResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
