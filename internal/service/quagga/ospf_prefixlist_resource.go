package quagga

import (
	"context"
	"errors"

	"github.com/biptec/opnsense-go/pkg/errs"
	"github.com/biptec/opnsense-go/pkg/opnsense"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ resource.Resource = &ospfPrefixListResource{}
var _ resource.ResourceWithConfigure = &ospfPrefixListResource{}
var _ resource.ResourceWithImportState = &ospfPrefixListResource{}

type ospfPrefixListResource struct{ client opnsense.Client }

func newOspfPrefixListResource() resource.Resource { return &ospfPrefixListResource{} }
func (r *ospfPrefixListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quagga_ospf_prefixlist"
}
func (r *ospfPrefixListResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = ospfPrefixListResourceSchema()
}
func (r *ospfPrefixListResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureQuaggaResource(req, resp)
}
func (r *ospfPrefixListResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ospfPrefixListModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := r.client.Quagga().AddOSPFPrefixList(ctx, ospfPrefixListToAPI(&data))
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create OSPFv2 Prefix List", err.Error())
		return
	}
	remote, err := r.client.Quagga().GetOSPFPrefixList(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("OSPFv2 Prefix List Created but Read Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, ospfPrefixListFromAPI(remote, id))...)
}
func (r *ospfPrefixListResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ospfPrefixListModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	remote, err := r.client.Quagga().GetOSPFPrefixList(ctx, data.ID.ValueString())
	if err != nil {
		var nf *errs.NotFoundError
		if errors.As(err, &nf) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to Read OSPFv2 Prefix List", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, ospfPrefixListFromAPI(remote, data.ID.ValueString()))...)
}
func (r *ospfPrefixListResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ospfPrefixListModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.Quagga().UpdateOSPFPrefixList(ctx, data.ID.ValueString(), ospfPrefixListToAPI(&data)); err != nil {
		resp.Diagnostics.AddError("Unable to Update OSPFv2 Prefix List", err.Error())
		return
	}
	remote, err := r.client.Quagga().GetOSPFPrefixList(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("OSPFv2 Prefix List Updated but Read Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, ospfPrefixListFromAPI(remote, data.ID.ValueString()))...)
}
func (r *ospfPrefixListResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ospfPrefixListModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := unlinkOSPFPrefixListFromRouteMaps(ctx, r.client, data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Unable to Unlink OSPFv2 Prefix List", err.Error())
		return
	}
	if err := r.client.Quagga().DeleteOSPFPrefixList(ctx, data.ID.ValueString()); err != nil {
		var nf *errs.NotFoundError
		if !errors.As(err, &nf) {
			resp.Diagnostics.AddError("Unable to Delete OSPFv2 Prefix List", err.Error())
		}
	}
}
func (r *ospfPrefixListResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
