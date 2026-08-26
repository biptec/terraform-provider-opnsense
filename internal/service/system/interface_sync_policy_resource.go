package system

import (
	"context"
	"errors"

	"github.com/biptec/opnsense-go/pkg/api"
	apiextensions "github.com/biptec/opnsense-go/pkg/api_extensions"
	"github.com/biptec/opnsense-go/pkg/errs"
	"github.com/biptec/opnsense-go/pkg/opnsense"
	"github.com/biptec/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &interfaceSyncPolicyResource{}
var _ resource.ResourceWithConfigure = &interfaceSyncPolicyResource{}
var _ resource.ResourceWithImportState = &interfaceSyncPolicyResource{}

type interfaceSyncPolicyResource struct{ client opnsense.Client }

func newInterfaceSyncPolicyResource() resource.Resource { return &interfaceSyncPolicyResource{} }

func (r *interfaceSyncPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_interface_sync_policy"
}

func (r *interfaceSyncPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = interfaceSyncPolicyResourceSchema()
}

func (r *interfaceSyncPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureClient(req, resp)
}

func interfaceSyncPolicyToAPI(data *interfaceSyncPolicyResourceModel) *apiextensions.InterfaceSyncPolicy {
	sync := api.BoolString("0")
	if data.Synchronize.ValueBool() {
		sync = api.BoolString("1")
	}
	return &apiextensions.InterfaceSyncPolicy{
		ID:          data.PolicyID.ValueString(),
		Description: data.Description.ValueString(),
		Synchronize: sync,
	}
}

func interfaceSyncPolicyFromAPI(remote *apiextensions.InterfaceSyncPolicy, id string) *interfaceSyncPolicyResourceModel {
	return &interfaceSyncPolicyResourceModel{
		PolicyID:    types.StringValue(remote.ID),
		Description: types.StringValue(remote.Description),
		Synchronize: types.BoolValue(tools.StringToBool(string(remote.Synchronize))),
		ID:          types.StringValue(id),
	}
}

func (r *interfaceSyncPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data interfaceSyncPolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := r.client.ApiExtensions().AddInterfaceSyncPolicy(ctx, interfaceSyncPolicyToAPI(&data))
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create Interface Sync Policy", err.Error())
		return
	}
	remote, err := r.client.ApiExtensions().GetInterfaceSyncPolicy(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Interface Sync Policy Created but Read Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, interfaceSyncPolicyFromAPI(remote, id))...)
}

func (r *interfaceSyncPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data interfaceSyncPolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	remote, err := r.client.ApiExtensions().GetInterfaceSyncPolicy(ctx, data.ID.ValueString())
	if err != nil {
		var notFound *errs.NotFoundError
		if errors.As(err, &notFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to Read Interface Sync Policy", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, interfaceSyncPolicyFromAPI(remote, data.ID.ValueString()))...)
}

func (r *interfaceSyncPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data interfaceSyncPolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.ApiExtensions().UpdateInterfaceSyncPolicy(ctx, data.ID.ValueString(), interfaceSyncPolicyToAPI(&data)); err != nil {
		resp.Diagnostics.AddError("Unable to Update Interface Sync Policy", err.Error())
		return
	}
	remote, err := r.client.ApiExtensions().GetInterfaceSyncPolicy(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Interface Sync Policy Updated but Read Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, interfaceSyncPolicyFromAPI(remote, data.ID.ValueString()))...)
}

func (r *interfaceSyncPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data interfaceSyncPolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.ApiExtensions().DeleteInterfaceSyncPolicy(ctx, data.ID.ValueString()); err != nil {
		var notFound *errs.NotFoundError
		if !errors.As(err, &notFound) {
			resp.Diagnostics.AddError("Unable to Delete Interface Sync Policy", err.Error())
		}
	}
}

func (r *interfaceSyncPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
