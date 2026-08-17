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

func TestCarpHealthCheckAutomaticInterfaceRoundTrip(t *testing.T) {
	model := &carpHealthCheckResourceModel{
		Enabled: types.BoolValue(true), Name: types.StringValue("wan-l2"),
		Interface: types.StringValue("opt2"), Target: types.StringValue("192.0.2.2"),
	}
	remote, err := carpHealthCheckToAPI(model)
	if err != nil {
		t.Fatalf("carpHealthCheckToAPI(): %v", err)
	}
	if remote.Scope.String() != "interface" || remote.VHID != "0" || remote.FailureAdvSkew != "254" || len(remote.VHIDTargets) != 0 {
		t.Fatalf("unexpected automatic API check: %#v", remote)
	}
	remote.Interface = api.SelectedMap("opt3")
	state := carpHealthCheckFromAPI(remote, "11111111-2222-4333-8444-555555555555")
	if state.Interface.ValueString() != "opt3" || state.Scope.ValueString() != "interface" || state.FailureAdvSkew.ValueInt64() != 254 || state.ID.ValueString() == "" {
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
		Scope: types.StringValue("vhid"), VHID: types.Int64Value(51), FailureAdvSkew: types.Int64Value(200),
	}
	remote, err := carpHealthCheckToAPI(model)
	if err != nil {
		t.Fatalf("carpHealthCheckToAPI(): %v", err)
	}
	if remote.Scope.String() != "vhid" || remote.VHID != "51" || remote.FailureAdvSkew != "200" {
		t.Fatalf("unexpected scoped API check: %#v", remote)
	}
	state := carpHealthCheckFromAPI(remote, "11111111-2222-4333-8444-555555555555")
	if state.Scope.ValueString() != "vhid" || state.VHID.ValueInt64() != 51 || state.FailureAdvSkew.ValueInt64() != 200 {
		t.Fatalf("unexpected scoped Terraform state: %#v", state)
	}
}

