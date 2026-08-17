package system

import (
	"testing"

	"github.com/biptec/opnsense-go/pkg/api"
	apiextensions "github.com/biptec/opnsense-go/pkg/api_extensions"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestCarpHealthSettingsRoundTrip(t *testing.T) {
	model := &carpHealthResourceModel{
		Enabled: types.BoolValue(true), Interval: types.Int64Value(1),
		FailureThreshold: types.Int64Value(2), RecoveryThreshold: types.Int64Value(3),
	}
	remote := carpHealthToAPI(model)
	if !remote.Enabled.Bool() || remote.Interval != "1" || remote.FailureThreshold != "2" || remote.RecoveryThreshold != "3" {
		t.Fatalf("unexpected API settings: %#v", remote)
	}
	state := carpHealthFromAPI(remote)
	if state.ID.ValueString() != "carp_health" || !state.Enabled.ValueBool() || state.RecoveryThreshold.ValueInt64() != 3 {
		t.Fatalf("unexpected Terraform state: %#v", state)
	}
	if !carpHealthEqual(remote, &apiextensions.CarpHealthSettings{Enabled: api.BoolString("1"), Interval: "1", FailureThreshold: "2", RecoveryThreshold: "3"}) {
		t.Fatal("equivalent CARP health settings were not equal")
	}
}

func TestCarpHealthCheckRoundTrip(t *testing.T) {
	model := &carpHealthCheckResourceModel{Enabled: types.BoolValue(true), Name: types.StringValue("wan-l2"), Interface: types.StringValue("opt2"), Target: types.StringValue("192.0.2.2")}
	remote, err := carpHealthCheckToAPI(model)
	if err != nil {
		t.Fatalf("carpHealthCheckToAPI(): %v", err)
	}
	if remote.Enabled != "1" || remote.Interface.String() != "opt2" || remote.Target != "192.0.2.2" {
		t.Fatalf("unexpected API check: %#v", remote)
	}
	// getCheck returns InterfaceField as a selected-map; SelectedMap.String must normalize it.
	remote.Interface = api.SelectedMap("opt3")
	state := carpHealthCheckFromAPI(remote, "11111111-2222-4333-8444-555555555555")
	if state.Interface.ValueString() != "opt3" || state.ID.ValueString() == "" {
		t.Fatalf("unexpected Terraform check state: %#v", state)
	}
}

func TestCarpHealthCheckRejectsIPv6Target(t *testing.T) {
	model := &carpHealthCheckResourceModel{Enabled: types.BoolValue(true), Name: types.StringValue("bad"), Interface: types.StringValue("opt2"), Target: types.StringValue("2001:db8::1")}
	if _, err := carpHealthCheckToAPI(model); err == nil {
		t.Fatal("IPv6 target was accepted for ARP health check")
	}
}

func TestCarpHealthActionValidation(t *testing.T) {
	if err := validateCarpHealthSet(&apiextensions.CarpHealthActionResult{Result: "saved"}); err != nil {
		t.Fatalf("valid set result rejected: %v", err)
	}
	if err := validateCarpHealthSet(&apiextensions.CarpHealthActionResult{Result: "failed"}); err == nil {
		t.Fatal("failed set result accepted")
	}
	if err := validateCarpHealthReconfigure(&apiextensions.CarpHealthActionResult{Status: "ok"}); err != nil {
		t.Fatalf("valid reconfigure rejected: %v", err)
	}
	if err := validateCarpHealthReconfigure(&apiextensions.CarpHealthActionResult{Status: "failed"}); err == nil {
		t.Fatal("failed reconfigure result accepted")
	}
}

func TestCarpHealthCheckVHIDScopeRoundTrip(t *testing.T) {
	model := &carpHealthCheckResourceModel{
		Enabled: types.BoolValue(true), Name: types.StringValue("leaf-vhid-51"),
		Interface: types.StringValue("opt2"), Target: types.StringValue("192.0.2.2"),
		Scope: types.StringValue("vhid"), VHID: types.Int64Value(51),
	}
	remote, err := carpHealthCheckToAPI(model)
	if err != nil {
		t.Fatalf("carpHealthCheckToAPI(): %v", err)
	}
	if remote.Scope.String() != "vhid" || remote.VHID != "51" {
		t.Fatalf("unexpected scoped API check: %#v", remote)
	}
	state := carpHealthCheckFromAPI(remote, "11111111-2222-4333-8444-555555555555")
	if state.Scope.ValueString() != "vhid" || state.VHID.ValueInt64() != 51 {
		t.Fatalf("unexpected scoped Terraform state: %#v", state)
	}
}

func TestCarpHealthCheckVHIDScopeRejectsZero(t *testing.T) {
	model := &carpHealthCheckResourceModel{
		Enabled: types.BoolValue(true), Name: types.StringValue("bad-vhid"),
		Interface: types.StringValue("opt2"), Target: types.StringValue("192.0.2.2"),
		Scope: types.StringValue("vhid"), VHID: types.Int64Value(0),
	}
	if _, err := carpHealthCheckToAPI(model); err == nil {
		t.Fatal("scope=vhid accepted vhid=0")
	}
}

func TestCarpHealthCheckLegacyScopeDefaultsGlobal(t *testing.T) {
	remote := &apiextensions.CarpHealthCheck{Enabled: "1", Name: "legacy", Interface: api.SelectedMap("opt2"), Target: "192.0.2.2"}
	state := carpHealthCheckFromAPI(remote, "11111111-2222-4333-8444-555555555555")
	if state.Scope.ValueString() != "global" || state.VHID.ValueInt64() != 0 {
		t.Fatalf("legacy check did not normalize to global scope: %#v", state)
	}
}

func TestNullableAdvSkew(t *testing.T) {
	if got := nullableInt(nil); !got.IsNull() {
		t.Fatalf("nil advskew must remain null, got %v", got)
	}
	value := 100
	if got := nullableInt(&value); got.IsNull() || got.ValueInt64() != 100 {
		t.Fatalf("advskew conversion failed: %v", got)
	}
}
