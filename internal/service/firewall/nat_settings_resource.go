package firewall

import (
	"context"
	"fmt"
	"strings"

	"github.com/biptec/opnsense-go/pkg/api"
	opnfirewall "github.com/biptec/opnsense-go/pkg/firewall"
	"github.com/biptec/opnsense-go/pkg/opnsense"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &natSettingsResource{}
var _ resource.ResourceWithConfigure = &natSettingsResource{}
var _ resource.ResourceWithImportState = &natSettingsResource{}

type natSettingsResource struct {
	client opnsense.Client
}

func newNATSettingsResource() resource.Resource { return &natSettingsResource{} }

func (r *natSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_nat_settings"
}

func (r *natSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = natSettingsResourceSchema()
}

func (r *natSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *natSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan natSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.apply(ctx, plan.Mode.ValueString()); err != nil {
		resp.Diagnostics.AddError("Unable to Configure Outbound NAT Mode", err.Error())
		return
	}
	state, err := r.read(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Outbound NAT Mode Configured but Read Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *natSettingsResource) Read(ctx context.Context, _ resource.ReadRequest, resp *resource.ReadResponse) {
	state, err := r.read(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Outbound NAT Mode", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *natSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan natSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.apply(ctx, plan.Mode.ValueString()); err != nil {
		resp.Diagnostics.AddError("Unable to Configure Outbound NAT Mode", err.Error())
		return
	}
	state, err := r.read(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Outbound NAT Mode Configured but Read Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *natSettingsResource) Delete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning(
		"Singleton Resource Removed From State Only",
		"Outbound NAT mode remains unchanged in OPNsense. Re-import with ID `firewall_nat_settings` to manage it again.",
	)
}

func (r *natSettingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != natSettingsID {
		resp.Diagnostics.AddError("Invalid Import ID", fmt.Sprintf("Outbound NAT settings must be imported with ID %q, got %q.", natSettingsID, req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

func (r *natSettingsResource) apply(ctx context.Context, mode string) error {
	settings := &opnfirewall.SourceNATSettings{
		General: opnfirewall.SourceNATGeneralSettings{Mode: api.SelectedMap(mode)},
	}
	result, err := r.client.Firewall().SourceNATSettingsSet(ctx, settings)
	if err != nil {
		return fmt.Errorf("set Source NAT mode: %w", err)
	}
	if err = validateSourceNATSettingsSetResult(result); err != nil {
		return err
	}
	applied, err := r.client.Firewall().SourceNATSettingsApply(ctx)
	if err != nil {
		return fmt.Errorf("apply Source NAT mode: %w", err)
	}
	if err = validateSourceNATSettingsApplyResult(applied); err != nil {
		return err
	}
	tflog.Info(ctx, "configured outbound Source NAT mode", map[string]any{"mode": mode})
	return nil
}

func validateSourceNATSettingsSetResult(result *api.ActionResult) error {
	if result == nil {
		return fmt.Errorf("source NAT settings API returned an empty response")
	}
	if result.Result != "saved" {
		return fmt.Errorf("source NAT settings API returned result %q instead of %q", result.Result, "saved")
	}
	return nil
}

func validateSourceNATSettingsApplyResult(result *opnfirewall.SourceNATApplyResult) error {
	if result == nil {
		return fmt.Errorf("source NAT apply API returned an empty response")
	}
	if !strings.EqualFold(strings.TrimSpace(result.Status), "OK") {
		return fmt.Errorf("source NAT apply API returned status %q instead of OK", result.Status)
	}
	return nil
}

func (r *natSettingsResource) read(ctx context.Context) (*natSettingsResourceModel, error) {
	settings, err := r.client.Firewall().SourceNATSettingsGet(ctx)
	if err != nil {
		return nil, err
	}
	mode := settings.Filter.General.Mode.String()
	if mode == "" {
		return nil, fmt.Errorf("source NAT settings API returned an empty mode")
	}
	return &natSettingsResourceModel{
		ID:   types.StringValue(natSettingsID),
		Mode: types.StringValue(mode),
	}, nil
}
