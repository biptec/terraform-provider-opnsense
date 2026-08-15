package acmeclient

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &actionResource{}
var _ resource.ResourceWithConfigure = &actionResource{}
var _ resource.ResourceWithImportState = &actionResource{}

type actionResource struct{ resourceClient }

func newActionResource() resource.Resource { return &actionResource{} }
func (r *actionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_acme_action"
}
func (r *actionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = actionResourceSchema()
}
func (r *actionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan actionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := r.client.Acmeclient().AddAction(ctx, actionToAPI(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create ACME Action", err.Error())
		return
	}
	remote, err := r.client.Acmeclient().GetAction(ctx, id)
	if err != nil {
		plan.ID = types.StringValue(id)
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		resp.Diagnostics.AddError("ACME Action Created but Read Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, actionFromAPI(id, remote))...)
}
func (r *actionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var old actionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &old)...)
	if resp.Diagnostics.HasError() {
		return
	}
	remote, err := r.client.Acmeclient().GetAction(ctx, old.ID.ValueString())
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to Read ACME Action", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, actionFromAPI(old.ID.ValueString(), remote))...)
}
func (r *actionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, old actionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &old)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.Acmeclient().UpdateAction(ctx, old.ID.ValueString(), actionToAPI(&plan)); err != nil {
		resp.Diagnostics.AddError("Unable to Update ACME Action", err.Error())
		return
	}
	remote, err := r.client.Acmeclient().GetAction(ctx, old.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("ACME Action Updated but Read Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, actionFromAPI(old.ID.ValueString(), remote))...)
}
func (r *actionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state actionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.Acmeclient().DeleteAction(ctx, state.ID.ValueString()); err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Unable to Delete ACME Action", err.Error())
	}
}
func (r *actionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
