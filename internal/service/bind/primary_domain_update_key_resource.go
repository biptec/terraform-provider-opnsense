package bind

import (
	"context"
	"errors"
	"slices"
	"strings"
	"sync"

	"github.com/biptec/opnsense-go/pkg/api"
	apibind "github.com/biptec/opnsense-go/pkg/bind"
	"github.com/biptec/opnsense-go/pkg/errs"
	"github.com/biptec/terraform-provider-opnsense/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var primaryDomainUpdateKeyMu sync.Mutex

var _ resource.Resource = &primaryDomainUpdateKeyResource{}
var _ resource.ResourceWithConfigure = &primaryDomainUpdateKeyResource{}
var _ resource.ResourceWithImportState = &primaryDomainUpdateKeyResource{}

type primaryDomainUpdateKeyResource struct{ resourceClient }

type primaryDomainUpdateKeyResourceModel struct {
	ID          types.String `tfsdk:"id"`
	DomainID    types.String `tfsdk:"domain_id"`
	UpdateKeyID types.String `tfsdk:"update_key_id"`
}

func newPrimaryDomainUpdateKeyResource() resource.Resource {
	return &primaryDomainUpdateKeyResource{}
}

func (r *primaryDomainUpdateKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bind_primary_domain_update_key"
}

func (r *primaryDomainUpdateKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Attaches one TSIG key to an existing BIND primary zone for RFC2136 updates without owning the zone or its other update keys. Delete removes only this membership.",
		Attributes: map[string]schema.Attribute{
			"id":            schema.StringAttribute{Computed: true, MarkdownDescription: "Attachment identity in `domain_id/update_key_id` form."},
			"domain_id":     schema.StringAttribute{Required: true, Validators: []validator.String{validators.IsUUIDv4()}, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}, MarkdownDescription: "UUID of the existing primary zone."},
			"update_key_id": schema.StringAttribute{Required: true, Validators: []validator.String{validators.IsUUIDv4()}, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}, MarkdownDescription: "UUID of the TSIG key to add to the zone update-key membership."},
		},
	}
}

func primaryDomainUpdateKeyID(domainID, updateKeyID string) string {
	return domainID + "/" + updateKeyID
}

func remoteUpdateKeyIDs(domain *apibind.PrimaryDomain) []string {
	keys := make([]string, 0, len(domain.UpdateKeys))
	for _, value := range domain.UpdateKeys {
		value = strings.TrimSpace(value)
		if value != "" {
			keys = append(keys, value)
		}
	}
	slices.Sort(keys)
	return slices.Compact(keys)
}

func hasUpdateKey(domain *apibind.PrimaryDomain, updateKeyID string) bool {
	return slices.Contains(remoteUpdateKeyIDs(domain), updateKeyID)
}

func addUpdateKey(domain *apibind.PrimaryDomain, updateKeyID string) {
	keys := remoteUpdateKeyIDs(domain)
	if !slices.Contains(keys, updateKeyID) {
		keys = append(keys, updateKeyID)
		slices.Sort(keys)
	}
	domain.UpdateKeys = api.SelectedMapList(keys)
}

func removeUpdateKey(domain *apibind.PrimaryDomain, updateKeyID string) {
	keys := remoteUpdateKeyIDs(domain)
	keys = slices.DeleteFunc(keys, func(value string) bool { return value == updateKeyID })
	domain.UpdateKeys = api.SelectedMapList(keys)
}

func (r *primaryDomainUpdateKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan primaryDomainUpdateKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	primaryDomainUpdateKeyMu.Lock()
	defer primaryDomainUpdateKeyMu.Unlock()

	domainID := plan.DomainID.ValueString()
	updateKeyID := plan.UpdateKeyID.ValueString()
	domain, err := r.client.Bind().GetPrimaryDomain(ctx, domainID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read BIND Primary Domain", err.Error())
		return
	}
	if hasUpdateKey(domain, updateKeyID) {
		resp.Diagnostics.AddError("BIND Primary Domain Update Key Already Attached", "The primary zone already contains this update-key membership. Refusing to adopt a relation that may belong to another Terraform state; import it explicitly if this state is the intended owner.")
		return
	}
	addUpdateKey(domain, updateKeyID)
	if err := r.client.Bind().UpdatePrimaryDomain(ctx, domainID, domain); err != nil {
		resp.Diagnostics.AddError("Unable to Attach BIND Primary Domain Update Key", err.Error())
		return
	}
	plan.ID = types.StringValue(primaryDomainUpdateKeyID(domainID, updateKeyID))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *primaryDomainUpdateKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state primaryDomainUpdateKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain, err := r.client.Bind().GetPrimaryDomain(ctx, state.DomainID.ValueString())
	if err != nil {
		var notFound *errs.NotFoundError
		if errors.As(err, &notFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to Read BIND Primary Domain Update Key", err.Error())
		return
	}
	if !hasUpdateKey(domain, state.UpdateKeyID.ValueString()) {
		resp.State.RemoveResource(ctx)
		return
	}
	state.ID = types.StringValue(primaryDomainUpdateKeyID(state.DomainID.ValueString(), state.UpdateKeyID.ValueString()))
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *primaryDomainUpdateKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan primaryDomainUpdateKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = types.StringValue(primaryDomainUpdateKeyID(plan.DomainID.ValueString(), plan.UpdateKeyID.ValueString()))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *primaryDomainUpdateKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state primaryDomainUpdateKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	primaryDomainUpdateKeyMu.Lock()
	defer primaryDomainUpdateKeyMu.Unlock()

	domainID := state.DomainID.ValueString()
	updateKeyID := state.UpdateKeyID.ValueString()
	domain, err := r.client.Bind().GetPrimaryDomain(ctx, domainID)
	if err != nil {
		var notFound *errs.NotFoundError
		if errors.As(err, &notFound) {
			return
		}
		resp.Diagnostics.AddError("Unable to Read BIND Primary Domain Update Key", err.Error())
		return
	}
	if !hasUpdateKey(domain, updateKeyID) {
		return
	}
	removeUpdateKey(domain, updateKeyID)
	if err := r.client.Bind().UpdatePrimaryDomain(ctx, domainID, domain); err != nil {
		resp.Diagnostics.AddError("Unable to Detach BIND Primary Domain Update Key", err.Error())
	}
}

func (r *primaryDomainUpdateKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError("Invalid Import ID", "Expected `domain_uuid/update_key_uuid` for an existing BIND primary-domain update-key attachment.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("domain_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("update_key_id"), parts[1])...)
}
