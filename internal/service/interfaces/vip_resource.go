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
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &vipResource{}
var _ resource.ResourceWithConfigure = &vipResource{}
var _ resource.ResourceWithImportState = &vipResource{}

func newVipResource() resource.Resource {
	return &vipResource{}
}

// vipResource defines the resource implementation.
type vipResource struct {
	client opnsense.Client
}

func (r *vipResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_interfaces_vip"
}

func (r *vipResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = vipResourceSchema()
}

func (r *vipResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	apiClient, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *opnsense.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = opnsense.NewClient(apiClient)
}

func (r *vipResource) readState(ctx context.Context, id string, prior *vipResourceModel) (*vipResourceModel, error) {
	vip, err := r.client.Interfaces().GetVip(ctx, id)
	if err != nil {
		return nil, err
	}
	state, err := convertVipStructToSchema(vip)
	if err != nil {
		return nil, err
	}
	state.Id = types.StringValue(id)
	if prior != nil && state.Password.IsNull() && !prior.Password.IsNull() && !prior.Password.IsUnknown() {
		state.Password = prior.Password
	}
	return state, nil
}

func vipFallbackState(data *vipResourceModel, id string) *vipResourceModel {
	data.Id = types.StringValue(id)
	if data.Address.IsUnknown() {
		data.Address = types.StringNull()
	}
	if data.VHIDText.IsUnknown() {
		data.VHIDText = types.StringNull()
	}
	return data
}

func (r *vipResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *vipResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vip, err := convertVipSchemaToStruct(data)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse vip, got error: %s", err))
		return
	}

	id, addErr := r.client.Interfaces().AddVip(ctx, vip)
	if id != "" {
		state, readErr := r.readState(ctx, id, data)
		if readErr != nil {
			state = vipFallbackState(data, id)
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
		if addErr == nil && readErr != nil {
			resp.Diagnostics.AddError("VIP Created but Read Failed", readErr.Error())
			return
		}
	}
	if addErr != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create vip, got error: %s", addErr))
		return
	}
	if id == "" {
		resp.Diagnostics.AddError("Client Error", "Unable to create vip: API returned an empty identifier")
		return
	}
	tflog.Trace(ctx, "created a resource", map[string]any{"id": id})
}

func (r *vipResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *vipResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state, err := r.readState(ctx, data.Id.ValueString(), data)
	if err != nil {
		var notFoundError *errs.NotFoundError
		if errors.As(err, &notFoundError) {
			tflog.Warn(ctx, "vip not present in remote, removing from state")
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read vip, got error: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *vipResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *vipResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vip, err := convertVipSchemaToStruct(data)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse vip, got error: %s", err))
		return
	}
	id := data.Id.ValueString()
	if err := r.client.Interfaces().UpdateVip(ctx, id, vip); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update vip, got error: %s", err))
		return
	}
	state, err := r.readState(ctx, id, data)
	if err != nil {
		resp.Diagnostics.Append(resp.State.Set(ctx, vipFallbackState(data, id))...)
		resp.Diagnostics.AddError("VIP Updated but Read Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *vipResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *vipResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Interfaces().DeleteVip(ctx, data.Id.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to delete vip, got error: %s", err))
		return
	}
}

func (r *vipResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
