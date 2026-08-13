package system

import (
	"context"
	"fmt"
	"strings"
	"time"

	apicore "github.com/biptec/opnsense-go/pkg/core"
	"github.com/biptec/opnsense-go/pkg/opnsense"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const firmwareOperationTimeout = 15 * time.Minute

var _ resource.Resource = &pluginResource{}
var _ resource.ResourceWithConfigure = &pluginResource{}
var _ resource.ResourceWithImportState = &pluginResource{}

type pluginResource struct{ client opnsense.Client }

func newPluginResource() resource.Resource { return &pluginResource{} }

func (r *pluginResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_plugin"
}

func (r *pluginResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = pluginResourceSchema()
}

func (r *pluginResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureClient(req, resp)
}

func (r *pluginResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data pluginResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()
	current, err := r.findPlugin(ctx, name)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read OPNsense Plugins", err.Error())
		return
	}
	if current == nil || !firmwareFlag(current.Installed) {
		if current == nil || !firmwareFlag(current.Provided) {
			resp.Diagnostics.AddError(
				"Plugin Is Not Available",
				fmt.Sprintf("Plugin %q is not installed and is not available from any configured OPNsense package repository.", name),
			)
			return
		}
		result, installErr := r.client.Core().FirmwareInstall(ctx, name)
		if installErr != nil {
			resp.Diagnostics.AddError("Unable to Install OPNsense Plugin", installErr.Error())
			return
		}
		if validationErr := validateFirmwareActionResult("install", result); validationErr != nil {
			resp.Diagnostics.AddError("Unable to Install OPNsense Plugin", validationErr.Error())
			return
		}
		if _, err = r.waitForPlugin(ctx, name, true, nil); err != nil {
			resp.Diagnostics.AddError("Timed Out Installing OPNsense Plugin", err.Error())
			return
		}
	}

	if err = r.setLock(ctx, name, data.Locked.ValueBool()); err != nil {
		resp.Diagnostics.AddError("Unable to Configure Plugin Lock", err.Error())
		return
	}
	state, err := r.pluginState(ctx, name, data.UninstallOnDestroy.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError("Plugin Installed but Read Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *pluginResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var prior pluginResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
	if resp.Diagnostics.HasError() {
		return
	}
	state, err := r.pluginState(ctx, prior.Name.ValueString(), prior.UninstallOnDestroy.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read OPNsense Plugin", err.Error())
		return
	}
	if !state.Installed.ValueBool() {
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *pluginResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data pluginResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	name := data.Name.ValueString()
	if err := r.setLock(ctx, name, data.Locked.ValueBool()); err != nil {
		resp.Diagnostics.AddError("Unable to Configure Plugin Lock", err.Error())
		return
	}
	state, err := r.pluginState(ctx, name, data.UninstallOnDestroy.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError("Plugin Updated but Read Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *pluginResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data pluginResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !data.UninstallOnDestroy.ValueBool() {
		resp.Diagnostics.AddWarning(
			"Plugin Removed From State Only",
			fmt.Sprintf("Plugin %q remains installed because uninstall_on_destroy is false.", data.Name.ValueString()),
		)
		return
	}

	name := data.Name.ValueString()
	current, err := r.findPlugin(ctx, name)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read OPNsense Plugins", err.Error())
		return
	}
	if current == nil || !firmwareFlag(current.Installed) {
		return
	}
	if firmwareFlag(current.Locked) {
		if err = r.setLock(ctx, name, false); err != nil {
			resp.Diagnostics.AddError("Unable to Unlock OPNsense Plugin", err.Error())
			return
		}
	}
	result, err := r.client.Core().FirmwareRemove(ctx, name)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Uninstall OPNsense Plugin", err.Error())
		return
	}
	if validationErr := validateFirmwareActionResult("remove", result); validationErr != nil {
		resp.Diagnostics.AddError("Unable to Uninstall OPNsense Plugin", validationErr.Error())
		return
	}
	if _, err = r.waitForPlugin(ctx, name, false, nil); err != nil {
		resp.Diagnostics.AddError("Timed Out Uninstalling OPNsense Plugin", err.Error())
	}
}

func (r *pluginResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("uninstall_on_destroy"), false)...)
}

func (r *pluginResource) findPlugin(ctx context.Context, name string) (*apicore.FirmwarePackage, error) {
	info, err := r.client.Core().FirmwareInfo(ctx)
	if err != nil {
		return nil, err
	}
	for index := range info.Plugins {
		if info.Plugins[index].Name == name {
			return &info.Plugins[index], nil
		}
	}
	return nil, nil
}

