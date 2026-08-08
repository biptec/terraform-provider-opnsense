package bind

import (
	"context"
	"errors"
	"slices"
	"strings"

	"github.com/biptec/opnsense-go/pkg/api"
	apibind "github.com/biptec/opnsense-go/pkg/bind"
	"github.com/biptec/opnsense-go/pkg/errs"
	"github.com/biptec/terraform-provider-opnsense/internal/tools"
	"github.com/biptec/terraform-provider-opnsense/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &primaryDomainTransferResource{}
var _ resource.ResourceWithConfigure = &primaryDomainTransferResource{}
var _ resource.ResourceWithImportState = &primaryDomainTransferResource{}

type primaryDomainTransferResource struct{ resourceClient }

type primaryDomainTransferResourceModel struct {
	ID            types.String `tfsdk:"id"`
	DomainID      types.String `tfsdk:"domain_id"`
	TransferKeyID types.String `tfsdk:"transfer_key_id"`
	AlsoNotify    types.Set    `tfsdk:"also_notify"`
}

func newPrimaryDomainTransferResource() resource.Resource { return &primaryDomainTransferResource{} }

func (r *primaryDomainTransferResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bind_primary_domain_transfer"
}

func (r *primaryDomainTransferResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Attaches secondary-transfer settings to an existing BIND primary zone without owning the zone itself. Delete removes only the transfer attachment.",
		Attributes: map[string]schema.Attribute{
			"id":              schema.StringAttribute{Computed: true, MarkdownDescription: "Attachment identity; equal to domain_id."},
			"domain_id":       schema.StringAttribute{Required: true, MarkdownDescription: "UUID of the existing primary zone.", Validators: []validator.String{validators.IsUUIDv4()}, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"transfer_key_id": schema.StringAttribute{Required: true, MarkdownDescription: "UUID of the TSIG key used for authenticated AXFR/IXFR and NOTIFY.", Validators: []validator.String{validators.IsUUIDv4()}},
			"also_notify": schema.SetAttribute{
				Required:    true,
				ElementType: types.StringType,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
					setvalidator.ValueStringsAre(validators.IPAddress()),
				},
				MarkdownDescription: "Secondary nameserver addresses that receive NOTIFY.",
			},
		},
	}
}

func transferValues(model *primaryDomainTransferResourceModel) (string, []string) {
	notify := tools.SetToStringSlice(model.AlsoNotify)
	slices.Sort(notify)
	return model.TransferKeyID.ValueString(), notify
}

func remoteTransferValues(domain *apibind.PrimaryDomain) (string, []string) {
	notify := make([]string, 0, len(domain.AlsoNotify))
	for _, value := range domain.AlsoNotify {
		value = strings.TrimSpace(value)
		if value != "" {
			notify = append(notify, value)
		}
	}
	slices.Sort(notify)
	return strings.TrimSpace(domain.TransferKey.String()), notify
}

func transferMatches(domain *apibind.PrimaryDomain, model *primaryDomainTransferResourceModel) bool {
	remoteKey, remoteNotify := remoteTransferValues(domain)
	key, notify := transferValues(model)
	return remoteKey == key && slices.Equal(remoteNotify, notify)
}

func transferEmpty(domain *apibind.PrimaryDomain) bool {
	key, notify := remoteTransferValues(domain)
	return key == "" && len(notify) == 0
}

func applyTransfer(domain *apibind.PrimaryDomain, model *primaryDomainTransferResourceModel) {
	key, notify := transferValues(model)
	domain.TransferKey = api.SelectedMap(key)
	domain.AlsoNotify = api.SelectedMapList(notify)
}

func clearTransfer(domain *apibind.PrimaryDomain) {
	domain.TransferKey = api.SelectedMap("")
	domain.AlsoNotify = api.SelectedMapList{}
}

