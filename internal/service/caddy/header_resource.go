package caddy

import (
	"context"

	apicaddy "github.com/biptec/opnsense-go/pkg/caddy"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &headerResource{}
var _ resource.ResourceWithConfigure = &headerResource{}
var _ resource.ResourceWithImportState = &headerResource{}

type headerResource struct{ resourceClient }

func newHeaderResource() resource.Resource { return &headerResource{} }

func (r *headerResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_caddy_header"
}

func (r *headerResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = headerResourceSchema()
}

func (r *headerResource) operations() crudOperations[headerResourceModel, apicaddy.Header] {
	client := r.client.Caddy()
	return crudOperations[headerResourceModel, apicaddy.Header]{
		Name:    "Caddy Header",
		Convert: convertHeaderSchemaToStruct,
		Expand:  convertHeaderStructToSchema,
		Add:     client.AddHeader,
		Get:     client.GetHeader,
		Update:  client.UpdateHeader,
		Delete:  client.DeleteHeader,
		GetID: func(d *headerResourceModel) string {
			return d.ID.ValueString()
		},
		SetID: func(d *headerResourceModel, id string) {
			d.ID = types.StringValue(id)
		},
	}
}

func (r *headerResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	createResource(ctx, req, resp, r.operations())
}

func (r *headerResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	readResource(ctx, req, resp, r.operations())
}

func (r *headerResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	updateResource(ctx, req, resp, r.operations())
}

func (r *headerResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	deleteResource(ctx, req, resp, r.operations())
}

func (r *headerResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
