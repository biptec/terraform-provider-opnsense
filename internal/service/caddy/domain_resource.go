package caddy

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/biptec/opnsense-go/pkg/api"
	apicaddy "github.com/biptec/opnsense-go/pkg/caddy"
	"github.com/biptec/opnsense-go/pkg/errs"
	apitrust "github.com/biptec/opnsense-go/pkg/trust"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &domainResource{}
var _ resource.ResourceWithConfigure = &domainResource{}
var _ resource.ResourceWithImportState = &domainResource{}

type domainResource struct{ resourceClient }

func newDomainResource() resource.Resource { return &domainResource{} }
func (r *domainResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_caddy_domain"
}
func (r *domainResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = domainResourceSchema()
}

func caFieldMatches(c apitrust.Ca, name string) bool {
	return strings.EqualFold(c.CommonName, name) || strings.EqualFold(c.Description, name) || strings.EqualFold(c.Name, name) || strings.EqualFold(c.RefId, name)
}
func (r *domainResource) resolveCA(ctx context.Context, name string) (*apitrust.Ca, error) {
	result, err := r.client.Trust().SearchCa(ctx)
	if err != nil {
		return nil, fmt.Errorf("search existing CAs: %w", err)
	}
	matches := make([]apitrust.Ca, 0, 1)
	seen := map[string]bool{}
	for _, item := range result.Rows {
		if caFieldMatches(item, name) && !seen[item.UUID] {
			matches = append(matches, item)
			seen[item.UUID] = true
		}
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("no existing OPNsense CA exactly matches %q", name)
	}
	if len(matches) > 1 {
		return nil, fmt.Errorf("multiple existing OPNsense CAs match %q; use a unique common name, description, or ref_id", name)
	}
	ca := matches[0]
	if ca.UUID != "" {
		full, err := r.client.Trust().GetCa(ctx, ca.UUID)
		if err != nil {
			return nil, fmt.Errorf("read matching CA %s: %w", ca.UUID, err)
		}
		ca = *full
		ca.UUID = matches[0].UUID
	}
	if ca.RefId == "" {
		return nil, fmt.Errorf("matching CA %q has no ref_id", name)
	}
	return &ca, nil
}
func (r *domainResource) createInternalCertificate(ctx context.Context, d *domainResourceModel) (string, string, error) {
	ca, err := r.resolveCA(ctx, d.InternalCAName.ValueString())
	if err != nil {
		return "", "", err
	}
	cert := &apitrust.Cert{
		Description: "Caddy certificate for " + d.Domain.ValueString(), CaRef: api.SelectedMap(ca.RefId),
		Action: api.SelectedMap("internal"), KeyType: api.SelectedMap(d.InternalCertificateKeyType.ValueString()),
		Digest: api.SelectedMap(d.InternalCertificateDigest.ValueString()), CertType: api.SelectedMap("server_cert"),
		Lifetime:           fmt.Sprintf("%d", d.InternalCertificateLifetimeDays.ValueInt64()),
		PrivateKeyLocation: api.SelectedMap("firewall"), Country: ca.Country, State: ca.State, City: ca.City,
		Organization: ca.Organization, OrganizationalUnit: ca.OrganizationalUnit, Email: ca.Email,
		CommonName: d.Domain.ValueString(), AltnamesDns: d.Domain.ValueString(),
	}
	id, err := r.client.Trust().AddCert(ctx, cert)
	if err != nil {
		return id, "", fmt.Errorf("issue internal certificate: %w", err)
	}
	created, err := r.client.Trust().GetCert(ctx, id)
	if err != nil {
		return id, "", fmt.Errorf("read issued internal certificate: %w", err)
	}
	if created.RefId == "" {
		return id, "", fmt.Errorf("issued internal certificate %s has no ref_id", id)
	}
	return id, created.RefId, nil
}
func (r *domainResource) deleteManagedCertificate(ctx context.Context, id string) error {
	if id == "" {
		return nil
	}
	err := r.client.Trust().DeleteCert(ctx, id)
	if err != nil {
		var nf *errs.NotFoundError
		if errors.As(err, &nf) {
			return nil
		}
	}
	return err
}
func internalCertificateNeedsReplacement(plan, state *domainResourceModel) bool {
	if state.CertificateMode.ValueString() != "internal" || state.GeneratedCertificateID.ValueString() == "" {
		return true
	}
	return plan.Domain.ValueString() != state.Domain.ValueString() || plan.InternalCAName.ValueString() != state.InternalCAName.ValueString() || plan.InternalCertificateLifetimeDays.ValueInt64() != state.InternalCertificateLifetimeDays.ValueInt64() || plan.InternalCertificateKeyType.ValueString() != state.InternalCertificateKeyType.ValueString() || plan.InternalCertificateDigest.ValueString() != state.InternalCertificateDigest.ValueString()
}
func (r *domainResource) certificateForPlan(ctx context.Context, plan, state *domainResourceModel) (id, ref string, created bool, err error) {
	switch plan.CertificateMode.ValueString() {
	case "acme", "none":
		return "", "", false, nil
	case "custom":
		return "", plan.CertificateRefID.ValueString(), false, nil
	case "internal":
		if state != nil && !internalCertificateNeedsReplacement(plan, state) {
			ref = state.CertificateRefID.ValueString()
			if ref == "" {
				cert, getErr := r.client.Trust().GetCert(ctx, state.GeneratedCertificateID.ValueString())
				if getErr != nil {
					return "", "", false, fmt.Errorf("read managed internal certificate: %w", getErr)
				}
				ref = cert.RefId
			}
			return state.GeneratedCertificateID.ValueString(), ref, false, nil
		}
		id, ref, err = r.createInternalCertificate(ctx, plan)
		return id, ref, true, err
	default:
		return "", "", false, fmt.Errorf("unsupported certificate_mode %q", plan.CertificateMode.ValueString())
	}
}

