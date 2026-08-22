package dns

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/biptec/opnsense-go/pkg/api"
	apibind "github.com/biptec/opnsense-go/pkg/bind"
	apidnsmasq "github.com/biptec/opnsense-go/pkg/dnsmasq"
	"github.com/biptec/opnsense-go/pkg/opnsense"
	apiunbound "github.com/biptec/opnsense-go/pkg/unbound"
	"github.com/biptec/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const serviceCutoverID = "dns_service_cutover"

var _ resource.Resource = &serviceCutoverResource{}
var _ resource.ResourceWithConfigure = &serviceCutoverResource{}
var _ resource.ResourceWithImportState = &serviceCutoverResource{}
var _ resource.ResourceWithModifyPlan = &serviceCutoverResource{}

type serviceCutoverResource struct {
	backend serviceCutoverBackend
}

type serviceCutoverBackend interface {
	Observe(context.Context) (string, serviceSnapshot, error)
	WriteBind(context.Context, apibind.GeneralSettings) error
	WriteUnbound(context.Context, apiunbound.Settings) error
	WriteDnsmasq(context.Context, apidnsmasq.GeneralSettings) error
	BindStatus(context.Context) (string, error)
	UnboundStatus(context.Context) (string, error)
}

type opnsenseServiceCutoverBackend struct {
	client    opnsense.Client
	apiClient *api.Client
}

type serviceSnapshot struct {
	Bind    apibind.GeneralSettings
	Unbound apiunbound.Settings
	Dnsmasq apidnsmasq.GeneralSettings
}

func newServiceCutoverResource() resource.Resource { return &serviceCutoverResource{} }

func (r *serviceCutoverResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_service_cutover"
}

func (r *serviceCutoverResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = serviceCutoverResourceSchema()
}

func (r *serviceCutoverResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *api.Client, got %T.", req.ProviderData))
		return
	}
	r.backend = &opnsenseServiceCutoverBackend{
		client:    opnsense.NewClient(client),
		apiClient: client,
	}
}

