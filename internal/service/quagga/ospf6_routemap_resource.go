package quagga

import (
	"context"
	"errors"
	"github.com/biptec/opnsense-go/pkg/errs"
	"github.com/biptec/opnsense-go/pkg/opnsense"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ resource.Resource = &ospf6RouteMapResource{}
var _ resource.ResourceWithConfigure = &ospf6RouteMapResource{}
var _ resource.ResourceWithImportState = &ospf6RouteMapResource{}

type ospf6RouteMapResource struct{ client opnsense.Client }

func newOspf6RouteMapResource() resource.Resource { return &ospf6RouteMapResource{} }
func (r *ospf6RouteMapResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quagga_ospf6_routemap"
}
func (r *ospf6RouteMapResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = ospf6RouteMapResourceSchema()
}
func (r *ospf6RouteMapResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureQuaggaResource(req, resp)
}
func (r *ospf6RouteMapResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ospf6RouteMapModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := r.client.Quagga().AddOSPF6RouteMap(ctx, ospf6RouteMapToAPI(&data))
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create OSPFv3 Route Map", err.Error())
		return
	}
	remote, err := r.client.Quagga().GetOSPF6RouteMap(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("OSPFv3 Route Map Created but Read Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, ospf6RouteMapFromAPI(remote, id))...)
}
func (r *ospf6RouteMapResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ospf6RouteMapModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	remote, err := r.client.Quagga().GetOSPF6RouteMap(ctx, data.ID.ValueString())
	if err != nil {
		var nf *errs.NotFoundError
		if errors.As(err, &nf) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to Read OSPFv3 Route Map", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, ospf6RouteMapFromAPI(remote, data.ID.ValueString()))...)
}
func (r *ospf6RouteMapResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ospf6RouteMapModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.Quagga().UpdateOSPF6RouteMap(ctx, data.ID.ValueString(), ospf6RouteMapToAPI(&data)); err != nil {
		resp.Diagnostics.AddError("Unable to Update OSPFv3 Route Map", err.Error())
		return
	}
	remote, err := r.client.Quagga().GetOSPF6RouteMap(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("OSPFv3 Route Map Updated but Read Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, ospf6RouteMapFromAPI(remote, data.ID.ValueString()))...)
}
func (r *ospf6RouteMapResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ospf6RouteMapModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.Quagga().DeleteOSPF6RouteMap(ctx, data.ID.ValueString()); err != nil {
		var nf *errs.NotFoundError
		if !errors.As(err, &nf) {
			resp.Diagnostics.AddError("Unable to Delete OSPFv3 Route Map", err.Error())
		}
	}
}
func (r *ospf6RouteMapResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
