package haproxy

import (
	"context"
	"fmt"

	apihaproxy "github.com/biptec/opnsense-go/pkg/haproxy"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &backendResource{}
var _ resource.ResourceWithConfigure = &backendResource{}
var _ resource.ResourceWithImportState = &backendResource{}

type backendResource struct{ resourceClient }

func newBackendResource() resource.Resource { return &backendResource{} }
func (r *backendResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_haproxy_backend"
}
func (r *backendResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = backendResourceSchema()
}
func (r *backendResource) validateUniqueName(ctx context.Context, ownID string, data *apihaproxy.Backend) error {
	result, err := r.client.Haproxy().SearchBackend(ctx)
	if err != nil {
		return fmt.Errorf("check existing HAProxy backends: %w", err)
	}
	items := make([]namedRemoteItem, 0, len(result.Rows))
	for _, item := range result.Rows {
		items = append(items, namedRemoteItem{ID: item.UUID, Name: item.Name})
	}
	return validateUniqueRemoteName("HAProxy backend", data.Name, ownID, items)
}

func (r *backendResource) ops() crudOperations[backendModel, apihaproxy.Backend] {
	return crudOperations[backendModel, apihaproxy.Backend]{
		Name: "HAProxy Backend", Convert: backendModelToAPI, Expand: backendAPIToModel,
		Add: r.client.Haproxy().AddBackend, Get: r.client.Haproxy().GetBackend, Update: r.client.Haproxy().UpdateBackend, Delete: r.client.Haproxy().DeleteBackend,
		GetID: func(d *backendModel) string { return d.ID.ValueString() }, SetID: func(d *backendModel, id string) { d.ID = types.StringValue(id) },
		ValidateCreate: func(ctx context.Context, data *apihaproxy.Backend) error { return r.validateUniqueName(ctx, "", data) },
		ValidateCreated: func(ctx context.Context, id string, data *apihaproxy.Backend) error {
			return r.validateUniqueName(ctx, id, data)
		},
		ValidateUpdate: func(ctx context.Context, id string, data *apihaproxy.Backend) error {
			return r.validateUniqueName(ctx, id, data)
		},
		Apply: r.applyHAProxyConfig,
	}
}
func (r *backendResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	createResource(ctx, req, resp, r.ops())
}
func (r *backendResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	readResource(ctx, req, resp, r.ops())
}
func (r *backendResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	updateResource(ctx, req, resp, r.ops())
}
func (r *backendResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	deleteResource(ctx, req, resp, r.ops())
}
func (r *backendResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