func (r *serviceCutoverResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan serviceCutoverResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.reconcileInitial(ctx, &plan); err != nil {
		resp.Diagnostics.AddError("Unable to Change DNS Service Owner", err.Error())
		return
	}
	plan.ID = types.StringValue(serviceCutoverID)
	plan.ActiveService = types.StringValue(plan.Target.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *serviceCutoverResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state serviceCutoverResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	active, _, err := r.backend.Observe(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read DNS Service Owner", err.Error())
		return
	}
	state.ID = types.StringValue(serviceCutoverID)
	state.ActiveService = types.StringValue(active)
	state.Target = types.StringValue(active)
	if active != "bind" && active != "unbound" {
		resp.Diagnostics.AddWarning(
			"Inconsistent DNS Service Ownership",
			fmt.Sprintf("Observed DNS service state %q. A subsequent plan will require allow_cutover = true to reconcile it.", active),
		)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *serviceCutoverResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan serviceCutoverResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.reconcile(ctx, &plan); err != nil {
		resp.Diagnostics.AddError("Unable to Change DNS Service Owner", err.Error())
		return
	}
	plan.ID = types.StringValue(serviceCutoverID)
	plan.ActiveService = types.StringValue(plan.Target.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *serviceCutoverResource) Delete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning(
		"DNS Cutover Removed From State Only",
		"The active DNS service remains unchanged in OPNsense. Set target = \"unbound\" before removing this resource when an operational rollback is required.",
	)
}

func (r *serviceCutoverResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != serviceCutoverID {
		resp.Diagnostics.AddError("Invalid Import ID", fmt.Sprintf("DNS service cutover must be imported with ID %s, got %q.", serviceCutoverID, req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

func (r *serviceCutoverResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}
	var state, plan serviceCutoverResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() || state.Target.IsNull() || plan.Target.IsUnknown() || plan.AllowCutover.IsUnknown() {
		return
	}
	if !state.Target.Equal(plan.Target) && !plan.AllowCutover.ValueBool() {
		resp.Diagnostics.AddError(
			"DNS Cutover Requires Explicit Approval",
			"Changing the active DNS service can interrupt name resolution. Set allow_cutover = true for the planned transition and return it to false after the apply succeeds.",
		)
	}
}

func (r *serviceCutoverResource) reconcile(ctx context.Context, plan *serviceCutoverResourceModel) error {
	return r.reconcileServiceOwner(ctx, plan, true)
}

func (r *serviceCutoverResource) reconcileInitial(ctx context.Context, plan *serviceCutoverResourceModel) error {
	return r.reconcileServiceOwner(ctx, plan, false)
}

func (r *serviceCutoverResource) reconcileServiceOwner(ctx context.Context, plan *serviceCutoverResourceModel, requireCutoverApproval bool) error {
	target := plan.Target.ValueString()
	active, snapshot, err := r.backend.Observe(ctx)
	if err != nil {
		return err
	}
	if active == target {
		if runtimeErr := r.verifyTargetRuntime(ctx, target); runtimeErr != nil {
			return fmt.Errorf("configured DNS owner %q is not healthy: %w", target, runtimeErr)
		}
		return nil
	}
	if active == "conflict" && primaryServiceOwner(snapshot) == "conflict" {
		return fmt.Errorf("refusing cutover while BIND and Unbound are simultaneously enabled")
	}
	if requireCutoverApproval && !plan.AllowCutover.ValueBool() {
		return fmt.Errorf("changing DNS ownership from %q to %q requires allow_cutover = true", active, target)
	}

	if target == "bind" && primaryServiceOwner(snapshot) != "bind" {
		if err := r.backend.WriteBind(ctx, snapshot.Bind); err != nil {
			return fmt.Errorf("BIND preflight failed; the previous DNS owner remains active: %w", err)
		}
	}

	var transitionErr error
	switch target {
	case "bind":
		transitionErr = r.transitionToBind(ctx)
	case "unbound":
		transitionErr = r.transitionToUnbound(ctx)
	default:
		transitionErr = fmt.Errorf("unsupported DNS target %q", target)
	}
	if transitionErr != nil {
		rollbackErr := r.restore(ctx, snapshot)
		if rollbackErr != nil {
			return fmt.Errorf("cutover failed: %w; rollback also failed: %v", transitionErr, rollbackErr)
		}
		return fmt.Errorf("cutover failed and previous DNS state was restored: %w", transitionErr)
	}

	timeout := time.Duration(plan.VerifyTimeoutSeconds.ValueInt64()) * time.Second
	if err := r.waitForTarget(ctx, target, timeout); err != nil {
		rollbackErr := r.restore(ctx, snapshot)
		if rollbackErr != nil {
			return fmt.Errorf("cutover verification failed: %w; rollback also failed: %v", err, rollbackErr)
		}
		return fmt.Errorf("cutover verification failed and previous DNS state was restored: %w", err)
	}
	return nil
}

func (r *serviceCutoverResource) transitionToBind(ctx context.Context) error {
	if err := r.setDnsmasqPort(ctx, 0); err != nil {
		return fmt.Errorf("disable dnsmasq DNS listener: %w", err)
	}
	if err := r.setUnboundEnabled(ctx, false); err != nil {
		return fmt.Errorf("stop Unbound: %w", err)
	}
	if err := r.setBindEnabled(ctx, true); err != nil {
		return fmt.Errorf("start BIND: %w", err)
	}
	return nil
}

func (r *serviceCutoverResource) transitionToUnbound(ctx context.Context) error {
	if err := r.setBindEnabled(ctx, false); err != nil {
		return fmt.Errorf("stop BIND: %w", err)
	}
	if err := r.setDnsmasqPort(ctx, 0); err != nil {
		return fmt.Errorf("disable dnsmasq DNS listener: %w", err)
	}
	if err := r.setUnboundEnabled(ctx, true); err != nil {
		return fmt.Errorf("start Unbound: %w", err)
	}
	return nil
}

func (b *opnsenseServiceCutoverBackend) Observe(ctx context.Context) (string, serviceSnapshot, error) {
	bindSettings, err := b.client.Bind().SettingsGet(ctx)
	if err != nil {
		return "", serviceSnapshot{}, fmt.Errorf("read BIND settings: %w", err)
	}
	if bindSettings == nil {
		return "", serviceSnapshot{}, fmt.Errorf("BIND settings API returned an empty response")
	}
	unboundSettings, err := b.client.Unbound().SettingsGet(ctx)
	if err != nil {
		return "", serviceSnapshot{}, fmt.Errorf("read Unbound settings: %w", err)
	}
	if unboundSettings == nil {
		return "", serviceSnapshot{}, fmt.Errorf("unbound settings API returned an empty response")
	}
	dnsmasqSettings, err := b.client.Dnsmasq().GeneralSettingsGet(ctx)
	if err != nil {
		return "", serviceSnapshot{}, fmt.Errorf("read dnsmasq settings: %w", err)
	}
	if dnsmasqSettings == nil {
		return "", serviceSnapshot{}, fmt.Errorf("dnsmasq settings API returned an empty response")
	}

	snapshot := serviceSnapshot{
		Bind:    bindSettings.General,
		Unbound: unboundSettings.Unbound,
		Dnsmasq: dnsmasqSettings.Dnsmasq,
	}
	return classifyServiceState(snapshot), snapshot, nil
}

func primaryServiceOwner(snapshot serviceSnapshot) string {
	bindEnabled := tools.StringToBool(snapshot.Bind.Enabled)
	unboundEnabled := tools.StringToBool(snapshot.Unbound.General.Enabled)
	switch {
	case bindEnabled && !unboundEnabled:
		return "bind"
	case !bindEnabled && unboundEnabled:
		return "unbound"
	case !bindEnabled && !unboundEnabled:
		return "none"
	default:
		return "conflict"
	}
}

func classifyServiceState(snapshot serviceSnapshot) string {
	bindEnabled := tools.StringToBool(snapshot.Bind.Enabled)
	unboundEnabled := tools.StringToBool(snapshot.Unbound.General.Enabled)
	dnsmasqEnabled := tools.StringToBool(snapshot.Dnsmasq.IsEnabled)
	dnsmasqPort := dnsmasqPort(snapshot.Dnsmasq.DNS_Port)
	dnsmasqOwnsDNS := dnsmasqEnabled && (dnsmasqPort == 53 || dnsmasqPort < 0)

	switch {
	case bindEnabled && !unboundEnabled && !dnsmasqOwnsDNS:
		return "bind"
	case !bindEnabled && unboundEnabled && !dnsmasqOwnsDNS:
		return "unbound"
	case !bindEnabled && !unboundEnabled && !dnsmasqOwnsDNS:
		return "none"
	default:
		return "conflict"
	}
}

func dnsmasqPort(value string) int {
	if strings.TrimSpace(value) == "" {
		return 53
	}
	port, err := strconv.Atoi(value)
	if err != nil {
		return -1
	}
	return port
}

func (r *serviceCutoverResource) setBindEnabled(ctx context.Context, enabled bool) error {
	_, snapshot, err := r.backend.Observe(ctx)
	if err != nil {
		return err
	}
	snapshot.Bind.Enabled = tools.BoolToString(enabled)
	return r.backend.WriteBind(ctx, snapshot.Bind)
}

func (b *opnsenseServiceCutoverBackend) WriteBind(ctx context.Context, settings apibind.GeneralSettings) error {
	result, err := b.client.Bind().SettingsSet(ctx, &settings)
	if err != nil {
		return err
	}
	if result == nil || result.Result != "saved" {
		actual := "<empty>"
		if result != nil {
			actual = result.Result
		}
		return fmt.Errorf("BIND settings API returned result %q instead of %q", actual, "saved")
	}
	reconfigure, err := b.client.Bind().ServiceReconfigure(ctx)
	if err != nil {
		return err
	}
	if reconfigure == nil || reconfigure.Status != "ok" {
		actual := "<empty>"
		if reconfigure != nil {
			actual = reconfigure.Status
		}
		return fmt.Errorf("BIND reconfigure API returned status %q instead of %q", actual, "ok")
	}
	return nil
}

func (r *serviceCutoverResource) setUnboundEnabled(ctx context.Context, enabled bool) error {
	_, snapshot, err := r.backend.Observe(ctx)
	if err != nil {
		return err
	}
	snapshot.Unbound.General.Enabled = tools.BoolToString(enabled)
	return r.backend.WriteUnbound(ctx, snapshot.Unbound)
}

func (b *opnsenseServiceCutoverBackend) BindStatus(ctx context.Context) (string, error) {
	status, err := b.client.Bind().ServiceStatus(ctx)
	if err != nil {
		return "", err
	}
	if status == nil {
		return "", fmt.Errorf("BIND status API returned an empty response")
	}
	return status.Status, nil
}

func (b *opnsenseServiceCutoverBackend) UnboundStatus(ctx context.Context) (string, error) {
	status := &struct {
		Status string `json:"status"`
	}{}
	result, err := api.Call(b.apiClient, ctx, api.RPCOpts{
		Endpoint: api.Endpoint{Path: "/unbound/service/status", Method: "GET"},
	}, status)
	if err != nil {
		return "", err
	}
	if result == nil {
		return "", fmt.Errorf("unbound status API returned an empty response")
	}
	return strings.ToLower(strings.TrimSpace(result.Status)), nil
}

func (b *opnsenseServiceCutoverBackend) WriteUnbound(ctx context.Context, settings apiunbound.Settings) error {
	result, err := b.client.Unbound().SettingsUpdate(ctx, &settings)
	if err != nil {
		return err
	}
	if result == nil || result.Result != "saved" {
		actual := "<empty>"
		if result != nil {
			actual = result.Result
		}
		return fmt.Errorf("unbound settings API returned result %q instead of %q", actual, "saved")
	}
	reconfigureActions := []struct {
		name string
		run  func(context.Context) (*apiunbound.ActionResult, error)
	}{
		{name: "service", run: b.client.Unbound().SettingsReconfigure},
		{name: "general", run: b.client.Unbound().SettingsReconfigureGeneral},
	}
	for _, reconfigure := range reconfigureActions {
		action, actionErr := reconfigure.run(ctx)
		if actionErr != nil {
			return fmt.Errorf("%s reconfigure: %w", reconfigure.name, actionErr)
		}
		if action == nil || action.Status != "ok" {
			actual := "<empty>"
			if action != nil {
				actual = action.Status
			}
			return fmt.Errorf("unbound %s reconfigure returned status %q instead of %q", reconfigure.name, actual, "ok")
		}
	}
	return nil
}

func (r *serviceCutoverResource) setDnsmasqPort(ctx context.Context, port int) error {
	_, snapshot, err := r.backend.Observe(ctx)
	if err != nil {
		return err
	}
	snapshot.Dnsmasq.DNS_Port = strconv.Itoa(port)
	return r.backend.WriteDnsmasq(ctx, snapshot.Dnsmasq)
}

func (b *opnsenseServiceCutoverBackend) WriteDnsmasq(ctx context.Context, settings apidnsmasq.GeneralSettings) error {
	result, err := b.client.Dnsmasq().GeneralSettingsSet(ctx, &settings)
	if err != nil {
		return err
	}
	if result == nil || result.Result != "saved" {
		actual := "<empty>"
		if result != nil {
			actual = result.Result
		}
		return fmt.Errorf("dnsmasq settings API returned result %q instead of %q", actual, "saved")
	}
	reconfigure, err := b.client.Dnsmasq().ServiceReconfigure(ctx)
	if err != nil {
		return err
	}
	if reconfigure == nil || reconfigure.Status != "ok" {
		actual := "<empty>"
		if reconfigure != nil {
			actual = reconfigure.Status
		}
		return fmt.Errorf("dnsmasq reconfigure API returned status %q instead of %q", actual, "ok")
	}
	return nil
}

func (r *serviceCutoverResource) restore(ctx context.Context, snapshot serviceSnapshot) error {
	var rollbackErrors []error

	// Always free port 53 before restoring the previous listeners.
	if err := r.setBindEnabled(ctx, false); err != nil {
		rollbackErrors = append(rollbackErrors, fmt.Errorf("disable BIND: %w", err))
	}
	if err := r.backend.WriteDnsmasq(ctx, snapshot.Dnsmasq); err != nil {
		rollbackErrors = append(rollbackErrors, fmt.Errorf("restore dnsmasq: %w", err))
	}
	if err := r.backend.WriteUnbound(ctx, snapshot.Unbound); err != nil {
		rollbackErrors = append(rollbackErrors, fmt.Errorf("restore Unbound: %w", err))
	}
	if err := r.backend.WriteBind(ctx, snapshot.Bind); err != nil {
		rollbackErrors = append(rollbackErrors, fmt.Errorf("restore BIND: %w", err))
	}
	return errors.Join(rollbackErrors...)
}

func (r *serviceCutoverResource) verifyTargetRuntime(ctx context.Context, target string) error {
	bindStatus, err := r.backend.BindStatus(ctx)
	if err != nil {
		return fmt.Errorf("read BIND runtime status: %w", err)
	}
	bindStatus = strings.ToLower(strings.TrimSpace(bindStatus))

	switch target {
	case "bind":
		if bindStatus != "running" {
			return fmt.Errorf("BIND status is %q", bindStatus)
		}
		return nil
	case "unbound":
		if bindStatus == "running" {
			return fmt.Errorf("BIND is still running")
		}
		unboundStatus, statusErr := r.backend.UnboundStatus(ctx)
		if statusErr != nil {
			return fmt.Errorf("read Unbound runtime status: %w", statusErr)
		}
		unboundStatus = strings.ToLower(strings.TrimSpace(unboundStatus))
		if unboundStatus != "running" {
			return fmt.Errorf("unbound status is %q", unboundStatus)
		}
		return nil
	default:
		return fmt.Errorf("unsupported DNS target %q", target)
	}
}

func (r *serviceCutoverResource) waitForTarget(ctx context.Context, target string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var lastState string
	var lastErr error
	for time.Now().Before(deadline) {
		active, _, err := r.backend.Observe(ctx)
		if err != nil {
			lastErr = err
		} else {
			lastState = active
			if active == target {
				if runtimeErr := r.verifyTargetRuntime(ctx, target); runtimeErr == nil {
					return nil
				} else {
					lastErr = runtimeErr
				}
			}
		}

		timer := time.NewTimer(2 * time.Second)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			return ctx.Err()
		case <-timer.C:
		}
	}
	if lastErr != nil {
		return fmt.Errorf("DNS owner did not converge to %q; last observed state %q: %w", target, lastState, lastErr)
	}
	return fmt.Errorf("DNS owner did not converge to %q; last observed state %q", target, lastState)
}
