package bind

import (
	"context"
	"errors"

	"github.com/biptec/opnsense-go/pkg/errs"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &tsigKeyResource{}
var _ resource.ResourceWithConfigure = &tsigKeyResource{}
var _ resource.ResourceWithImportState = &tsigKeyResource{}
var _ resource.ResourceWithUpgradeState = &tsigKeyResource{}

type tsigKeyResource struct{ resourceClient }

type tsigKeyResourceModelV0 struct {
	ID        types.String `tfsdk:"id"`
	Enabled   types.Bool   `tfsdk:"enabled"`
	Name      types.String `tfsdk:"name"`
	Algorithm types.String `tfsdk:"algorithm"`
	Secret    types.String `tfsdk:"secret"`
}

func newTsigKeyResource() resource.Resource { return &tsigKeyResource{} }

func (r *tsigKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bind_tsig_key"
}

func (r *tsigKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = tsigKeyResourceSchema()
}

func (r *tsigKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan tsigKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	var secret types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("secret"), &secret)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if secret.IsNull() || secret.IsUnknown() || secret.ValueString() == "" {
		resp.Diagnostics.AddError("Missing BIND TSIG Secret", "secret is required when creating a BIND TSIG key. Supply it through a write-only configuration value.")
		return
	}

	id, err := r.client.Bind().AddTsigKey(ctx, tsigKeyModelToAPI(&plan, secret.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create BIND TSIG Key", err.Error())
		return
	}
	remote, err := r.client.Bind().GetTsigKey(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("BIND TSIG Key Created but Read Failed", err.Error())
		return
	}
	state := tsigKeyAPIToModel(remote)
	state.ID = types.StringValue(id)
	state.SecretVersion = plan.SecretVersion
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *tsigKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var old tsigKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &old)...)
	if resp.Diagnostics.HasError() {
		return
	}
	remote, err := r.client.Bind().GetTsigKey(ctx, old.ID.ValueString())
	if err != nil {
		var notFound *errs.NotFoundError
		if errors.As(err, &notFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to Read BIND TSIG Key", err.Error())
		return
	}
	state := tsigKeyAPIToModel(remote)
	state.ID = old.ID
	state.SecretVersion = old.SecretVersion
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *tsigKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, old tsigKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &old)...)
	var secret types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("secret"), &secret)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !plan.SecretVersion.IsUnknown() && !old.SecretVersion.IsUnknown() && plan.SecretVersion.ValueInt64() != old.SecretVersion.ValueInt64() {
		if secret.IsNull() || secret.IsUnknown() || secret.ValueString() == "" {
			resp.Diagnostics.AddError("Missing Rotated BIND TSIG Secret", "secret_version changed but no write-only secret was supplied. Provide the new secret together with the incremented secret_version.")
			return
		}
	}

	remote, err := r.client.Bind().GetTsigKey(ctx, old.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read BIND TSIG Key", err.Error())
		return
	}
	applyTsigKeyModel(remote, &plan, secret)
	if err := r.client.Bind().UpdateTsigKey(ctx, old.ID.ValueString(), remote); err != nil {
		resp.Diagnostics.AddError("Unable to Update BIND TSIG Key", err.Error())
		return
	}
	updated, err := r.client.Bind().GetTsigKey(ctx, old.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("BIND TSIG Key Updated but Read Failed", err.Error())
		return
	}
	state := tsigKeyAPIToModel(updated)
	state.ID = old.ID
	state.SecretVersion = plan.SecretVersion
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *tsigKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state tsigKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.Bind().DeleteTsigKey(ctx, state.ID.ValueString()); err != nil {
		var notFound *errs.NotFoundError
		if errors.As(err, &notFound) {
			return
		}
		resp.Diagnostics.AddError("Unable to Delete BIND TSIG Key", err.Error())
	}
}

func (r *tsigKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("secret_version"), int64(0))...)
}

func (r *tsigKeyResource) UpgradeState(context.Context) map[int64]resource.StateUpgrader {
	prior := tsigKeyResourceSchemaV0()
	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema: &prior,
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				var old tsigKeyResourceModelV0
				resp.Diagnostics.Append(req.State.Get(ctx, &old)...)
				if resp.Diagnostics.HasError() {
					return
				}
				configured := !old.Secret.IsNull() && !old.Secret.IsUnknown() && old.Secret.ValueString() != ""
				resp.Diagnostics.Append(resp.State.Set(ctx, &tsigKeyResourceModel{
					ID:               old.ID,
					Enabled:          old.Enabled,
					Name:             old.Name,
					Algorithm:        old.Algorithm,
					Secret:           types.StringNull(),
					SecretVersion:    types.Int64Value(0),
					SecretConfigured: types.BoolValue(configured),
				})...)
			},
		},
	}
}