func TestCarpHealthCheckVHIDGroupAndFallbackRoundTrip(t *testing.T) {
	model := &carpHealthCheckResourceModel{
		Enabled: types.BoolValue(true), Name: types.StringValue("wan-fallback"),
		Interface: types.StringValue("wan"), Target: types.StringValue("192.0.2.1"),
		Scope: types.StringValue("vhid_group"), FailureAdvSkew: types.Int64Value(200),
		VHIDTargets:        stringSetValue([]string{"opt3:52", "opt2:51"}),
		FallbackIPv4Target: types.StringValue("192.0.2.2"), FallbackIPv4Gateway: types.StringValue("10.16.224.5"),
		FallbackIPv6Target: types.StringValue("2001:db8:1::2"), FallbackIPv6Gateway: types.StringValue("2001:db8:2::1"),
		FallbackIPv4DefaultGateway: types.StringValue("10.16.224.6"), FallbackIPv6DefaultGateway: types.StringValue("2001:db8:2::2"),
	}
	remote, err := carpHealthCheckToAPI(model)
	if err != nil {
		t.Fatalf("carpHealthCheckToAPI(): %v", err)
	}
	if remote.Scope.String() != "vhid_group" || remote.VHID != "0" || remote.FailureAdvSkew != "200" || len(remote.VHIDTargets) != 2 ||
		remote.VHIDTargets[0] != "opt2:51" || remote.VHIDTargets[1] != "opt3:52" || remote.FallbackIPv4Gateway != "10.16.224.5" || remote.FallbackIPv6Target != "2001:db8:1::2" ||
		remote.FallbackIPv4DefaultGateway != "10.16.224.6" || remote.FallbackIPv6DefaultGateway != "2001:db8:2::2" {
		t.Fatalf("unexpected grouped API check: %#v", remote)
	}
	state := carpHealthCheckFromAPI(remote, "11111111-2222-4333-8444-555555555555")
	targets, err := stringSet(t.Context(), state.VHIDTargets)
	if err != nil || len(targets) != 2 || targets[0] != "opt2:51" || targets[1] != "opt3:52" {
		t.Fatalf("unexpected grouped Terraform targets: %v, err=%v", targets, err)
	}
	if state.FallbackIPv4Target.ValueString() != "192.0.2.2" || state.FallbackIPv6Gateway.ValueString() != "2001:db8:2::1" ||
		state.FallbackIPv4DefaultGateway.ValueString() != "10.16.224.6" || state.FallbackIPv6DefaultGateway.ValueString() != "2001:db8:2::2" {
		t.Fatalf("unexpected grouped Terraform state: %#v", state)
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

func TestCarpHealthCheckVHIDGroupRequiresTargets(t *testing.T) {
	model := &carpHealthCheckResourceModel{
		Enabled: types.BoolValue(true), Name: types.StringValue("bad-group"),
		Interface: types.StringValue("opt2"), Target: types.StringValue("192.0.2.2"), Scope: types.StringValue("vhid_group"),
	}
	if _, err := carpHealthCheckToAPI(model); err == nil {
		t.Fatal("scope=vhid_group accepted an empty target set")
	}
}

func TestCarpHealthCheckAutomaticScopeRejectsExplicitTargets(t *testing.T) {
	model := &carpHealthCheckResourceModel{
		Enabled: types.BoolValue(true), Name: types.StringValue("bad-auto"),
		Interface: types.StringValue("opt2"), Target: types.StringValue("192.0.2.2"), Scope: types.StringValue("interface"),
		VHIDTargets: stringSetValue([]string{"opt2:51"}),
	}
	if _, err := carpHealthCheckToAPI(model); err == nil {
		t.Fatal("automatic interface scope accepted explicit vhid_targets")
	}
}

func TestCarpHealthCheckFallbackRequiresPair(t *testing.T) {
	model := &carpHealthCheckResourceModel{
		Enabled: types.BoolValue(true), Name: types.StringValue("bad-route"),
		Interface: types.StringValue("opt2"), Target: types.StringValue("192.0.2.2"),
		FallbackIPv4Target: types.StringValue("192.0.2.3"),
	}
	if _, err := carpHealthCheckToAPI(model); err == nil {
		t.Fatal("fallback target without gateway was accepted")
	}
}

func TestCarpHealthCheckGlobalScopeRejectsDefaultFallback(t *testing.T) {
	model := &carpHealthCheckResourceModel{
		Enabled: types.BoolValue(true), Name: types.StringValue("bad-default"),
		Interface: types.StringValue("wan"), Target: types.StringValue("192.0.2.1"), Scope: types.StringValue("global"),
		FallbackIPv4DefaultGateway: types.StringValue("10.16.224.6"),
	}
	if _, err := carpHealthCheckToAPI(model); err == nil {
		t.Fatal("global scope accepted fallback default routing")
	}
}

func TestCarpHealthCheckLegacyScopeDefaultsGlobal(t *testing.T) {
	remote := &apiextensions.CarpHealthCheck{Enabled: "1", Name: "legacy", Interface: api.SelectedMap("opt2"), Target: "192.0.2.2"}
	state := carpHealthCheckFromAPI(remote, "11111111-2222-4333-8444-555555555555")
	if state.Scope.ValueString() != "global" || state.VHID.ValueInt64() != 0 || state.FailureAdvSkew.ValueInt64() != 254 {
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

func TestCarpHealthStatusRuntimeLists(t *testing.T) {
	desired, configured, current := 200, 10, 200
	vhids, diagnostics := carpHealthStatusVHIDList(t.Context(), []apiextensions.CarpHealthVHIDStatus{{
		Key: "opt2:51", Interface: "opt2", Device: "vlan02", VHID: 51,
		Checks: []string{"check-1"}, Ready: true, Healthy: false, DesiredDemoted: true,
		DesiredAdvSkew: &desired, ConfiguredAdvSkew: &configured, CurrentAdvSkew: &current,
		CarpState: "BACKUP", ControlOK: true,
	}})
	if diagnostics.HasError() || len(vhids.Elements()) != 1 {
		t.Fatalf("unexpected VHID runtime list: %v, diagnostics=%v", vhids, diagnostics)
	}

	routes, diagnostics := carpHealthStatusRouteList(t.Context(), []apiextensions.CarpHealthRouteStatus{{
		Key: "inet:network:0.0.0.0/1", CheckUUID: "check-1", Check: "wan-fallback", Family: "inet", RouteType: "network",
		Destination: "0.0.0.0/1", Gateway: "10.16.224.5", DesiredInstalled: true,
		Installed: true, Managed: true, ControlOK: true,
	}})
	if diagnostics.HasError() || len(routes.Elements()) != 1 {
		t.Fatalf("unexpected route runtime list: %v, diagnostics=%v", routes, diagnostics)
	}

	checks := []carpHealthStatusCheckModel{{
		UUID: types.StringValue("check-1"), Name: types.StringValue("wan-fallback"),
		Interface: types.StringValue("wan"), Device: types.StringValue("vtnet1"), Target: types.StringValue("192.0.2.1"),
		Scope: types.StringValue("vhid_group"), VHID: types.Int64Value(0),
		VHIDTargets: stringSetValue([]string{"opt2:51"}), ConfiguredVHIDTargets: stringSetValue([]string{"opt2:51"}),
		FailureAdvSkew: types.Int64Value(200), VHIDStates: vhids, FallbackRoutes: routes,
		CarpState: types.StringValue("BACKUP"), ConfiguredAdvSkew: types.Int64Value(10), CurrentAdvSkew: types.Int64Value(200),
		ControlOK: types.BoolValue(true), Healthy: types.BoolValue(false), Failures: types.Int64Value(2), Successes: types.Int64Value(0),
	}}
	checkList, diagnostics := types.ListValueFrom(t.Context(), types.ObjectType{AttrTypes: carpHealthStatusCheckTypes}, checks)
	if diagnostics.HasError() || len(checkList.Elements()) != 1 {
		t.Fatalf("unexpected check runtime list: %v, diagnostics=%v", checkList, diagnostics)
	}
}