func (r *primaryDomainTransferResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan primaryDomainTransferResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domainID := plan.DomainID.ValueString()
	domain, err := r.client.Bind().GetPrimaryDomain(ctx, domainID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read BIND Primary Domain", err.Error())
		return
	}
	if !transferEmpty(domain) {
		resp.Diagnostics.AddError("BIND Primary Domain Transfer Already Owned", "The primary zone already has transfer settings. Refusing to adopt or overwrite an attachment that may belong to another Terraform state; import it explicitly if this state is the intended owner.")
		return
	}
	applyTransfer(domain, &plan)
	if err := r.client.Bind().UpdatePrimaryDomain(ctx, domainID, domain); err != nil {
		resp.Diagnostics.AddError("Unable to Attach BIND Primary Domain Transfer", err.Error())
		return
	}
	plan.ID = types.StringValue(domainID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *primaryDomainTransferResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state primaryDomainTransferResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domainID := state.DomainID.ValueString()
	domain, err := r.client.Bind().GetPrimaryDomain(ctx, domainID)
	if err != nil {
		var notFound *errs.NotFoundError
		if errors.As(err, &notFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to Read BIND Primary Domain Transfer", err.Error())
		return
	}
	state.ID = types.StringValue(domainID)
	if state.TransferKeyID.IsNull() || state.TransferKeyID.IsUnknown() || state.AlsoNotify.IsNull() || state.AlsoNotify.IsUnknown() {
		// Import intentionally adopts the attachment that already exists on this domain.
		key, notify := remoteTransferValues(domain)
		if key == "" || len(notify) == 0 {
			resp.Diagnostics.AddError("BIND Primary Domain Transfer Not Found", "The imported primary zone does not have a complete transfer attachment.")
			return
		}
		state.TransferKeyID = types.StringValue(key)
		state.AlsoNotify = tools.StringSliceToSet(notify)
	} else if !transferMatches(domain, &state) {
		resp.Diagnostics.AddError("BIND Primary Domain Transfer Changed Externally", "Current transfer settings no longer match this Terraform state. Refusing to adopt or overwrite another owner's configuration.")
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *primaryDomainTransferResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state primaryDomainTransferResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domainID := state.DomainID.ValueString()
	if plan.DomainID.ValueString() != domainID {
		resp.Diagnostics.AddError("Cannot Move BIND Transfer Attachment", "domain_id cannot be changed in place; replace the attachment instead.")
		return
	}
	domain, err := r.client.Bind().GetPrimaryDomain(ctx, domainID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read BIND Primary Domain Transfer", err.Error())
		return
	}
	if !transferMatches(domain, &state) {
		resp.Diagnostics.AddError("BIND Primary Domain Transfer Changed Externally", "Current transfer settings no longer match this Terraform state. Refusing to overwrite them.")
		return
	}
	applyTransfer(domain, &plan)
	if err := r.client.Bind().UpdatePrimaryDomain(ctx, domainID, domain); err != nil {
		resp.Diagnostics.AddError("Unable to Update BIND Primary Domain Transfer", err.Error())
		return
	}
	plan.ID = types.StringValue(domainID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *primaryDomainTransferResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state primaryDomainTransferResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domainID := state.DomainID.ValueString()
	domain, err := r.client.Bind().GetPrimaryDomain(ctx, domainID)
	if err != nil {
		var notFound *errs.NotFoundError
		if errors.As(err, &notFound) {
			return
		}
		resp.Diagnostics.AddError("Unable to Read BIND Primary Domain Transfer", err.Error())
		return
	}
	if transferEmpty(domain) {
		return
	}
	if !transferMatches(domain, &state) {
		resp.Diagnostics.AddError("BIND Primary Domain Transfer Changed Externally", "Current transfer settings no longer match this Terraform state. Refusing to clear another owner's configuration.")
		return
	}
	clearTransfer(domain)
	if err := r.client.Bind().UpdatePrimaryDomain(ctx, domainID, domain); err != nil {
		resp.Diagnostics.AddError("Unable to Detach BIND Primary Domain Transfer", err.Error())
		return
	}
}

func (r *primaryDomainTransferResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID == "" {
		resp.Diagnostics.AddError("Invalid Import ID", "Expected an existing BIND primary-domain UUID.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("domain_id"), req.ID)...)
}
