package system

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

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

type haproxyTLSIdentitySearchResponse struct {
	Total   *int                 `json:"total"`
	Rows    *[]map[string]string `json:"rows"`
	Message string               `json:"message"`
}

func validateHAProxyTLSIdentitySearchResponse(result *haproxyTLSIdentitySearchResponse) error {
	if result == nil {
		return fmt.Errorf("empty search response; cannot safely reattach")
	}
	if message := strings.TrimSpace(result.Message); message != "" {
		return fmt.Errorf("search endpoint returned an error: %s", message)
	}
	if result.Total == nil || result.Rows == nil {
		return fmt.Errorf("search response is missing total or rows; cannot safely reattach")
	}
	if *result.Total != len(*result.Rows) {
		return fmt.Errorf("incomplete search response (total=%d rows=%d); cannot safely reattach", *result.Total, len(*result.Rows))
	}
	return nil
}

// HAProxy TLS identity reconciliation is global to the Router, not scoped to
// one Terraform resource instance. Serialize identity lifecycle operations so
// concurrent Create calls cannot run overlapping global materializations.
var haproxyTLSIdentityLifecycleMu sync.Mutex

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

func (r *haproxyTLSIdentityResource) preflight(ctx context.Context, data *haproxyTLSIdentityResourceModel) error {
	result, err := r.client.ApiExtensions().HAProxyTLSIdentityServicePreflight(ctx, haproxyTLSIdentityToAPI(data))
	if err != nil {
		return fmt.Errorf("preflight FreeIPA DNS and service principal for %q: %w", data.DNSName.ValueString(), err)
	}
	if result == nil {
		return fmt.Errorf("preflight FreeIPA DNS and service principal for %q: empty response", data.DNSName.ValueString())
	}
	if !strings.EqualFold(result.Status, "ok") {
		return fmt.Errorf(
			"preflight FreeIPA DNS and service principal for %q: status=%q message=%q",
			data.DNSName.ValueString(),
			result.Status,
			result.Message,
		)
	}
	return nil
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
	haproxyTLSIdentityLifecycleMu.Lock()
	defer haproxyTLSIdentityLifecycleMu.Unlock()

	var data haproxyTLSIdentityResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// A previous Create may have committed the remote registration before its
	// response/state write failed. Reattach only an exact frontend + DNS +
	// policy match; ambiguous or differently-owned matches fail closed.
	id, err := r.reattachAfterCreate(ctx, &data)
	var notFound *errs.NotFoundError
	if err != nil && !errors.As(err, &notFound) {
		resp.Diagnostics.AddError("Unable to Check Existing HAProxy TLS Identity", err.Error())
		return
	}

	if errors.As(err, &notFound) {
		// FreeIPA DNS and the HAPROXY/<dns-name> service principal must exist
		// before creating a new remote registration. The reconciler repeats a
		// batch preflight only for identities that actually need certificate
		// issuance, so reattaching an already-materialized identity stays safe.
		if err := r.preflight(ctx, &data); err != nil {
			resp.Diagnostics.AddError("HAProxy TLS Identity Prerequisites Not Ready", err.Error())
			return
		}
		id, err = r.client.ApiExtensions().AddHAProxyTLSIdentity(ctx, haproxyTLSIdentityToAPI(&data))
		if err != nil || id == "" {
			// The add endpoint can commit the registration before the response is
			// lost or an after-save operation reports failure. Search again before
			// returning so the next retry cannot create a duplicate or leave an
			// untracked remote identity.
			reattachedID, reattachErr := r.reattachAfterCreate(ctx, &data)
			if reattachErr != nil {
				var duplicateNotFound *errs.NotFoundError
				if err != nil && errors.As(reattachErr, &duplicateNotFound) {
					resp.Diagnostics.AddError("Unable to Create HAProxy TLS Identity", err.Error())
					return
				}
				if err != nil {
					resp.Diagnostics.AddError(
						"Unable to Create or Reattach HAProxy TLS Identity",
						fmt.Sprintf("create failed: %v; exact reattach failed: %v", err, reattachErr),
					)
				} else {
					resp.Diagnostics.AddError("Unable to Reattach HAProxy TLS Identity", reattachErr.Error())
				}
				return
			}
			id = reattachedID
		}
	}

	remote, readErr := r.client.ApiExtensions().GetHAProxyTLSIdentity(ctx, id)
	if readErr != nil {
		originalReadErr := readErr
		reattachedID, reattachErr := r.reattachAfterCreate(ctx, &data)
		if reattachErr != nil {
			resp.Diagnostics.AddError(
				"HAProxy TLS Identity Created but Read Failed",
				fmt.Sprintf("read by returned UUID failed: %v; exact reattach failed: %v", originalReadErr, reattachErr),
			)
			return
		}
		id = reattachedID
		remote, readErr = r.client.ApiExtensions().GetHAProxyTLSIdentity(ctx, id)
		if readErr != nil {
			resp.Diagnostics.AddError(
				"HAProxy TLS Identity Created but Read Failed",
				fmt.Sprintf("read by returned UUID failed: %v; read by reattached UUID %q failed: %v", originalReadErr, id, readErr),
			)
			return
		}
	}

	if remote.FrontendName != data.FrontendName.ValueString() ||
		!strings.EqualFold(strings.TrimSuffix(remote.DNSName, "."), strings.TrimSuffix(data.DNSName.ValueString(), ".")) ||
		remote.PolicyID.String() != data.PolicyID.ValueString() {
		resp.Diagnostics.AddError(
			"HAProxy TLS Identity Create Returned a Different Object",
			"the remote identity does not match the requested frontend_name, dns_name and policy_id; refusing to adopt it",
		)
		return
	}

	// Persist the remote registration before certificate materialization. If
	// reconciliation fails and Terraform cannot commit this response state,
	// the next Create searches and reattaches the exact semantic identity above.
	resp.Diagnostics.Append(resp.State.Set(ctx, haproxyTLSIdentityFromAPI(remote, id))...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.reconfigure(ctx); err != nil {
		resp.Diagnostics.AddError("HAProxy TLS Identity Registered but Reconciliation Failed", err.Error())
	}
}

func exactHAProxyTLSIdentityReplacement(
	prior *haproxyTLSIdentityResourceModel,
	rows []map[string]string,
) (string, error) {
	if prior.FrontendName.IsNull() || prior.FrontendName.IsUnknown() || prior.FrontendName.ValueString() == "" ||
		prior.DNSName.IsNull() || prior.DNSName.IsUnknown() || prior.DNSName.ValueString() == "" ||
		prior.PolicyID.IsNull() || prior.PolicyID.IsUnknown() || prior.PolicyID.ValueString() == "" {
		return "", fmt.Errorf("prior HAProxy TLS identity lacks stable frontend_name, dns_name and policy_id")
	}

	frontend := prior.FrontendName.ValueString()
	dns := strings.ToLower(strings.TrimSuffix(prior.DNSName.ValueString(), "."))
	policy := prior.PolicyID.ValueString()
	matches := []string{}
	for _, row := range rows {
		if row["frontend_name"] != frontend ||
			strings.ToLower(strings.TrimSuffix(row["dns_name"], ".")) != dns {
			continue
		}
		if row["policy_id"] != policy {
			return "", fmt.Errorf(
				"matching HAProxy TLS identity has policy_id %q, expected %q; refusing to reattach a different owner policy",
				row["policy_id"], policy,
			)
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

func exactHAProxyTLSIdentityCreateMatch(
	wanted *haproxyTLSIdentityResourceModel,
	rows []map[string]string,
) (string, error) {
	if wanted.FrontendName.IsNull() || wanted.FrontendName.IsUnknown() || wanted.FrontendName.ValueString() == "" ||
		wanted.DNSName.IsNull() || wanted.DNSName.IsUnknown() || wanted.DNSName.ValueString() == "" ||
		wanted.PolicyID.IsNull() || wanted.PolicyID.IsUnknown() || wanted.PolicyID.ValueString() == "" {
		return "", fmt.Errorf("planned HAProxy TLS identity lacks stable frontend_name, dns_name and policy_id")
	}

	frontend := wanted.FrontendName.ValueString()
	dns := strings.ToLower(strings.TrimSuffix(wanted.DNSName.ValueString(), "."))
	policy := wanted.PolicyID.ValueString()
	matches := []string{}
	for _, row := range rows {
		if row["frontend_name"] != frontend ||
			strings.ToLower(strings.TrimSuffix(row["dns_name"], ".")) != dns {
			continue
		}
		if row["policy_id"] != policy {
			return "", fmt.Errorf(
				"matching HAProxy TLS identity has policy_id %q, expected %q; refusing to adopt a different owner policy",
				row["policy_id"], policy,
			)
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

func (r *haproxyTLSIdentityResource) searchIdentities(ctx context.Context) (*haproxyTLSIdentitySearchResponse, error) {
	result, err := api.Call(
		r.client.ApiExtensions().Client(),
		ctx,
		api.RPCOpts{Endpoint: apiextensions.HAProxyTLSIdentityOpts.Search},
		&haproxyTLSIdentitySearchResponse{},
	)
	if err != nil {
		return nil, fmt.Errorf("search HAProxy TLS identities: %w", err)
	}
	if err := validateHAProxyTLSIdentitySearchResponse(result); err != nil {
		return nil, fmt.Errorf("search HAProxy TLS identities: %w", err)
	}
	return result, nil
}

func (r *haproxyTLSIdentityResource) reattachAfterCreate(ctx context.Context, wanted *haproxyTLSIdentityResourceModel) (string, error) {
	result, err := r.searchIdentities(ctx)
	if err != nil {
		return "", fmt.Errorf("search HAProxy TLS identities after create failure: %w", err)
	}
	return exactHAProxyTLSIdentityCreateMatch(wanted, *result.Rows)
}

func (r *haproxyTLSIdentityResource) reattach(ctx context.Context, prior *haproxyTLSIdentityResourceModel) (string, error) {
	result, err := r.searchIdentities(ctx)
	if err != nil {
		return "", err
	}
	return exactHAProxyTLSIdentityReplacement(prior, *result.Rows)
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
	haproxyTLSIdentityLifecycleMu.Lock()
	defer haproxyTLSIdentityLifecycleMu.Unlock()
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
	haproxyTLSIdentityLifecycleMu.Lock()
	defer haproxyTLSIdentityLifecycleMu.Unlock()
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
