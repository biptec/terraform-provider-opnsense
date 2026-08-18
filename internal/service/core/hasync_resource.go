package core

import (
	"context"
	"fmt"

	"github.com/biptec/opnsense-go/pkg/api"
	"github.com/biptec/opnsense-go/pkg/opnsense"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &hasyncResource{}
var _ resource.ResourceWithConfigure = &hasyncResource{}
var _ resource.ResourceWithImportState = &hasyncResource{}

type hasyncResource struct{ client opnsense.Client }

func newHasyncResource() resource.Resource { return &hasyncResource{} }
func (r *hasyncResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_core_hasync"
}
func (r *hasyncResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = hasyncResourceSchema()
}
func (r *hasyncResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *api.Client, got %T.", req.ProviderData))
		return
	}
	r.client = opnsense.NewClient(c)
}
func (r *hasyncResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan hasyncModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	var password types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("password"), &password)...)
	if resp.Diagnostics.HasError() {
		return
	}
	state, err := r.applySettings(ctx, &plan, password)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Adopt OPNsense HA Settings", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	tflog.Info(ctx, "adopted existing OPNsense HA settings", map[string]any{"id": hasyncID})
}
func (r *hasyncResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var old hasyncModel
	resp.Diagnostics.Append(req.State.Get(ctx, &old)...)
	if resp.Diagnostics.HasError() {
		return
	}
	remote, err := r.client.Core().HasyncGet(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read OPNsense HA Settings", err.Error())
		return
	}
	state := hasyncAPIToModel(&remote.Hasync)
	preserveHasyncConfiguredEmptyStrings(state, &old)
	state.PasswordVersion = old.PasswordVersion
	if state.PasswordVersion.IsNull() || state.PasswordVersion.IsUnknown() {
		state.PasswordVersion = types.Int64Value(0)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
func (r *hasyncResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, old hasyncModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &old)...)
	var password types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("password"), &password)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !plan.PasswordVersion.IsUnknown() && !old.PasswordVersion.IsUnknown() && !plan.PasswordVersion.IsNull() && !old.PasswordVersion.IsNull() && plan.PasswordVersion.ValueInt64() != old.PasswordVersion.ValueInt64() {
		if password.IsNull() || password.IsUnknown() || password.ValueString() == "" {
			resp.Diagnostics.AddError("Missing Rotated XMLRPC Password", "password_version changed but no write-only password was supplied. Provide the new password together with the incremented password_version.")
			return
		}
	}
	state, err := r.applySettings(ctx, &plan, password)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Update OPNsense HA Settings", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *hasyncResource) applySettings(ctx context.Context, plan *hasyncModel, password types.String) (*hasyncModel, error) {
	remote, err := r.client.Core().HasyncGet(ctx)
	if err != nil {
		return nil, fmt.Errorf("read existing OPNsense HA settings: %w", err)
	}
	current := hasyncAPIToModel(&remote.Hasync)
	completeHasyncModel(plan, current)
	applyHasyncModel(&remote.Hasync, plan, password)
	result, err := r.client.Core().HasyncSet(ctx, &remote.Hasync)
	if err != nil {
		return nil, fmt.Errorf("save OPNsense HA settings: %w", err)
	}
	if result == nil || result.Result != "saved" {
		return nil, fmt.Errorf("unexpected API result: %#v", result)
	}
	if _, err := r.client.Core().HasyncReconfigure(ctx); err != nil {
		return nil, fmt.Errorf("reconfigure pfsync: %w", err)
	}
	updated, err := r.client.Core().HasyncGet(ctx)
	if err != nil {
		return nil, fmt.Errorf("read updated OPNsense HA settings: %w", err)
	}
	state := hasyncAPIToModel(&updated.Hasync)
	preserveHasyncConfiguredEmptyStrings(state, plan)
	state.PasswordVersion = plan.PasswordVersion
	if state.PasswordVersion.IsNull() || state.PasswordVersion.IsUnknown() {
		state.PasswordVersion = types.Int64Value(0)
	}
	return state, nil
}

func preserveHasyncConfiguredEmptyStrings(state, configured *hasyncModel) {
	if !configured.PfsyncPeerIP.IsNull() && !configured.PfsyncPeerIP.IsUnknown() && configured.PfsyncPeerIP.ValueString() == "" {
		state.PfsyncPeerIP = types.StringValue("")
	}
	if !configured.SynchronizeToIP.IsNull() && !configured.SynchronizeToIP.IsUnknown() && configured.SynchronizeToIP.ValueString() == "" {
		state.SynchronizeToIP = types.StringValue("")
	}
	if !configured.Username.IsNull() && !configured.Username.IsUnknown() && configured.Username.ValueString() == "" {
		state.Username = types.StringValue("")
	}
}

func completeHasyncModel(plan, current *hasyncModel) {
	if plan.DisablePreempt.IsNull() || plan.DisablePreempt.IsUnknown() {
		plan.DisablePreempt = current.DisablePreempt
	}
	if plan.DisconnectPPPs.IsNull() || plan.DisconnectPPPs.IsUnknown() {
		plan.DisconnectPPPs = current.DisconnectPPPs
	}
	if plan.PfsyncInterface.IsNull() || plan.PfsyncInterface.IsUnknown() {
		plan.PfsyncInterface = current.PfsyncInterface
	}
	if plan.PfsyncPeerIP.IsNull() || plan.PfsyncPeerIP.IsUnknown() {
		plan.PfsyncPeerIP = current.PfsyncPeerIP
	}
	if plan.PfsyncVersion.IsNull() || plan.PfsyncVersion.IsUnknown() {
		plan.PfsyncVersion = current.PfsyncVersion
	}
	if plan.PfsyncDefer.IsNull() || plan.PfsyncDefer.IsUnknown() {
		plan.PfsyncDefer = current.PfsyncDefer
	}
	if plan.SynchronizeToIP.IsNull() || plan.SynchronizeToIP.IsUnknown() {
		plan.SynchronizeToIP = current.SynchronizeToIP
	}
	if plan.VerifyPeer.IsNull() || plan.VerifyPeer.IsUnknown() {
		plan.VerifyPeer = current.VerifyPeer
	}
	if plan.Username.IsNull() || plan.Username.IsUnknown() {
		plan.Username = current.Username
	}
	if plan.SyncItems.IsNull() || plan.SyncItems.IsUnknown() {
		plan.SyncItems = current.SyncItems
	}
	if plan.PasswordVersion.IsNull() || plan.PasswordVersion.IsUnknown() {
		plan.PasswordVersion = types.Int64Value(0)
	}
}

func (r *hasyncResource) Delete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning("Singleton Resource Removed From State Only", "OPNsense HA settings remain unchanged. Re-import with ID `core_hasync` to manage them again.")
}
func (r *hasyncResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != hasyncID {
		resp.Diagnostics.AddError("Invalid Import ID", fmt.Sprintf("OPNsense HA settings must be imported with ID %s, got %q.", hasyncID, req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("password_version"), int64(0))...)
	tflog.Info(ctx, "imported OPNsense HA settings")
}
