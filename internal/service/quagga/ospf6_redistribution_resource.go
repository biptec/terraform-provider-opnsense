package quagga

import (
	"context"
	"errors"
	"github.com/biptec/opnsense-go/pkg/errs"
	"github.com/biptec/opnsense-go/pkg/opnsense"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ resource.Resource = &ospf6RedistributionResource{}
var _ resource.ResourceWithConfigure = &ospf6RedistributionResource{}
var _ resource.ResourceWithImportState = &ospf6RedistributionResource{}

type ospf6RedistributionResource struct{ client opnsense.Client }

func newOspf6RedistributionResource() resource.Resource { return &ospf6RedistributionResource{} }
func (r *ospf6RedistributionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quagga_ospf6_redistribution"
}
func (r *ospf6RedistributionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = ospf6RedistributionResourceSchema()
}
func (r *ospf6RedistributionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureQuaggaResource(req, resp)
}
func (r *ospf6RedistributionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ospf6RedistributionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := r.client.Quagga().AddOSPF6Redistribution(ctx, ospf6RedistributionToAPI(&data))
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create OSPFv3 Redistribution", err.Error())
		return
	}
	remote, err := r.client.Quagga().GetOSPF6Redistribution(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("OSPFv3 Redistribution Created but Read Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, ospf6RedistributionFromAPI(remote, id))...)
}
func (r *ospf6RedistributionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ospf6RedistributionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	remote, err := r.client.Quagga().GetOSPF6Redistribution(ctx, data.ID.ValueString())
	if err != nil {
		var nf *errs.NotFoundError
		if errors.As(err, &nf) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to Read OSPFv3 Redistribution", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, ospf6RedistributionFromAPI(remote, data.ID.ValueString()))...)
}
func (r *ospf6RedistributionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ospf6RedistributionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.Quagga().UpdateOSPF6Redistribution(ctx, data.ID.ValueString(), ospf6RedistributionToAPI(&data)); err != nil {
		resp.Diagnostics.AddError("Unable to Update OSPFv3 Redistribution", err.Error())
		return
	}
	remote, err := r.client.Quagga().GetOSPF6Redistribution(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("OSPFv3 Redistribution Updated but Read Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, ospf6RedistributionFromAPI(remote, data.ID.ValueString()))...)
}
func (r *ospf6RedistributionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ospf6RedistributionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.Quagga().DeleteOSPF6Redistribution(ctx, data.ID.ValueString()); err != nil {
		var nf *errs.NotFoundError
		if !errors.As(err, &nf) {
			resp.Diagnostics.AddError("Unable to Delete OSPFv3 Redistribution", err.Error())
		}
	}
}
func (r *ospf6RedistributionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
