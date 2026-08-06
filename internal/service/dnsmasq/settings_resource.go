package dnsmasq

import (
	"context"
	"fmt"

	"github.com/biptec/opnsense-go/pkg/api"
	"github.com/biptec/opnsense-go/pkg/opnsense"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &settingsResource{}
var _ resource.ResourceWithConfigure = &settingsResource{}
var _ resource.ResourceWithImportState = &settingsResource{}

type settingsResource struct {
	client opnsense.Client
}

func newSettingsResource() resource.Resource {
	return &settingsResource{}
}
func (r *settingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dnsmasq_settings"
}

func (r *settingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = settingsResourceSchema()
}

func (r *settingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *api.Client, got %T.", req.ProviderData))
		return
	}
	r.client = opnsense.NewClient(client)
}

func (r *settingsResource) Create(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.AddError("Cannot Create Singleton Resource", "dnsmasq settings already exist in OPNsense. Import them first with: terraform import opnsense_dnsmasq_settings.<name> dnsmasq_settings")
}
func (r *settingsResource) Read(ctx context.Context, _ resource.ReadRequest, resp *resource.ReadResponse) {
	remote, err := r.client.Dnsmasq().GeneralSettingsGet(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read dnsmasq Settings", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, settingsAPIToModel(remote))...)
}

func (r *settingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan settingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	remote, err := r.client.Dnsmasq().GeneralSettingsGet(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read dnsmasq Settings", err.Error())
		return
	}
	applySettingsModel(&remote.Dnsmasq, &plan)
	setResult, err := r.client.Dnsmasq().GeneralSettingsSet(ctx, &remote.Dnsmasq)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Update dnsmasq Settings", err.Error())
		return
	}
	if err := validateDnsmasqSetResult(setResult); err != nil {
		resp.Diagnostics.AddError("Unable to Update dnsmasq Settings", err.Error())
		return
	}
	reconfigureResult, err := r.client.Dnsmasq().ServiceReconfigure(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Reconfigure dnsmasq", err.Error())
		return
	}
	if err := validateDnsmasqReconfigureResult(reconfigureResult); err != nil {
		resp.Diagnostics.AddError("Unable to Reconfigure dnsmasq", err.Error())
		return
	}
	updated, err := r.client.Dnsmasq().GeneralSettingsGet(ctx)
	if err != nil {
		resp.Diagnostics.AddError("dnsmasq Settings Updated but Read Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, settingsAPIToModel(updated))...)
}

func validateDnsmasqSetResult(result *api.ActionResult) error {
	if result == nil {
		return fmt.Errorf("dnsmasq settings API returned an empty response")
	}
	if result.Result != "saved" {
		return fmt.Errorf("dnsmasq settings API returned result %q instead of %q", result.Result, "saved")
	}
	return nil
}
func validateDnsmasqReconfigureResult(result *api.ReconfigureStatusResult) error {
	if result == nil {
		return fmt.Errorf("dnsmasq reconfigure API returned an empty response")
	}
	if result.Status != "ok" {
		return fmt.Errorf("dnsmasq reconfigure API returned status %q instead of %q", result.Status, "ok")
	}
	return nil
}

func (r *settingsResource) Delete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning("Singleton Resource Removed From State Only", "dnsmasq settings remain unchanged in OPNsense. Re-import with ID `dnsmasq_settings` to manage them again.")
}

func (r *settingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != "dnsmasq_settings" {
		resp.Diagnostics.AddError("Invalid Import ID", fmt.Sprintf("dnsmasq settings must be imported with ID dnsmasq_settings, got %q.", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	tflog.Info(ctx, "imported dnsmasq settings")
}
