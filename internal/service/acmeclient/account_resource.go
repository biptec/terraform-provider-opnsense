package acmeclient

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &accountResource{}
var _ resource.ResourceWithConfigure = &accountResource{}
var _ resource.ResourceWithImportState = &accountResource{}

type accountResource struct{ resourceClient }

func newAccountResource() resource.Resource { return &accountResource{} }
func (r *accountResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_acme_account"
}
func (r *accountResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = accountResourceSchema()
}
func (r *accountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan accountResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := r.client.Acmeclient().AddAccount(ctx, accountToAPI(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create ACME Account", err.Error())
		return
	}
	remote, err := r.client.Acmeclient().GetAccount(ctx, id)
	if err != nil {
		plan.ID = types.StringValue(id)
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		resp.Diagnostics.AddError("ACME Account Created but Read Failed", err.Error())
		return
	}
	state := accountFromAPI(id, remote, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if plan.Register.ValueBool() {
		if _, err = r.client.Acmeclient().AccountRegister(ctx, id); err != nil {
			resp.Diagnostics.AddError("Unable to Register ACME Account", err.Error())
			return
		}
		remote, err = waitAccountRegistration(ctx, r.client.Acmeclient(), id)
		if err != nil {
			resp.Diagnostics.AddError("Unable to Register ACME Account", err.Error())
			return
		}
		state = accountFromAPI(id, remote, &plan)
		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	}
}
func (r *accountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var old accountResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &old)...)
	if resp.Diagnostics.HasError() {
		return
	}
	remote, err := r.client.Acmeclient().GetAccount(ctx, old.ID.ValueString())
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to Read ACME Account", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, accountFromAPI(old.ID.ValueString(), remote, &old))...)
}
func (r *accountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, old accountResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &old)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := old.ID.ValueString()
	if err := r.client.Acmeclient().UpdateAccount(ctx, id, accountToAPI(&plan)); err != nil {
		resp.Diagnostics.AddError("Unable to Update ACME Account", err.Error())
		return
	}
	remote, err := r.client.Acmeclient().GetAccount(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("ACME Account Updated but Read Failed", err.Error())
		return
	}
	state := accountFromAPI(id, remote, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	shouldRegister := plan.Register.ValueBool() && (!old.Register.ValueBool() || plan.RegistrationVersion.ValueInt64() != old.RegistrationVersion.ValueInt64())
	if shouldRegister {
		if _, err = r.client.Acmeclient().AccountRegister(ctx, id); err != nil {
			resp.Diagnostics.AddError("Unable to Register ACME Account", err.Error())
			return
		}
		remote, err = waitAccountRegistration(ctx, r.client.Acmeclient(), id)
		if err != nil {
			resp.Diagnostics.AddError("Unable to Register ACME Account", err.Error())
			return
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, accountFromAPI(id, remote, &plan))...)
	}
}
func (r *accountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state accountResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.Acmeclient().DeleteAccount(ctx, state.ID.ValueString()); err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Unable to Delete ACME Account", err.Error())
	}
}
func (r *accountResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("register"), false)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("registration_version"), int64(0))...)
}
