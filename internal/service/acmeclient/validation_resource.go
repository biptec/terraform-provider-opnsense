package acmeclient

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &validationResource{}
var _ resource.ResourceWithConfigure = &validationResource{}
var _ resource.ResourceWithImportState = &validationResource{}

type validationResource struct{ resourceClient }

func newValidationResource() resource.Resource { return &validationResource{} }
func (r *validationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_acme_validation"
}
func (r *validationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = validationResourceSchema()
}
func (r *validationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan validationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	var key types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("dns_nsupdate_key"), &key)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if key.IsNull() || key.IsUnknown() || key.ValueString() == "" {
		resp.Diagnostics.AddError("Missing RFC2136 Key", "dns_nsupdate_key is required when creating an ACME nsupdate validation.")
		return
	}
	id, err := r.client.Acmeclient().AddValidation(ctx, validationToAPI(&plan, key.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create ACME Validation", err.Error())
		return
	}
	remote, err := r.client.Acmeclient().GetValidation(ctx, id)
	if err != nil {
		plan.ID = types.StringValue(id)
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		resp.Diagnostics.AddError("ACME Validation Created but Read Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, validationFromAPI(id, remote, &plan))...)
}
func (r *validationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var old validationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &old)...)
	if resp.Diagnostics.HasError() {
		return
	}
	remote, err := r.client.Acmeclient().GetValidation(ctx, old.ID.ValueString())
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to Read ACME Validation", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, validationFromAPI(old.ID.ValueString(), remote, &old))...)
}
func (r *validationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, old validationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &old)...)
	var key types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("dns_nsupdate_key"), &key)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if plan.DNSNsupdateKeyVersion.ValueInt64() != old.DNSNsupdateKeyVersion.ValueInt64() && (key.IsNull() || key.IsUnknown() || key.ValueString() == "") {
		resp.Diagnostics.AddError("Missing Rotated RFC2136 Key", "dns_nsupdate_key_version changed but no write-only dns_nsupdate_key was supplied.")
		return
	}
	remote, err := r.client.Acmeclient().GetValidation(ctx, old.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read ACME Validation", err.Error())
		return
	}
	currentKey := remote.DNSNsupdateKey
	if !key.IsNull() && !key.IsUnknown() && key.ValueString() != "" {
		currentKey = key.ValueString()
	}
	desired := validationToAPI(&plan, currentKey)
	if err := r.client.Acmeclient().UpdateValidation(ctx, old.ID.ValueString(), desired); err != nil {
		resp.Diagnostics.AddError("Unable to Update ACME Validation", err.Error())
		return
	}
	updated, err := r.client.Acmeclient().GetValidation(ctx, old.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("ACME Validation Updated but Read Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, validationFromAPI(old.ID.ValueString(), updated, &plan))...)
}
func (r *validationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state validationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.Acmeclient().DeleteValidation(ctx, state.ID.ValueString()); err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Unable to Delete ACME Validation", err.Error())
	}
}
func (r *validationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("dns_nsupdate_key_version"), int64(0))...)
}
