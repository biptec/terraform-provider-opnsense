package system

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/biptec/opnsense-go/pkg/api"
	apiextensions "github.com/biptec/opnsense-go/pkg/api_extensions"
	"github.com/biptec/opnsense-go/pkg/errs"
	"github.com/biptec/opnsense-go/pkg/opnsense"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &haproxyTLSIdentityResource{}
var _ resource.ResourceWithConfigure = &haproxyTLSIdentityResource{}
var _ resource.ResourceWithImportState = &haproxyTLSIdentityResource{}

type haproxyTLSIdentityResource struct{ client opnsense.Client }

func newHAProxyTLSIdentityResource() resource.Resource {
	return &haproxyTLSIdentityResource{}
}

func (r *haproxyTLSIdentityResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_haproxy_tls_identity"
}

func (r *haproxyTLSIdentityResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = haproxyTLSIdentityResourceSchema()
}

func (r *haproxyTLSIdentityResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureClient(req, resp)
}

func haproxyTLSIdentityToAPI(data *haproxyTLSIdentityResourceModel) *apiextensions.HAProxyTLSIdentity {
	return &apiextensions.HAProxyTLSIdentity{
		FrontendName: data.FrontendName.ValueString(),
		DNSName:      data.DNSName.ValueString(),
		PolicyID:     api.SelectedMap(data.PolicyID.ValueString()),
	}
}

func haproxyTLSIdentityFromAPI(remote *apiextensions.HAProxyTLSIdentity, id string) *haproxyTLSIdentityResourceModel {
	return &haproxyTLSIdentityResourceModel{
		FrontendName: types.StringValue(remote.FrontendName),
		DNSName:      types.StringValue(strings.ToLower(strings.TrimSuffix(remote.DNSName, "."))),
		PolicyID:     types.StringValue(remote.PolicyID.String()),
		ID:           types.StringValue(id),
	}
}

func (r *haproxyTLSIdentityResource) reconfigure(ctx context.Context) error {
	result, err := r.client.ApiExtensions().HAProxyTLSIdentityServiceReconfigure(ctx)
	if err != nil {
		return fmt.Errorf("reconfigure Router-local HAProxy TLS identities: %w", err)
	}
	if result == nil {
		return fmt.Errorf("reconfigure Router-local HAProxy TLS identities: empty response")
	}
	if !strings.EqualFold(result.Status, "ok") {
		return fmt.Errorf(
			"reconfigure Router-local HAProxy TLS identities: status=%q message=%q",
			result.Status,
			result.Message,
		)
	}
	return nil
}

func (r *haproxyTLSIdentityResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data haproxyTLSIdentityResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := r.client.ApiExtensions().AddHAProxyTLSIdentity(ctx, haproxyTLSIdentityToAPI(&data))
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create HAProxy TLS Identity", err.Error())
		return
	}
	remote, err := r.client.ApiExtensions().GetHAProxyTLSIdentity(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("HAProxy TLS Identity Created but Read Failed", err.Error())
		return
	}

	// Persist remote identity before certificate issuance. If FreeIPA/OCSP or
	// HAProxy runtime reconciliation fails, Terraform retains enough state to
	// retry the same declarative object instead of creating an untracked one.
	resp.Diagnostics.Append(resp.State.Set(ctx, haproxyTLSIdentityFromAPI(remote, id))...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.reconfigure(ctx); err != nil {
		resp.Diagnostics.AddError("HAProxy TLS Identity Created but Reconciliation Failed", err.Error())
	}
}

