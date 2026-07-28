package interfaces

import (
	"context"
	"errors"
	"fmt"

	"github.com/biptec/opnsense-go/pkg/api"
	"github.com/biptec/opnsense-go/pkg/errs"
	"github.com/biptec/opnsense-go/pkg/opnsense"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &assignmentResource{}
var _ resource.ResourceWithConfigure = &assignmentResource{}
var _ resource.ResourceWithImportState = &assignmentResource{}
var _ resource.ResourceWithModifyPlan = &assignmentResource{}

type assignmentResource struct {
	client opnsense.Client
}

func newAssignmentResource() resource.Resource { return &assignmentResource{} }

func (r *assignmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_interfaces_assignment"
}

func (r *assignmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = assignmentResourceSchema()
}

func (r *assignmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	apiClient, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *api.Client, got %T.", req.ProviderData))
		return
	}
	r.client = opnsense.NewClient(apiClient)
}

func (r *assignmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data assignmentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	assignment, err := convertAssignmentSchemaToStruct(&data)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Interface Assignment", err.Error())
		return
	}
	id, err := r.client.Interfaces().AddAssignmentResolved(ctx, assignment)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create Interface Assignment", err.Error())
		return
	}

	created, err := r.client.Interfaces().GetAssignment(ctx, id)
	if err != nil {
		data.Id = typesString(id)
		data.Name = typesString(id)
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		resp.Diagnostics.AddError("Interface Assignment Created but Read Failed", err.Error())
		return
	}
	state := convertAssignmentStructToResourceSchema(created, id, data.AllowReaddress)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	tflog.Trace(ctx, "created interface assignment", map[string]any{"id": id})
}

func (r *assignmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data assignmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	assignment, err := r.client.Interfaces().GetAssignment(ctx, data.Id.ValueString())
	if err != nil {
		var notFound *errs.NotFoundError
		if errors.As(err, &notFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to Read Interface Assignment", err.Error())
		return
	}
	state := convertAssignmentStructToResourceSchema(assignment, data.Id.ValueString(), data.AllowReaddress)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *assignmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data assignmentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	assignment, err := convertAssignmentSchemaToStruct(&data)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Interface Assignment", err.Error())
		return
	}
	id := data.Id.ValueString()
	if err := r.client.Interfaces().UpdateAssignment(ctx, id, assignment); err != nil {
		resp.Diagnostics.AddError("Unable to Update Interface Assignment", err.Error())
		return
	}
	updated, err := r.client.Interfaces().GetAssignment(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Interface Assignment Updated but Read Failed", err.Error())
		return
	}
	state := convertAssignmentStructToResourceSchema(updated, id, data.AllowReaddress)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *assignmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data assignmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if data.AllowReaddress.IsNull() || data.AllowReaddress.IsUnknown() || !data.AllowReaddress.ValueBool() {
		resp.Diagnostics.AddError(
			"Interface Assignment Deletion Requires Explicit Approval",
			"Deleting an interface assignment can disconnect management access or remove dependent firewall configuration. Set allow_readdress = true and apply that change before destroying the assignment.",
		)
		return
	}
	if err := r.client.Interfaces().DeleteAssignment(ctx, data.Id.ValueString()); err != nil {
		resp.Diagnostics.AddError("Unable to Delete Interface Assignment", err.Error())
	}
}

func (r *assignmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *assignmentResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}
	var state, plan assignmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() || plan.AllowReaddress.IsUnknown() || plan.AllowReaddress.ValueBool() {
		return
	}
	if assignmentAddressingChanged(&state, &plan) {
		resp.Diagnostics.AddError(
			"Interface Readdressing Requires Explicit Approval",
			"Changing an existing assignment's device or address configuration can disconnect the OPNsense API during apply. Set allow_readdress = true only when an alternate management path or recovery method is available.",
		)
	}
}

func assignmentAddressingChanged(state, plan *assignmentResourceModel) bool {
	if !state.Device.Equal(plan.Device) {
		return true
	}
	if state.IPv4 == nil || plan.IPv4 == nil || state.IPv6 == nil || plan.IPv6 == nil {
		return true
	}
	return !state.IPv4.Mode.Equal(plan.IPv4.Mode) || !state.IPv4.Address.Equal(plan.IPv4.Address) ||
		!state.IPv4.Prefix.Equal(plan.IPv4.Prefix) || !state.IPv4.Gateway.Equal(plan.IPv4.Gateway) ||
		!state.IPv6.Mode.Equal(plan.IPv6.Mode) || !state.IPv6.Address.Equal(plan.IPv6.Address) ||
		!state.IPv6.Prefix.Equal(plan.IPv6.Prefix) || !state.IPv6.Gateway.Equal(plan.IPv6.Gateway) ||
		!state.IPv6.TrackInterface.Equal(plan.IPv6.TrackInterface) || !state.IPv6.TrackPrefixID.Equal(plan.IPv6.TrackPrefixID)
}
