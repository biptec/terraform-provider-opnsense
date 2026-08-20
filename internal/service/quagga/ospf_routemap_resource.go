package quagga

import (
	"context"
	"errors"
	"github.com/biptec/opnsense-go/pkg/errs"
	"github.com/biptec/opnsense-go/pkg/opnsense"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ resource.Resource = &ospfRouteMapResource{}
var _ resource.ResourceWithConfigure = &ospfRouteMapResource{}
var _ resource.ResourceWithImportState = &ospfRouteMapResource{}

type ospfRouteMapResource struct{ client opnsense.Client }

func newOspfRouteMapResource() resource.Resource { return &ospfRouteMapResource{} }
func (r *ospfRouteMapResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quagga_ospf_routemap"
}
func (r *ospfRouteMapResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = ospfRouteMapResourceSchema()
}
func (r *ospfRouteMapResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureQuaggaResource(req, resp)
}
func (r *ospfRouteMapResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ospfRouteMapModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := r.client.Quagga().AddOSPFRouteMap(ctx, ospfRouteMapToAPI(&data))
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create OSPFv2 Route Map", err.Error())
		return
	}
	remote, err := r.client.Quagga().GetOSPFRouteMap(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("OSPFv2 Route Map Created but Read Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, ospfRouteMapFromAPI(remote, id))...)
}
func (r *ospfRouteMapResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ospfRouteMapModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	remote, err := r.client.Quagga().GetOSPFRouteMap(ctx, data.ID.ValueString())
	if err != nil {
		var nf *errs.NotFoundError
		if errors.As(err, &nf) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to Read OSPFv2 Route Map", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, ospfRouteMapFromAPI(remote, data.ID.ValueString()))...)
}
func (r *ospfRouteMapResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ospfRouteMapModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.Quagga().UpdateOSPFRouteMap(ctx, data.ID.ValueString(), ospfRouteMapToAPI(&data)); err != nil {
		resp.Diagnostics.AddError("Unable to Update OSPFv2 Route Map", err.Error())
		return
	}
	remote, err := r.client.Quagga().GetOSPFRouteMap(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("OSPFv2 Route Map Updated but Read Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, ospfRouteMapFromAPI(remote, data.ID.ValueString()))...)
}
func (r *ospfRouteMapResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ospfRouteMapModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.Quagga().DeleteOSPFRouteMap(ctx, data.ID.ValueString()); err != nil {
		var nf *errs.NotFoundError
		if !errors.As(err, &nf) {
			resp.Diagnostics.AddError("Unable to Delete OSPFv2 Route Map", err.Error())
		}
	}
}
func (r *ospfRouteMapResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
