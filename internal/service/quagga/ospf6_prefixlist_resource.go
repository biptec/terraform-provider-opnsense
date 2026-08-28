package quagga

import (
	"context"
	"errors"

	"github.com/biptec/opnsense-go/pkg/errs"
	"github.com/biptec/opnsense-go/pkg/opnsense"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ resource.Resource = &ospf6PrefixListResource{}
var _ resource.ResourceWithConfigure = &ospf6PrefixListResource{}
var _ resource.ResourceWithImportState = &ospf6PrefixListResource{}

type ospf6PrefixListResource struct{ client opnsense.Client }

func newOspf6PrefixListResource() resource.Resource { return &ospf6PrefixListResource{} }
func (r *ospf6PrefixListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quagga_ospf6_prefixlist"
}
func (r *ospf6PrefixListResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = ospf6PrefixListResourceSchema()
}
func (r *ospf6PrefixListResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureQuaggaResource(req, resp)
}
func (r *ospf6PrefixListResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ospf6PrefixListModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := r.client.Quagga().AddOSPF6PrefixList(ctx, ospf6PrefixListToAPI(&data))
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create OSPFv3 Prefix List", err.Error())
		return
	}
	remote, err := r.client.Quagga().GetOSPF6PrefixList(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("OSPFv3 Prefix List Created but Read Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, ospf6PrefixListFromAPI(remote, id))...)
}
func (r *ospf6PrefixListResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ospf6PrefixListModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	remote, err := r.client.Quagga().GetOSPF6PrefixList(ctx, data.ID.ValueString())
	if err != nil {
		var nf *errs.NotFoundError
		if errors.As(err, &nf) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to Read OSPFv3 Prefix List", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, ospf6PrefixListFromAPI(remote, data.ID.ValueString()))...)
}
func (r *ospf6PrefixListResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ospf6PrefixListModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.Quagga().UpdateOSPF6PrefixList(ctx, data.ID.ValueString(), ospf6PrefixListToAPI(&data)); err != nil {
		resp.Diagnostics.AddError("Unable to Update OSPFv3 Prefix List", err.Error())
		return
	}
	remote, err := r.client.Quagga().GetOSPF6PrefixList(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("OSPFv3 Prefix List Updated but Read Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, ospf6PrefixListFromAPI(remote, data.ID.ValueString()))...)
}
func (r *ospf6PrefixListResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ospf6PrefixListModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := unlinkOSPF6PrefixListFromRouteMaps(ctx, r.client, data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Unable to Unlink OSPFv3 Prefix List", err.Error())
		return
	}
	if err := r.client.Quagga().DeleteOSPF6PrefixList(ctx, data.ID.ValueString()); err != nil {
		var nf *errs.NotFoundError
		if !errors.As(err, &nf) {
			resp.Diagnostics.AddError("Unable to Delete OSPFv3 Prefix List", err.Error())
		}
	}
}
func (r *ospf6PrefixListResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
