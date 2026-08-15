package acmeclient

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &certificateResource{}
var _ resource.ResourceWithConfigure = &certificateResource{}
var _ resource.ResourceWithImportState = &certificateResource{}

type certificateResource struct{ resourceClient }

func newCertificateResource() resource.Resource { return &certificateResource{} }
func (r *certificateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_acme_certificate"
}
func (r *certificateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = certificateResourceSchema()
}
func (r *certificateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan certificateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiModel, err := certificateToAPI(ctx, &plan)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ACME Certificate", err.Error())
		return
	}
	id, err := r.client.Acmeclient().AddCertificate(ctx, apiModel)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create ACME Certificate", err.Error())
		return
	}
	remote, err := r.client.Acmeclient().GetCertificate(ctx, id)
	if err != nil {
		plan.ID = types.StringValue(id)
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		resp.Diagnostics.AddError("ACME Certificate Created but Read Failed", err.Error())
		return
	}
	state := certificateFromAPI(id, remote, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if plan.Issue.ValueBool() {
		if _, err = r.client.Acmeclient().CertificateSign(ctx, id); err != nil {
			resp.Diagnostics.AddError("Unable to Issue ACME Certificate", err.Error())
			return
		}
		remote, err = waitCertificateIssued(ctx, r.client.Acmeclient(), id)
		if err != nil {
			resp.Diagnostics.AddError("Unable to Issue ACME Certificate", err.Error())
			return
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, certificateFromAPI(id, remote, &plan))...)
	}
}
func (r *certificateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var old certificateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &old)...)
	if resp.Diagnostics.HasError() {
		return
	}
	remote, err := r.client.Acmeclient().GetCertificate(ctx, old.ID.ValueString())
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to Read ACME Certificate", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, certificateFromAPI(old.ID.ValueString(), remote, &old))...)
}
func (r *certificateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, old certificateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &old)...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiModel, err := certificateToAPI(ctx, &plan)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ACME Certificate", err.Error())
		return
	}
	id := old.ID.ValueString()
	if err := r.client.Acmeclient().UpdateCertificate(ctx, id, apiModel); err != nil {
		resp.Diagnostics.AddError("Unable to Update ACME Certificate", err.Error())
		return
	}
	remote, err := r.client.Acmeclient().GetCertificate(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("ACME Certificate Updated but Read Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, certificateFromAPI(id, remote, &plan))...)
	if resp.Diagnostics.HasError() {
		return
	}
	shouldIssue := plan.Issue.ValueBool() && (!old.Issue.ValueBool() || plan.IssuanceVersion.ValueInt64() != old.IssuanceVersion.ValueInt64())
	if shouldIssue {
		if _, err = r.client.Acmeclient().CertificateSign(ctx, id); err != nil {
			resp.Diagnostics.AddError("Unable to Issue ACME Certificate", err.Error())
			return
		}
		remote, err = waitCertificateIssued(ctx, r.client.Acmeclient(), id)
		if err != nil {
			resp.Diagnostics.AddError("Unable to Issue ACME Certificate", err.Error())
			return
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, certificateFromAPI(id, remote, &plan))...)
	}
}
func (r *certificateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state certificateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.Acmeclient().DeleteCertificate(ctx, state.ID.ValueString()); err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Unable to Delete ACME Certificate", err.Error())
	}
}
func (r *certificateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("issue"), false)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("issuance_version"), int64(0))...)
}
