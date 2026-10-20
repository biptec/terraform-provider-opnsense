package system

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/biptec/opnsense-go/pkg/api"
	apiextensions "github.com/biptec/opnsense-go/pkg/api_extensions"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type fakeCarpHealthStatusClient struct {
	statuses []*apiextensions.CarpHealthStatusResponse
	index    int
}

func (f *fakeCarpHealthStatusClient) CarpHealthStatus(context.Context) (*apiextensions.CarpHealthStatusResponse, error) {
	if len(f.statuses) == 0 {
		return &apiextensions.CarpHealthStatusResponse{}, nil
	}
	index := f.index
	if index >= len(f.statuses) {
		index = len(f.statuses) - 1
	}
	f.index++
	return f.statuses[index], nil
}

func TestWaitForCarpHealthCheckHealthy(t *testing.T) {
	client := &fakeCarpHealthStatusClient{statuses: []*apiextensions.CarpHealthStatusResponse{
		{Status: "ok", Running: true, Checks: []apiextensions.CarpHealthCheckStatus{{UUID: "check-1", Healthy: false}}},
		{Status: "ok", Running: true, Checks: []apiextensions.CarpHealthCheckStatus{{UUID: "check-1", Healthy: true}}},
	}}

	err := waitForCarpHealthCheckHealthy(context.Background(), client, "check-1", time.Millisecond, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("waitForCarpHealthCheckHealthy() error = %v", err)
	}
	if client.index < 2 {
		t.Fatalf("status calls = %d, want at least 2", client.index)
	}
}

func TestWaitForCarpHealthCheckHealthyRequiresRunningMonitor(t *testing.T) {
	client := &fakeCarpHealthStatusClient{statuses: []*apiextensions.CarpHealthStatusResponse{
		{Status: "ok", Running: false, Checks: []apiextensions.CarpHealthCheckStatus{{UUID: "check-1", Healthy: true}}},
		{Status: "ok", Running: true, Checks: []apiextensions.CarpHealthCheckStatus{{UUID: "check-1", Healthy: true}}},
	}}

	err := waitForCarpHealthCheckHealthy(context.Background(), client, "check-1", time.Millisecond, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("waitForCarpHealthCheckHealthy() error = %v", err)
	}
	if client.index < 2 {
		t.Fatalf("status calls = %d, want at least 2", client.index)
	}
}

func TestWaitForCarpHealthCheckHealthyIgnoresOtherChecks(t *testing.T) {
	client := &fakeCarpHealthStatusClient{statuses: []*apiextensions.CarpHealthStatusResponse{
		{Status: "ok", Running: true, Checks: []apiextensions.CarpHealthCheckStatus{{UUID: "other", Healthy: true}}},
		{Status: "ok", Running: true, Checks: []apiextensions.CarpHealthCheckStatus{{UUID: "check-1", Healthy: true}}},
	}}

	err := waitForCarpHealthCheckHealthy(context.Background(), client, "check-1", time.Millisecond, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("waitForCarpHealthCheckHealthy() error = %v", err)
	}
	if client.index < 2 {
		t.Fatalf("status calls = %d, want at least 2", client.index)
	}
}

func TestWaitForCarpHealthCheckHealthyTimesOut(t *testing.T) {
	client := &fakeCarpHealthStatusClient{statuses: []*apiextensions.CarpHealthStatusResponse{
		{Status: "ok", Running: true, Checks: []apiextensions.CarpHealthCheckStatus{{UUID: "check-1", Healthy: false}}},
	}}

	err := waitForCarpHealthCheckHealthy(context.Background(), client, "check-1", time.Millisecond, 5*time.Millisecond)
	if err == nil {
		t.Fatal("waitForCarpHealthCheckHealthy() error = nil, want timeout")
	}
	if !strings.Contains(err.Error(), "did not become probe-healthy") {
		t.Fatalf("waitForCarpHealthCheckHealthy() error = %q", err)
	}
}

func TestCarpHealthReadinessTimeout(t *testing.T) {
	if got := carpHealthReadinessTimeout(nil); got != 30*time.Second {
		t.Fatalf("default readiness timeout = %s, want 30s", got)
	}
	slow := &apiextensions.CarpHealthSettings{
		Interval:          "60",
		RecoveryThreshold: "20",
	}
	if got := carpHealthReadinessTimeout(slow); got != 22*time.Minute {
		t.Fatalf("slow readiness timeout = %s, want 22m", got)
	}
}

func TestCarpHealthCheckStatePreservesWaitForHealthy(t *testing.T) {
	planned := &carpHealthCheckResourceModel{WaitForHealthy: types.BoolValue(true)}
	remote := &apiextensions.CarpHealthCheck{
		Enabled:        api.BoolString("1"),
		Name:           "wan-l2",
		Interface:      api.SelectedMap("opt1"),
		Target:         "192.0.2.1",
		Scope:          api.SelectedMap("interface"),
		FailureAdvSkew: "200",
	}

	state := carpHealthCheckStateFromRemote(planned, remote, "check-1")
	if !state.WaitForHealthy.ValueBool() {
		t.Fatal("wait_for_healthy was not preserved in state")
	}
}
