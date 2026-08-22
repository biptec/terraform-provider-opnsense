package dns

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/biptec/opnsense-go/pkg/api"
	apibind "github.com/biptec/opnsense-go/pkg/bind"
	apidnsmasq "github.com/biptec/opnsense-go/pkg/dnsmasq"
	apiunbound "github.com/biptec/opnsense-go/pkg/unbound"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestClassifyServiceState(t *testing.T) {
	tests := []struct {
		name    string
		bind    string
		unbound string
		dnsmasq string
		port    string
		want    string
	}{
		{name: "bind", bind: "1", unbound: "0", dnsmasq: "1", port: "0", want: "bind"},
		{name: "unbound", bind: "0", unbound: "1", dnsmasq: "1", port: "0", want: "unbound"},
		{name: "none", bind: "0", unbound: "0", dnsmasq: "1", port: "0", want: "none"},
		{name: "both services", bind: "1", unbound: "1", dnsmasq: "1", port: "0", want: "conflict"},
		{name: "dnsmasq owns dns", bind: "0", unbound: "1", dnsmasq: "1", port: "53", want: "conflict"},
		{name: "dnsmasq nonstandard port does not own dns", bind: "0", unbound: "1", dnsmasq: "1", port: "53053", want: "unbound"},
		{name: "disabled dnsmasq port ignored", bind: "0", unbound: "1", dnsmasq: "0", port: "53", want: "unbound"},
		{name: "invalid dnsmasq port", bind: "0", unbound: "0", dnsmasq: "1", port: "invalid", want: "conflict"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			snapshot := serviceSnapshot{
				Bind: apibind.GeneralSettings{Enabled: test.bind},
				Unbound: apiunbound.Settings{General: apiunbound.General{
					Enabled: test.unbound,
				}},
				Dnsmasq: apidnsmasq.GeneralSettings{
					IsEnabled: test.dnsmasq,
					DNS_Port:  test.port,
				},
			}
			if got := classifyServiceState(snapshot); got != test.want {
				t.Fatalf("classifyServiceState() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestPrimaryServiceOwner(t *testing.T) {
	tests := []struct {
		bind, unbound, want string
	}{
		{bind: "1", unbound: "0", want: "bind"},
		{bind: "0", unbound: "1", want: "unbound"},
		{bind: "0", unbound: "0", want: "none"},
		{bind: "1", unbound: "1", want: "conflict"},
	}
	for _, test := range tests {
		snapshot := serviceSnapshot{
			Bind:    apibind.GeneralSettings{Enabled: test.bind},
			Unbound: apiunbound.Settings{General: apiunbound.General{Enabled: test.unbound}},
		}
		if got := primaryServiceOwner(snapshot); got != test.want {
			t.Errorf("primaryServiceOwner(%q, %q) = %q, want %q", test.bind, test.unbound, got, test.want)
		}
	}
}

func TestDnsmasqPort(t *testing.T) {
	for input, want := range map[string]int{"": 53, "0": 0, "53": 53, "invalid": -1} {
		if got := dnsmasqPort(input); got != want {
			t.Errorf("dnsmasqPort(%q) = %d, want %d", input, got, want)
		}
	}
}

type fakeCutoverBackend struct {
	snapshot          serviceSnapshot
	bindStatus        string
	unboundStatus     string
	failBindEnable    bool
	failUnboundEnable bool
	calls             []string
}

func (f *fakeCutoverBackend) Observe(context.Context) (string, serviceSnapshot, error) {
	return classifyServiceState(f.snapshot), f.snapshot, nil
}

func (f *fakeCutoverBackend) WriteBind(_ context.Context, settings apibind.GeneralSettings) error {
	f.calls = append(f.calls, "bind="+settings.Enabled)
	if f.failBindEnable && settings.Enabled == "1" {
		return errors.New("simulated BIND start failure")
	}
	f.snapshot.Bind = settings
	if settings.Enabled == "1" {
		f.bindStatus = "running"
	} else {
		f.bindStatus = "stopped"
	}
	return nil
}

func (f *fakeCutoverBackend) WriteUnbound(_ context.Context, settings apiunbound.Settings) error {
	f.calls = append(f.calls, "unbound="+settings.General.Enabled)
	if f.failUnboundEnable && settings.General.Enabled == "1" {
		return errors.New("simulated Unbound start failure")
	}
	f.snapshot.Unbound = settings
	if settings.General.Enabled == "1" {
		f.unboundStatus = "running"
	} else {
		f.unboundStatus = "stopped"
	}
	return nil
}

func (f *fakeCutoverBackend) WriteDnsmasq(_ context.Context, settings apidnsmasq.GeneralSettings) error {
	f.calls = append(f.calls, "dnsmasq_port="+settings.DNS_Port)
	f.snapshot.Dnsmasq = settings
	return nil
}

func (f *fakeCutoverBackend) BindStatus(context.Context) (string, error) {
	return f.bindStatus, nil
}

func (f *fakeCutoverBackend) UnboundStatus(context.Context) (string, error) {
	return f.unboundStatus, nil
}

func unboundSnapshot() serviceSnapshot {
	return serviceSnapshot{
		Bind: apibind.GeneralSettings{Enabled: "0"},
		Unbound: apiunbound.Settings{General: apiunbound.General{
			Enabled: "1",
		}},
		Dnsmasq: apidnsmasq.GeneralSettings{
			IsEnabled: "1",
			DNS_Port:  "0",
		},
	}
}

func bindSnapshot() serviceSnapshot {
	return serviceSnapshot{
		Bind: apibind.GeneralSettings{Enabled: "1"},
		Unbound: apiunbound.Settings{General: apiunbound.General{
			Enabled: "0",
		}},
		Dnsmasq: apidnsmasq.GeneralSettings{
			IsEnabled: "1",
			DNS_Port:  "0",
		},
	}
}

func cutoverPlan(target string) *serviceCutoverResourceModel {
	return &serviceCutoverResourceModel{
		Target:               types.StringValue(target),
		AllowCutover:         types.BoolValue(true),
		VerifyTimeoutSeconds: types.Int64Value(5),
	}
}

func TestReconcileInitialToBindDoesNotRequireApproval(t *testing.T) {
	backend := &fakeCutoverBackend{
		snapshot:      unboundSnapshot(),
		bindStatus:    "stopped",
		unboundStatus: "running",
	}
	resource := &serviceCutoverResource{backend: backend}
	plan := cutoverPlan("bind")
	plan.AllowCutover = types.BoolValue(false)

	if err := resource.reconcileInitial(context.Background(), plan); err != nil {
		t.Fatalf("reconcileInitial() error = %v", err)
	}
	if got := classifyServiceState(backend.snapshot); got != "bind" {
		t.Fatalf("active service = %q, want bind", got)
	}
}

func TestReconcileManagedToBindRequiresApproval(t *testing.T) {
	backend := &fakeCutoverBackend{
		snapshot:      unboundSnapshot(),
		bindStatus:    "stopped",
		unboundStatus: "running",
	}
	resource := &serviceCutoverResource{backend: backend}
	plan := cutoverPlan("bind")
	plan.AllowCutover = types.BoolValue(false)

	err := resource.reconcile(context.Background(), plan)
	if err == nil || !strings.Contains(err.Error(), "requires allow_cutover = true") {
		t.Fatalf("reconcile() error = %v, want approval requirement", err)
	}
	if got := classifyServiceState(backend.snapshot); got != "unbound" {
		t.Fatalf("active service = %q, want unbound", got)
	}
}

func TestReconcileToBind(t *testing.T) {
	backend := &fakeCutoverBackend{
		snapshot:      unboundSnapshot(),
		bindStatus:    "stopped",
		unboundStatus: "running",
	}
	resource := &serviceCutoverResource{backend: backend}
	if err := resource.reconcile(context.Background(), cutoverPlan("bind")); err != nil {
		t.Fatalf("reconcile() error = %v", err)
	}
	if got := classifyServiceState(backend.snapshot); got != "bind" {
		t.Fatalf("active service = %q, want bind", got)
	}
	wantCalls := []string{
		"bind=0",
		"dnsmasq_port=0",
		"unbound=0",
		"bind=1",
	}
	if !reflect.DeepEqual(backend.calls, wantCalls) {
		t.Fatalf("calls = %#v, want %#v", backend.calls, wantCalls)
	}
}

func TestReconcileRollsBackBindStartFailure(t *testing.T) {
	backend := &fakeCutoverBackend{
		snapshot:       unboundSnapshot(),
		bindStatus:     "stopped",
		unboundStatus:  "running",
		failBindEnable: true,
	}
	resource := &serviceCutoverResource{backend: backend}
	err := resource.reconcile(context.Background(), cutoverPlan("bind"))
	if err == nil || !strings.Contains(err.Error(), "previous DNS state was restored") {
		t.Fatalf("reconcile() error = %v, want rollback confirmation", err)
	}
	if got := classifyServiceState(backend.snapshot); got != "unbound" {
		t.Fatalf("active service after rollback = %q, want unbound", got)
	}
	wantSuffix := []string{
		"bind=0",
		"dnsmasq_port=0",
		"unbound=1",
		"bind=0",
	}
	if len(backend.calls) < len(wantSuffix) ||
		!reflect.DeepEqual(backend.calls[len(backend.calls)-len(wantSuffix):], wantSuffix) {
		t.Fatalf("rollback calls = %#v, want suffix %#v", backend.calls, wantSuffix)
	}
}

func TestVerifyTargetRuntimeRequiresUnboundRunning(t *testing.T) {
	backend := &fakeCutoverBackend{
		snapshot:      unboundSnapshot(),
		bindStatus:    "stopped",
		unboundStatus: "stopped",
	}
	resource := &serviceCutoverResource{backend: backend}

	err := resource.verifyTargetRuntime(context.Background(), "unbound")
	if err == nil || !strings.Contains(err.Error(), `unbound status is "stopped"`) {
		t.Fatalf("verifyTargetRuntime() error = %v, want stopped Unbound error", err)
	}

	backend.unboundStatus = "running"
	if err := resource.verifyTargetRuntime(context.Background(), "unbound"); err != nil {
		t.Fatalf("verifyTargetRuntime() with running Unbound error = %v", err)
	}

	backend.bindStatus = "running"
	if err := resource.verifyTargetRuntime(context.Background(), "unbound"); err == nil || !strings.Contains(err.Error(), "BIND is still running") {
		t.Fatalf("verifyTargetRuntime() error = %v, want running BIND conflict", err)
	}
}

func TestUnboundStatus(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/api/unbound/service/status" {
			t.Fatalf("request path = %q, want Unbound service status endpoint", request.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":" RUNNING "}`))
	}))
	defer server.Close()

	backend := &opnsenseServiceCutoverBackend{
		apiClient: api.NewClient(api.Options{Uri: server.URL, AllowInsecure: true}),
	}
	status, err := backend.UnboundStatus(context.Background())
	if err != nil {
		t.Fatalf("UnboundStatus() error = %v", err)
	}
	if status != "running" {
		t.Fatalf("UnboundStatus() = %q, want running", status)
	}
}

func TestReconcileRejectsUnhealthyCurrentOwner(t *testing.T) {
	backend := &fakeCutoverBackend{
		snapshot:      unboundSnapshot(),
		bindStatus:    "stopped",
		unboundStatus: "stopped",
	}
	resource := &serviceCutoverResource{backend: backend}

	err := resource.reconcile(context.Background(), cutoverPlan("unbound"))
	if err == nil || !strings.Contains(err.Error(), `configured DNS owner "unbound" is not healthy`) {
		t.Fatalf("reconcile() error = %v, want unhealthy current owner error", err)
	}
	if len(backend.calls) != 0 {
		t.Fatalf("reconcile() changed configuration while checking current owner: %#v", backend.calls)
	}
}

func TestReconcileToUnbound(t *testing.T) {
	backend := &fakeCutoverBackend{
		snapshot:      bindSnapshot(),
		bindStatus:    "running",
		unboundStatus: "stopped",
	}
	resource := &serviceCutoverResource{backend: backend}
	if err := resource.reconcile(context.Background(), cutoverPlan("unbound")); err != nil {
		t.Fatalf("reconcile() error = %v", err)
	}
	if got := classifyServiceState(backend.snapshot); got != "unbound" {
		t.Fatalf("active service = %q, want unbound", got)
	}
	wantCalls := []string{
		"bind=0",
		"dnsmasq_port=0",
		"unbound=1",
	}
	if !reflect.DeepEqual(backend.calls, wantCalls) {
		t.Fatalf("calls = %#v, want %#v", backend.calls, wantCalls)
	}
}

func TestReconcileRollsBackUnboundStartFailure(t *testing.T) {
	backend := &fakeCutoverBackend{
		snapshot:          bindSnapshot(),
		bindStatus:        "running",
		unboundStatus:     "stopped",
		failUnboundEnable: true,
	}
	resource := &serviceCutoverResource{backend: backend}
	err := resource.reconcile(context.Background(), cutoverPlan("unbound"))
	if err == nil || !strings.Contains(err.Error(), "previous DNS state was restored") {
		t.Fatalf("reconcile() error = %v, want rollback confirmation", err)
	}
	if got := classifyServiceState(backend.snapshot); got != "bind" {
		t.Fatalf("active service after rollback = %q, want bind", got)
	}
	wantSuffix := []string{
		"bind=0",
		"dnsmasq_port=0",
		"unbound=0",
		"bind=1",
	}
	if len(backend.calls) < len(wantSuffix) ||
		!reflect.DeepEqual(backend.calls[len(backend.calls)-len(wantSuffix):], wantSuffix) {
		t.Fatalf("rollback calls = %#v, want suffix %#v", backend.calls, wantSuffix)
	}
}
