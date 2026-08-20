package quagga

import (
	"context"
	"errors"
	"github.com/biptec/opnsense-go/pkg/errs"
	"github.com/biptec/opnsense-go/pkg/opnsense"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ resource.Resource = &ospfRedistributionResource{}
var _ resource.ResourceWithConfigure = &ospfRedistributionResource{}
var _ resource.ResourceWithImportState = &ospfRedistributionResource{}

type ospfRedistributionResource struct{ client opnsense.Client }

func newOspfRedistributionResource() resource.Resource { return &ospfRedistributionResource{} }
func (r *ospfRedistributionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quagga_ospf_redistribution"
}
func (r *ospfRedistributionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = ospfRedistributionResourceSchema()
}
func (r *ospfRedistributionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureQuaggaResource(req, resp)
}
func (r *ospfRedistributionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ospfRedistributionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := r.client.Quagga().AddOSPFRedistribution(ctx, ospfRedistributionToAPI(&data))
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create OSPFv2 Redistribution", err.Error())
		return
	}
	remote, err := r.client.Quagga().GetOSPFRedistribution(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("OSPFv2 Redistribution Created but Read Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, ospfRedistributionFromAPI(remote, id))...)
}
func (r *ospfRedistributionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ospfRedistributionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	remote, err := r.client.Quagga().GetOSPFRedistribution(ctx, data.ID.ValueString())
	if err != nil {
		var nf *errs.NotFoundError
		if errors.As(err, &nf) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to Read OSPFv2 Redistribution", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, ospfRedistributionFromAPI(remote, data.ID.ValueString()))...)
}
func (r *ospfRedistributionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ospfRedistributionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.Quagga().UpdateOSPFRedistribution(ctx, data.ID.ValueString(), ospfRedistributionToAPI(&data)); err != nil {
		resp.Diagnostics.AddError("Unable to Update OSPFv2 Redistribution", err.Error())
		return
	}
	remote, err := r.client.Quagga().GetOSPFRedistribution(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("OSPFv2 Redistribution Updated but Read Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, ospfRedistributionFromAPI(remote, data.ID.ValueString()))...)
}
func (r *ospfRedistributionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ospfRedistributionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.Quagga().DeleteOSPFRedistribution(ctx, data.ID.ValueString()); err != nil {
		var nf *errs.NotFoundError
		if !errors.As(err, &nf) {
			resp.Diagnostics.AddError("Unable to Delete OSPFv2 Redistribution", err.Error())
		}
	}
}
func (r *ospfRedistributionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