func exactHAProxyTLSIdentityReplacement(
	prior *haproxyTLSIdentityResourceModel,
	rows []map[string]string,
) (string, error) {
	if prior.FrontendName.IsNull() || prior.FrontendName.IsUnknown() || prior.FrontendName.ValueString() == "" ||
		prior.DNSName.IsNull() || prior.DNSName.IsUnknown() || prior.DNSName.ValueString() == "" {
		return "", fmt.Errorf("prior HAProxy TLS identity lacks stable frontend_name and dns_name")
	}

	frontend := prior.FrontendName.ValueString()
	dns := strings.ToLower(strings.TrimSuffix(prior.DNSName.ValueString(), "."))
	matches := []string{}
	for _, row := range rows {
		if row["frontend_name"] != frontend ||
			strings.ToLower(strings.TrimSuffix(row["dns_name"], ".")) != dns {
			continue
		}
		if row["uuid"] == "" {
			return "", fmt.Errorf("matching HAProxy TLS identity has no UUID")
		}
		matches = append(matches, row["uuid"])
	}
	if len(matches) == 0 {
		return "", errs.NewNotFoundError()
	}
	if len(matches) != 1 {
		return "", fmt.Errorf("expected exactly one matching HAProxy TLS identity, found %d", len(matches))
	}
	return matches[0], nil
}

func (r *haproxyTLSIdentityResource) reattach(ctx context.Context, prior *haproxyTLSIdentityResourceModel) (string, error) {
	result, err := api.Search[map[string]string](
		r.client.ApiExtensions().Client(),
		ctx,
		apiextensions.HAProxyTLSIdentityOpts.Search,
	)
	if err != nil {
		return "", fmt.Errorf("search HAProxy TLS identities: %w", err)
	}
	if result == nil || result.Total > len(result.Rows) {
		return "", fmt.Errorf("incomplete HAProxy TLS identity search response; cannot reattach")
	}
	return exactHAProxyTLSIdentityReplacement(prior, result.Rows)
}

func (r *haproxyTLSIdentityResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data haproxyTLSIdentityResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	remote, err := r.client.ApiExtensions().GetHAProxyTLSIdentity(ctx, data.ID.ValueString())
	if err != nil {
		var notFound *errs.NotFoundError
		if errors.As(err, &notFound) {
			id, reattachErr := r.reattach(ctx, &data)
			if reattachErr != nil {
				var replacementNotFound *errs.NotFoundError
				if errors.As(reattachErr, &replacementNotFound) {
					resp.State.RemoveResource(ctx)
					return
				}
				resp.Diagnostics.AddError("Unable to Reattach HAProxy TLS Identity", reattachErr.Error())
				return
			}
			remote, err = r.client.ApiExtensions().GetHAProxyTLSIdentity(ctx, id)
			if err != nil {
				resp.Diagnostics.AddError("Unable to Read Reattached HAProxy TLS Identity", err.Error())
				return
			}
			data.ID = types.StringValue(id)
		} else {
			resp.Diagnostics.AddError("Unable to Read HAProxy TLS Identity", err.Error())
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, haproxyTLSIdentityFromAPI(remote, data.ID.ValueString()))...)
}

func (r *haproxyTLSIdentityResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data haproxyTLSIdentityResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.ApiExtensions().UpdateHAProxyTLSIdentity(
		ctx,
		data.ID.ValueString(),
		haproxyTLSIdentityToAPI(&data),
	); err != nil {
		resp.Diagnostics.AddError("Unable to Update HAProxy TLS Identity", err.Error())
		return
	}
	remote, err := r.client.ApiExtensions().GetHAProxyTLSIdentity(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("HAProxy TLS Identity Updated but Read Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, haproxyTLSIdentityFromAPI(remote, data.ID.ValueString()))...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.reconfigure(ctx); err != nil {
		resp.Diagnostics.AddError("HAProxy TLS Identity Updated but Reconciliation Failed", err.Error())
	}
}

func (r *haproxyTLSIdentityResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data haproxyTLSIdentityResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.ApiExtensions().DeleteHAProxyTLSIdentity(ctx, data.ID.ValueString())
	if err != nil {
		var notFound *errs.NotFoundError
		if !errors.As(err, &notFound) {
			resp.Diagnostics.AddError("Unable to Delete HAProxy TLS Identity", err.Error())
			return
		}
	}

	// Reconcile even after a not-found delete: a prior partial delete may have
	// removed metadata while leaving local Trust/files/runtime membership.
	if err := r.reconfigure(ctx); err != nil {
		resp.Diagnostics.AddError("HAProxy TLS Identity Deleted but Cleanup Failed", err.Error())
	}
}

func (r *haproxyTLSIdentityResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