func (r *domainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan domainResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := validateDomainModel(&plan); err != nil {
		resp.Diagnostics.AddError("Invalid Caddy Domain", err.Error())
		return
	}
	certID, certRef, certCreated, err := r.certificateForPlan(ctx, &plan, nil)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Prepare Caddy Certificate", err.Error())
		return
	}
	remotePlan, err := buildDomainAPI(&plan, certRef)
	if err != nil {
		if certCreated {
			_ = r.deleteManagedCertificate(ctx, certID)
		}
		resp.Diagnostics.AddError("Invalid Caddy Domain", err.Error())
		return
	}
	id, err := r.client.Caddy().AddDomain(ctx, remotePlan)
	if err != nil {
		plan.ID = types.StringValue(id)
		plan.GeneratedCertificateID = types.StringValue(certID)
		plan.CertificateRefID = types.StringValue(certRef)
		if id != "" {
			resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		} else if certCreated {
			_ = r.deleteManagedCertificate(ctx, certID)
		}
		resp.Diagnostics.AddError("Unable to Create Caddy Domain", err.Error())
		return
	}
	remote, err := r.client.Caddy().GetDomain(ctx, id)
	if err != nil {
		plan.ID = types.StringValue(id)
		plan.GeneratedCertificateID = types.StringValue(certID)
		plan.CertificateRefID = types.StringValue(certRef)
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		resp.Diagnostics.AddError("Caddy Domain Created but Read Failed", err.Error())
		return
	}
	state, err := domainStructToSchema(remote, &plan)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Decode Caddy Domain", err.Error())
		return
	}
	state.ID = types.StringValue(id)
	if certID != "" {
		state.GeneratedCertificateID = types.StringValue(certID)
	}
	state.CertificateRefID = types.StringValue(certRef)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	tflog.Trace(ctx, "created Caddy domain", map[string]any{"id": id, "certificate_mode": plan.CertificateMode.ValueString()})
}
func (r *domainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state domainResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	remote, err := r.client.Caddy().GetDomain(ctx, state.ID.ValueString())
	if err != nil {
		var nf *errs.NotFoundError
		if errors.As(err, &nf) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to Read Caddy Domain", err.Error())
		return
	}
	updated, err := domainStructToSchema(remote, &state)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Decode Caddy Domain", err.Error())
		return
	}
	updated.ID = state.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, updated)...)
}
func (r *domainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state domainResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := validateDomainModel(&plan); err != nil {
		resp.Diagnostics.AddError("Invalid Caddy Domain", err.Error())
		return
	}
	certID, certRef, certCreated, err := r.certificateForPlan(ctx, &plan, &state)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Prepare Caddy Certificate", err.Error())
		return
	}
	remotePlan, err := buildDomainAPI(&plan, certRef)
	if err != nil {
		if certCreated {
			_ = r.deleteManagedCertificate(ctx, certID)
		}
		resp.Diagnostics.AddError("Invalid Caddy Domain", err.Error())
		return
	}
	if err := r.client.Caddy().UpdateDomain(ctx, state.ID.ValueString(), remotePlan); err != nil {
		if certCreated {
			_ = r.deleteManagedCertificate(ctx, certID)
		}
		resp.Diagnostics.AddError("Unable to Update Caddy Domain", err.Error())
		return
	}
	oldCertID := state.GeneratedCertificateID.ValueString()
	if state.CertificateMode.ValueString() == "internal" && oldCertID != "" && (plan.CertificateMode.ValueString() != "internal" || certCreated) {
		if err := r.deleteManagedCertificate(ctx, oldCertID); err != nil {
			resp.Diagnostics.AddWarning("Old Internal Certificate Was Not Deleted", err.Error())
		}
	}
	remote, err := r.client.Caddy().GetDomain(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Caddy Domain Updated but Read Failed", err.Error())
		return
	}
	updated, err := domainStructToSchema(remote, &plan)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Decode Caddy Domain", err.Error())
		return
	}
	updated.ID = state.ID
	if plan.CertificateMode.ValueString() == "internal" {
		updated.GeneratedCertificateID = types.StringValue(certID)
		updated.CertificateRefID = types.StringValue(certRef)
	} else {
		updated.GeneratedCertificateID = types.StringNull()
		updated.CertificateRefID = types.StringValue(certRef)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, updated)...)
}
func (r *domainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state domainResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.client.Caddy().DeleteDomain(ctx, state.ID.ValueString())
	if err != nil {
		var nf *errs.NotFoundError
		if !errors.As(err, &nf) {
			resp.Diagnostics.AddError("Unable to Delete Caddy Domain", err.Error())
			return
		}
	}
	if state.CertificateMode.ValueString() == "internal" {
		if err := r.deleteManagedCertificate(ctx, state.GeneratedCertificateID.ValueString()); err != nil {
			resp.Diagnostics.AddError("Caddy Domain Deleted but Managed Certificate Delete Failed", err.Error())
			return
		}
	}
}
func (r *domainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

var _ = apicaddy.Domain{}