func (r *pluginResource) waitForPlugin(ctx context.Context, name string, installed bool, locked *bool) (*apicore.FirmwarePackage, error) {
	operationCtx, cancel := context.WithTimeout(ctx, firmwareOperationTimeout)
	defer cancel()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		plugin, err := r.findPlugin(operationCtx, name)
		if err == nil {
			actualInstalled := plugin != nil && firmwareFlag(plugin.Installed)
			lockMatches := locked == nil || (plugin != nil && firmwareFlag(plugin.Locked) == *locked)
			if actualInstalled == installed && lockMatches {
				running, runningErr := r.client.Core().FirmwareRunning(operationCtx)
				status, statusErr := r.client.Core().FirmwareUpgradeStatus(operationCtx)
				if firmwareOperationComplete(running, runningErr, status, statusErr) {
					return plugin, nil
				}
			}
		}

		select {
		case <-operationCtx.Done():
			status, statusErr := r.client.Core().FirmwareUpgradeStatus(ctx)
			return nil, fmt.Errorf(
				"waiting for plugin %q state timed out: %w; firmware status=%s",
				name,
				operationCtx.Err(),
				firmwareStatusDescription(status, statusErr),
			)
		case <-ticker.C:
		}
	}
}

func (r *pluginResource) setLock(ctx context.Context, name string, desired bool) error {
	current, err := r.findPlugin(ctx, name)
	if err != nil {
		return err
	}
	if current == nil || !firmwareFlag(current.Installed) {
		return fmt.Errorf("plugin %q is not installed", name)
	}
	if firmwareFlag(current.Locked) == desired {
		return nil
	}

	var result *apicore.FirmwareActionResult
	if desired {
		result, err = r.client.Core().FirmwareLock(ctx, name)
	} else {
		result, err = r.client.Core().FirmwareUnlock(ctx, name)
	}
	if err != nil {
		return err
	}
	action := "unlock"
	if desired {
		action = "lock"
	}
	if err = validateFirmwareActionResult(action, result); err != nil {
		return err
	}
	_, err = r.waitForPlugin(ctx, name, true, &desired)
	return err
}

func (r *pluginResource) pluginState(ctx context.Context, name string, uninstallOnDestroy bool) (*pluginResourceModel, error) {
	plugin, err := r.findPlugin(ctx, name)
	if err != nil {
		return nil, err
	}
	state := &pluginResourceModel{
		Name:               types.StringValue(name),
		UninstallOnDestroy: types.BoolValue(uninstallOnDestroy),
		ID:                 types.StringValue(name),
		Installed:          types.BoolValue(false),
		Locked:             types.BoolValue(false),
		Provided:           types.BoolValue(false),
		Version:            types.StringValue(""),
		Repository:         types.StringValue(""),
	}
	if plugin == nil {
		return state, nil
	}
	state.Version = types.StringValue(normalizeFirmwareValue(plugin.Version))
	state.Repository = types.StringValue(normalizeFirmwareValue(plugin.Repository))
	state.Installed = types.BoolValue(firmwareFlag(plugin.Installed))
	state.Provided = types.BoolValue(firmwareFlag(plugin.Provided))
	state.Locked = types.BoolValue(firmwareFlag(plugin.Locked))
	tflog.Trace(ctx, "read OPNsense plugin", map[string]any{"name": name, "installed": state.Installed.ValueBool()})
	return state, nil
}

func validateFirmwareActionResult(action string, result *apicore.FirmwareActionResult) error {
	if result == nil {
		return fmt.Errorf("firmware %s API returned an empty response", action)
	}
	if !strings.EqualFold(strings.TrimSpace(result.Status), "ok") {
		return fmt.Errorf("firmware %s API returned status %q instead of %q", action, result.Status, "ok")
	}
	return nil
}

func firmwareFlag(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "enabled", "locked":
		return true
	default:
		return false
	}
}

func normalizeFirmwareValue(value string) string {
	if strings.EqualFold(strings.TrimSpace(value), "N/A") {
		return ""
	}
	return value
}

func firmwareStatusDescription(status *apicore.FirmwareUpgradeStatusResponse, err error) string {
	if err != nil {
		return fmt.Sprintf("unavailable (%v)", err)
	}
	if status == nil || strings.TrimSpace(status.Status) == "" {
		return "unknown"
	}
	return fmt.Sprintf("%q", status.Status)
}
