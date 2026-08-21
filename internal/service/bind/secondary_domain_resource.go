package bind

import (
	"context"
	"errors"
	"fmt"

	"github.com/biptec/opnsense-go/pkg/errs"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &secondaryDomainResource{}
var _ resource.ResourceWithConfigure = &secondaryDomainResource{}
var _ resource.ResourceWithImportState = &secondaryDomainResource{}
var _ resource.ResourceWithUpgradeState = &secondaryDomainResource{}
var _ resource.ResourceWithConfigValidators = &secondaryDomainResource{}

type secondaryDomainResource struct{ resourceClient }

func newSecondaryDomainResource() resource.Resource { return &secondaryDomainResource{} }

func (r *secondaryDomainResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bind_secondary_domain"
}

func (r *secondaryDomainResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = secondaryDomainResourceSchema()
}

func (r *secondaryDomainResource) ConfigValidators(context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{secondaryDomainTransferConfigValidator{}}
}

func (r *secondaryDomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan secondaryDomainResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	var secret types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("transfer_key"), &secret)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := validateSecondaryTransferSecret(&plan, secret, true, false); err != nil {
		resp.Diagnostics.AddError("Invalid BIND Secondary Transfer Key", err.Error())
		return
	}

	key := ""
	if !secret.IsNull() && !secret.IsUnknown() {
		key = secret.ValueString()
	}
	id, err := r.client.Bind().AddSecondaryDomain(ctx, secondaryDomainModelToAPI(&plan, key))
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create BIND Secondary Domain", err.Error())
		return
	}
	remote, err := r.client.Bind().GetSecondaryDomain(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("BIND Secondary Domain Created but Read Failed", err.Error())
		return
	}
	state := secondaryDomainAPIToModel(remote)
	state.ID = types.StringValue(id)
	state.TransferKeyVersion = plan.TransferKeyVersion
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *secondaryDomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var old secondaryDomainResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &old)...)
	if resp.Diagnostics.HasError() {
		return
	}
	remote, err := r.client.Bind().GetSecondaryDomain(ctx, old.ID.ValueString())
	if err != nil {
		var notFound *errs.NotFoundError
		if errors.As(err, &notFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to Read BIND Secondary Domain", err.Error())
		return
	}
	state := secondaryDomainAPIToModel(remote)
	state.ID = old.ID
	state.TransferKeyVersion = old.TransferKeyVersion
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *secondaryDomainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, old secondaryDomainResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &old)...)
	var secret types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("transfer_key"), &secret)...)
	if resp.Diagnostics.HasError() {
		return
	}
	rotated := !plan.TransferKeyVersion.IsUnknown() && !old.TransferKeyVersion.IsUnknown() && plan.TransferKeyVersion.ValueInt64() != old.TransferKeyVersion.ValueInt64()
	if err := validateSecondaryTransferSecret(&plan, secret, false, rotated); err != nil {
		resp.Diagnostics.AddError("Invalid BIND Secondary Transfer Key", err.Error())
		return
	}

	remote, err := r.client.Bind().GetSecondaryDomain(ctx, old.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read BIND Secondary Domain", err.Error())
		return
	}
	applySecondaryDomainModel(remote, &plan, secret)
	if legacyTransferMetadataConfigured(&plan) && remote.TransferKey == "" {
		resp.Diagnostics.AddError("Missing BIND Secondary Transfer Key", "authenticated transfer metadata is configured but OPNsense has no transfer TSIG secret; provide transfer_key and increment transfer_key_version")
		return
	}
	if err := r.client.Bind().UpdateSecondaryDomain(ctx, old.ID.ValueString(), remote); err != nil {
		resp.Diagnostics.AddError("Unable to Update BIND Secondary Domain", err.Error())
		return
	}
	updated, err := r.client.Bind().GetSecondaryDomain(ctx, old.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("BIND Secondary Domain Updated but Read Failed", err.Error())
		return
	}
	state := secondaryDomainAPIToModel(updated)
	state.ID = old.ID
	state.TransferKeyVersion = plan.TransferKeyVersion
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *secondaryDomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state secondaryDomainResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.Bind().DeleteSecondaryDomain(ctx, state.ID.ValueString()); err != nil {
		var notFound *errs.NotFoundError
		if errors.As(err, &notFound) {
			return
		}
		resp.Diagnostics.AddError("Unable to Delete BIND Secondary Domain", err.Error())
	}
}

func (r *secondaryDomainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("transfer_key_version"), int64(0))...)
}

func (r *secondaryDomainResource) UpgradeState(context.Context) map[int64]resource.StateUpgrader {
	prior := secondaryDomainResourceSchemaV0()
	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema: &prior,
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				var old secondaryDomainResourceModelV0
				resp.Diagnostics.Append(req.State.Get(ctx, &old)...)
				if resp.Diagnostics.HasError() {
					return
				}
				configured := !old.TransferKey.IsNull() && !old.TransferKey.IsUnknown() && old.TransferKey.ValueString() != ""
				resp.Diagnostics.Append(resp.State.Set(ctx, &secondaryDomainResourceModel{
					ID: old.ID, ViewID: old.ViewID, DomainName: old.DomainName, Enabled: old.Enabled,
					PrimaryIPs: old.PrimaryIPs, AllowNotify: old.AllowNotify, TransferKeyID: types.StringValue(""),
					TransferKeyAlgorithm: old.TransferKeyAlgorithm, TransferKeyName: old.TransferKeyName,
					TransferKey: types.StringNull(), TransferKeyVersion: types.Int64Value(0), TransferKeyConfigured: types.BoolValue(configured),
					AllowTransferACLs: old.AllowTransferACLs, AllowQueryACLs: old.AllowQueryACLs,
				})...)
			},
		},
	}
}

func legacyTransferMetadataConfigured(d *secondaryDomainResourceModel) bool {
	return d.TransferKeyAlgorithm.ValueString() != "" || d.TransferKeyName.ValueString() != ""
}

func validateSecondaryTransferSecret(d *secondaryDomainResourceModel, secret types.String, creating, rotated bool) error {
	sharedKeyID := d.TransferKeyID.ValueString()
	algorithm := d.TransferKeyAlgorithm.ValueString()
	name := d.TransferKeyName.ValueString()
	supplied := !secret.IsNull() && !secret.IsUnknown() && secret.ValueString() != ""
	if sharedKeyID != "" {
		if algorithm != "" || name != "" || supplied {
			return fmt.Errorf("transfer_key_id is mutually exclusive with transfer_key_algorithm, transfer_key_name, and transfer_key")
		}
		return nil
	}
	if (algorithm == "") != (name == "") {
		return fmt.Errorf("transfer_key_algorithm and transfer_key_name must either both be set or both be empty")
	}
	if algorithm == "" {
		if supplied {
			return fmt.Errorf("transfer_key requires transfer_key_algorithm and transfer_key_name")
		}
		return nil
	}
	if creating && !supplied {
		return fmt.Errorf("transfer_key is required when creating an authenticated secondary zone")
	}
	if rotated && !supplied {
		return fmt.Errorf("transfer_key_version changed but no write-only transfer_key was supplied")
	}
	return nil
}
