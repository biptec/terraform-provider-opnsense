package system

import (
	"context"
	"fmt"
	"strings"
	"time"

	apiextensions "github.com/biptec/opnsense-go/pkg/api_extensions"
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
	if err := r.ensurePluginInstalled(ctx, name); err != nil {
		resp.Diagnostics.AddError("Unable to Ensure OPNsense Plugin Is Installed", err.Error())
		return
	}
	if err := r.setLock(ctx, name, data.Locked.ValueBool()); err != nil {
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
	current, err := r.findLocalPlugin(ctx, name)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Local OPNsense Plugin State", err.Error())
		return
	}
	if !current.Installed {
		return
	}
	if current.Locked {
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
	if name == "os-api-extensions" {
		err = r.waitForFoundationPluginRemoval(ctx, name)
	} else {
		_, err = r.waitForPlugin(ctx, name, false, nil)
	}
	if err != nil {
		resp.Diagnostics.AddError("Timed Out Uninstalling OPNsense Plugin", err.Error())
	}
}

func (r *pluginResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("uninstall_on_destroy"), false)...)
}

func findLocalPlugin(ctx context.Context, client opnsense.Client, name string) (*apiextensions.PackageState, error) {
	response, err := client.ApiExtensions().PackageGet(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("local package status API call failed: %w", err)
	}
	if response == nil {
		return nil, fmt.Errorf("local package status API returned an empty response")
	}
	if !strings.EqualFold(strings.TrimSpace(response.Status), "ok") {
		message := strings.TrimSpace(response.Message)
		if message == "" {
			message = "no error message returned"
		}
		return nil, fmt.Errorf("local package status API returned status %q: %s", response.Status, message)
	}
	if response.Package.Name != name {
		return nil, fmt.Errorf("local package status API returned package %q while %q was requested", response.Package.Name, name)
	}
	return &response.Package, nil
}

func (r *pluginResource) findLocalPlugin(ctx context.Context, name string) (*apiextensions.PackageState, error) {
	return findLocalPlugin(ctx, r.client, name)
}

