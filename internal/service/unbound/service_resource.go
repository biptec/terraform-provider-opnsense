package unbound

import (
	"context"
	"fmt"

	"github.com/biptec/opnsense-go/pkg/api"
	"github.com/biptec/opnsense-go/pkg/opnsense"
	apiunbound "github.com/biptec/opnsense-go/pkg/unbound"
	"github.com/biptec/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &serviceResource{}
var _ resource.ResourceWithConfigure = &serviceResource{}
var _ resource.ResourceWithImportState = &serviceResource{}

type serviceResource struct {
	client opnsense.Client
}

func newServiceResource() resource.Resource {
	return &serviceResource{}
}
func (r *serviceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_unbound_service"
}

func (r *serviceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = serviceResourceSchema()
}

func (r *serviceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *serviceResource) Create(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.AddError("Cannot Create Singleton Resource", "Unbound service settings already exist in OPNsense. Import them first with: terraform import opnsense_unbound_service.<name> unbound_service")
}
func (r *serviceResource) Read(ctx context.Context, _ resource.ReadRequest, resp *resource.ReadResponse) {
	remote, err := r.client.Unbound().SettingsGet(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Unbound Service", err.Error())
		return
	}
	if remote == nil {
		resp.Diagnostics.AddError("Unable to Read Unbound Service", "Unbound settings API returned an empty response")
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, serviceAPIToModel(remote.Unbound.General.Enabled))...)
}

func (r *serviceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan serviceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	remote, err := r.client.Unbound().SettingsGet(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Unbound Service", err.Error())
		return
	}
	if remote == nil {
		resp.Diagnostics.AddError("Unable to Read Unbound Service", "Unbound settings API returned an empty response")
		return
	}
	applyUnboundServiceModel(&remote.Unbound, &plan)
	updateResult, err := r.client.Unbound().SettingsUpdate(ctx, &remote.Unbound)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Update Unbound Service", err.Error())
		return
	}
	if err := validateUnboundUpdateResult(updateResult); err != nil {
		resp.Diagnostics.AddError("Unable to Update Unbound Service", err.Error())
		return
	}
	reconfigureResult, err := r.client.Unbound().SettingsReconfigure(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Reconfigure Unbound", err.Error())
		return
	}
	if err := validateUnboundActionResult(reconfigureResult); err != nil {
		resp.Diagnostics.AddError("Unable to Reconfigure Unbound", err.Error())
		return
	}
	generalResult, err := r.client.Unbound().SettingsReconfigureGeneral(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Reconfigure Unbound General Settings", err.Error())
		return
	}
	if err := validateUnboundActionResult(generalResult); err != nil {
		resp.Diagnostics.AddError("Unable to Reconfigure Unbound General Settings", err.Error())
		return
	}
	updated, err := r.client.Unbound().SettingsGet(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unbound Service Updated but Read Failed", err.Error())
		return
	}
	if updated == nil {
		resp.Diagnostics.AddError("Unbound Service Updated but Read Failed", "Unbound settings API returned an empty response")
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, serviceAPIToModel(updated.Unbound.General.Enabled))...)
}

func applyUnboundServiceModel(settings *apiunbound.Settings, model *serviceResourceModel) {
	settings.General.Enabled = tools.BoolToString(model.Enabled.ValueBool())
}

func validateUnboundUpdateResult(result *apiunbound.Result) error {
	if result == nil {
		return fmt.Errorf("unbound settings API returned an empty response")
	}
	if result.Result != "saved" {
		return fmt.Errorf("unbound settings API returned result %q instead of %q", result.Result, "saved")
	}
	return nil
}

func validateUnboundActionResult(result *apiunbound.ActionResult) error {
	if result == nil {
		return fmt.Errorf("unbound reconfigure API returned an empty response")
	}
	if result.Status != "ok" {
		return fmt.Errorf("unbound reconfigure API returned status %q instead of %q", result.Status, "ok")
	}
	return nil
}
func (r *serviceResource) Delete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning("Singleton Resource Removed From State Only", "Unbound service settings remain unchanged in OPNsense. Re-import with ID `unbound_service` to manage them again.")
}

func (r *serviceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != "unbound_service" {
		resp.Diagnostics.AddError("Invalid Import ID", fmt.Sprintf("Unbound service must be imported with ID unbound_service, got %q.", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	tflog.Info(ctx, "imported Unbound service")
}
