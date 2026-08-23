package haproxy

import (
	"context"
	"fmt"

	apihaproxy "github.com/biptec/opnsense-go/pkg/haproxy"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &serverResource{}
var _ resource.ResourceWithConfigure = &serverResource{}
var _ resource.ResourceWithImportState = &serverResource{}

type serverResource struct{ resourceClient }

func newServerResource() resource.Resource { return &serverResource{} }
func (r *serverResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_haproxy_server"
}
func (r *serverResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = serverResourceSchema()
}
func (r *serverResource) validateUniqueName(ctx context.Context, ownID string, data *apihaproxy.Server) error {
	result, err := r.client.Haproxy().SearchServer(ctx)
	if err != nil {
		return fmt.Errorf("check existing HAProxy servers: %w", err)
	}
	items := make([]namedRemoteItem, 0, len(result.Rows))
	for _, item := range result.Rows {
		items = append(items, namedRemoteItem{ID: item.UUID, Name: item.Name})
	}
	return validateUniqueRemoteName("HAProxy server", data.Name, ownID, items)
}

func (r *serverResource) ops() crudOperations[serverModel, apihaproxy.Server] {
	return crudOperations[serverModel, apihaproxy.Server]{
		Name: "HAProxy Server", Convert: serverModelToAPI, Expand: serverAPIToModel,
		Add: r.client.Haproxy().AddServer, Get: r.client.Haproxy().GetServer, Update: r.client.Haproxy().UpdateServer, Delete: r.client.Haproxy().DeleteServer,
		GetID: func(d *serverModel) string { return d.ID.ValueString() }, SetID: func(d *serverModel, id string) { d.ID = types.StringValue(id) },
		ValidateCreate: func(ctx context.Context, data *apihaproxy.Server) error { return r.validateUniqueName(ctx, "", data) },
		ValidateCreated: func(ctx context.Context, id string, data *apihaproxy.Server) error {
			return r.validateUniqueName(ctx, id, data)
		},
		ValidateUpdate: func(ctx context.Context, id string, data *apihaproxy.Server) error {
			return r.validateUniqueName(ctx, id, data)
		},
		Apply: r.applyHAProxyConfig,
	}
}
func (r *serverResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	createResource(ctx, req, resp, r.ops())
}
func (r *serverResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	readResource(ctx, req, resp, r.ops())
}
func (r *serverResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	updateResource(ctx, req, resp, r.ops())
}
func (r *serverResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	deleteResource(ctx, req, resp, r.ops())
}
func (r *serverResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