// findFirmwarePlugin is intentionally reserved for mutation-only discovery or
// verification. FirmwareInfo refreshes remote repository metadata on OPNsense
// and must never be used by Read or ordinary Terraform refresh paths.
func (r *pluginResource) findFirmwarePlugin(ctx context.Context, name string) (*apicore.FirmwarePackage, error) {
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

func (r *pluginResource) ensurePluginInstalled(ctx context.Context, name string) error {
	local, localErr := r.findLocalPlugin(ctx, name)
	if localErr == nil {
		if local.Installed {
			return nil
		}
		if local.Provided {
			return r.installPlugin(ctx, name)
		}
	} else if name != "os-api-extensions" {
		return fmt.Errorf(
			"os-api-extensions 0.12 or newer is required for local package state before managing plugin %q: %w",
			name,
			localErr,
		)
	}

	// A Create is a mutation path, so an explicit repository refresh is allowed
	// here when cached local metadata cannot prove the package is available.
	// Normal Read/plan never reaches this remote repository path.
	firmwarePlugin, err := r.findFirmwarePlugin(ctx, name)
	if err != nil {
		return fmt.Errorf("unable to discover plugin %q through the firmware API: %w", name, err)
	}
	if firmwarePlugin != nil && firmwareFlag(firmwarePlugin.Installed) {
		if localErr != nil {
			return fmt.Errorf(
				"plugin %q is installed but its local package status API is unavailable; upgrade os-api-extensions to 0.12 or newer: %w",
				name,
				localErr,
			)
		}
		return fmt.Errorf("local package status for %q reported not installed while firmware API reported installed", name)
	}
	if firmwarePlugin == nil || !firmwareFlag(firmwarePlugin.Provided) {
		return fmt.Errorf("plugin %q is not installed and is not available from any configured OPNsense package repository", name)
	}
	return r.installPlugin(ctx, name)
}

func (r *pluginResource) installPlugin(ctx context.Context, name string) error {
	result, err := r.client.Core().FirmwareInstall(ctx, name)
	if err != nil {
		return err
	}
	if err = validateFirmwareActionResult("install", result); err != nil {
		return err
	}
	_, err = r.waitForPlugin(ctx, name, true, nil)
	return err
}

func (r *pluginResource) waitForPlugin(ctx context.Context, name string, installed bool, locked *bool) (*apiextensions.PackageState, error) {
	operationCtx, cancel := context.WithTimeout(ctx, firmwareOperationTimeout)
	defer cancel()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	doneMismatches := 0
	var lastLocalErr error

	for {
		plugin, localErr := r.findLocalPlugin(operationCtx, name)
		lastLocalErr = localErr
		status, statusErr := r.client.Core().FirmwareUpgradeStatus(operationCtx)
		if localErr == nil {
			actualInstalled := plugin.Installed
			lockMatches := locked == nil || plugin.Locked == *locked
			if actualInstalled == installed && lockMatches {
				running, runningErr := r.client.Core().FirmwareRunning(operationCtx)
				if firmwareOperationComplete(running, runningErr, status, statusErr) {
					return plugin, nil
				}
			}
		}

		operationDone := locked == nil && statusErr == nil && status != nil && strings.EqualFold(strings.TrimSpace(status.Status), "done")
		if operationDone {
			doneMismatches++
			if doneMismatches >= 2 {
				logTail := strings.TrimSpace(status.Log)
				if len(logTail) > 2048 {
					logTail = logTail[len(logTail)-2048:]
				}
				if lastLocalErr != nil {
					return nil, fmt.Errorf("firmware operation completed but local package state for %q remained unavailable: %v; firmware log tail: %s", name, lastLocalErr, logTail)
				}
				return nil, fmt.Errorf("firmware operation completed but plugin %q did not reach installed=%t with requested lock state; firmware log tail: %s", name, installed, logTail)
			}
		} else {
			doneMismatches = 0
		}

		select {
		case <-operationCtx.Done():
			status, statusErr := r.client.Core().FirmwareUpgradeStatus(ctx)
			return nil, fmt.Errorf(
				"waiting for plugin %q local state timed out: %w; local state error=%v; firmware status=%s",
				name,
				operationCtx.Err(),
				lastLocalErr,
				firmwareStatusDescription(status, statusErr),
			)
		case <-ticker.C:
		}
	}
}

func (r *pluginResource) waitForFoundationPluginRemoval(ctx context.Context, name string) error {
	operationCtx, cancel := context.WithTimeout(ctx, firmwareOperationTimeout)
	defer cancel()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-operationCtx.Done():
			status, statusErr := r.client.Core().FirmwareUpgradeStatus(ctx)
			return fmt.Errorf(
				"waiting for foundation plugin %q removal timed out: %w; firmware status=%s",
				name,
				operationCtx.Err(),
				firmwareStatusDescription(status, statusErr),
			)
		case <-ticker.C:
			status, statusErr := r.client.Core().FirmwareUpgradeStatus(operationCtx)
			if statusErr != nil || status == nil || !strings.EqualFold(strings.TrimSpace(status.Status), "done") {
				continue
			}

			// Removing os-api-extensions also removes the local package-status
			// endpoint. A single firmware inventory read is therefore required
			// after the mutation has completed; this path is never used by Read.
			plugin, err := r.findFirmwarePlugin(operationCtx, name)
			if err != nil {
				return fmt.Errorf("unable to verify foundation plugin %q removal: %w", name, err)
			}
			if plugin == nil || !firmwareFlag(plugin.Installed) {
				return nil
			}
			return fmt.Errorf("firmware operation completed but foundation plugin %q is still installed", name)
		}
	}
}

func (r *pluginResource) setLock(ctx context.Context, name string, desired bool) error {
	current, err := r.findLocalPlugin(ctx, name)
	if err != nil {
		return err
	}
	if !current.Installed {
		return fmt.Errorf("plugin %q is not installed", name)
	}
	if current.Locked == desired {
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
	plugin, err := r.findLocalPlugin(ctx, name)
	if err != nil {
		return nil, err
	}
	state := &pluginResourceModel{
		Name:               types.StringValue(name),
		UninstallOnDestroy: types.BoolValue(uninstallOnDestroy),
		ID:                 types.StringValue(name),
		Installed:          types.BoolValue(plugin.Installed),
		Locked:             types.BoolValue(plugin.Locked),
		Provided:           types.BoolValue(plugin.Provided),
		Version:            types.StringValue(plugin.Version),
		Repository:         types.StringValue(plugin.Repository),
	}
	tflog.Trace(ctx, "read local OPNsense plugin state", map[string]any{"name": name, "installed": plugin.Installed})
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

func firmwareStatusDescription(status *apicore.FirmwareUpgradeStatusResponse, err error) string {
	if err != nil {
		return fmt.Sprintf("unavailable (%v)", err)
	}
	if status == nil || strings.TrimSpace(status.Status) == "" {
		return "unknown"
	}
	return fmt.Sprintf("%q", status.Status)
}
