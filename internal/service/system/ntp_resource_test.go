package system

import (
	"context"
	"testing"

	apiextensions "github.com/biptec/opnsense-go/pkg/api_extensions"
)

type fakeNTPSettingsClient struct {
	current          apiextensions.NtpSettings
	getCalls         int
	setCalls         int
	reconfigureCalls int
}

func (f *fakeNTPSettingsClient) NtpGet(context.Context) (*apiextensions.NtpResponse, error) {
	f.getCalls++
	return &apiextensions.NtpResponse{NTP: f.current}, nil
}

func (f *fakeNTPSettingsClient) NtpSet(_ context.Context, settings *apiextensions.NtpSettings) (*apiextensions.NtpActionResult, error) {
	f.setCalls++
	f.current = *settings
	return &apiextensions.NtpActionResult{Status: "ok"}, nil
}

func (f *fakeNTPSettingsClient) NtpReconfigure(context.Context) (*apiextensions.NtpActionResult, error) {
	f.reconfigureCalls++
	return &apiextensions.NtpActionResult{Status: "ok"}, nil
}

func testNTPSettings() *apiextensions.NtpSettings {
	return &apiextensions.NtpSettings{
		Enabled:    true,
		Servers:    []apiextensions.NtpServer{{Host: "0.pool.ntp.org", Prefer: true, IBurst: true, Pool: true}},
		Interfaces: []string{"service_ntp1"},
		Orphan:     12,
		MaxClock:   10,

		KissOfDeath:          true,
		RateLimiting:         true,
		DenyModifications:    true,
		DisableQueries:       true,
		DenyPeerAssociations: true,
		DenyTrapService:      true,
	}
}

func TestApplyNTPSettingsCreateReconfiguresEqualRuntime(t *testing.T) {
	desired := testNTPSettings()
	client := &fakeNTPSettingsClient{current: *desired}

	if err := applyNTPSettings(context.Background(), client, desired, true); err != nil {
		t.Fatalf("applyNTPSettings() error = %v", err)
	}
	if client.setCalls != 0 {
		t.Fatalf("set calls = %d, want 0 for equal persisted settings", client.setCalls)
	}
	if client.reconfigureCalls != 1 {
		t.Fatalf("reconfigure calls = %d, want 1 on resource create", client.reconfigureCalls)
	}
}

func TestApplyNTPSettingsUpdateSkipsEqualRuntime(t *testing.T) {
	desired := testNTPSettings()
	client := &fakeNTPSettingsClient{current: *desired}

	if err := applyNTPSettings(context.Background(), client, desired, false); err != nil {
		t.Fatalf("applyNTPSettings() error = %v", err)
	}
	if client.setCalls != 0 {
		t.Fatalf("set calls = %d, want 0 for equal persisted settings", client.setCalls)
	}
	if client.reconfigureCalls != 0 {
		t.Fatalf("reconfigure calls = %d, want 0 on equal update", client.reconfigureCalls)
	}
}
