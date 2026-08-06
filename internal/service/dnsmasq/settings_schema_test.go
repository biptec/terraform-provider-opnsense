package dnsmasq

import (
	"testing"

	"github.com/biptec/opnsense-go/pkg/api"
	apidnsmasq "github.com/biptec/opnsense-go/pkg/dnsmasq"
	"github.com/biptec/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestSettingsAPIToModel(t *testing.T) {
	remote := &apidnsmasq.GeneralSettingsWrapper{Dnsmasq: apidnsmasq.GeneralSettings{
		IsEnabled:              "1",
		Interface:              api.SelectedMapList{"wan", "lan"},
		StrictInterfaceBinding: "1",
		DNS_Port:               "53053",
	}}
	model := settingsAPIToModel(remote)
	if !model.Enabled.ValueBool() || !model.StrictInterfaceBinding.ValueBool() {
		t.Fatalf("expected enabled strict binding model: %#v", model)
	}
	if model.DNSPort.ValueInt64() != 53053 {
		t.Fatalf("expected DNS port 53053, got %d", model.DNSPort.ValueInt64())
	}
}
func TestApplySettingsModelPreservesUnmanagedFields(t *testing.T) {
	remote := apidnsmasq.GeneralSettings{
		IsEnabled:   "1",
		DNS_Port:    "53053",
		DNS_DNSSEC:  "1",
		DNS_NoIdent: "1",
		DHCPSettings: apidnsmasq.GeneralDHCPSettings{
			FQDN:                  "1",
			RegisterFirewallRules: "1",
			RouterAdvertisements:  "1",
		},
	}
	plan := &settingsResourceModel{
		Enabled:                types.BoolNull(),
		Interfaces:             types.SetNull(types.StringType),
		StrictInterfaceBinding: types.BoolNull(),
		DNSPort:                types.Int64Value(0),
	}
	applySettingsModel(&remote, plan)
	if remote.DNS_Port != "0" {
		t.Fatalf("expected DNS port 0, got %q", remote.DNS_Port)
	}
	if remote.DNS_DNSSEC != "1" || remote.DHCPSettings.RouterAdvertisements != "1" {
		t.Fatalf("unmanaged fields changed: %#v", remote)
	}
}
func TestApplySettingsModelSortsInterfaces(t *testing.T) {
	remote := apidnsmasq.GeneralSettings{}
	plan := &settingsResourceModel{
		Enabled:                types.BoolNull(),
		Interfaces:             tools.StringSliceToSet([]string{"wan", "lan"}),
		StrictInterfaceBinding: types.BoolNull(),
		DNSPort:                types.Int64Null(),
	}
	applySettingsModel(&remote, plan)
	if got := []string(remote.Interface); len(got) != 2 || got[0] != "lan" || got[1] != "wan" {
		t.Fatalf("expected sorted interfaces, got %#v", got)
	}
}

func TestValidateDnsmasqResults(t *testing.T) {
	if validateDnsmasqSetResult(&api.ActionResult{Result: "saved"}) != nil {
		t.Fatal("saved settings result should be accepted")
	}
	if validateDnsmasqSetResult(nil) == nil || validateDnsmasqSetResult(&api.ActionResult{Result: "failed"}) == nil {
		t.Fatal("invalid settings responses must be rejected")
	}
	if validateDnsmasqReconfigureResult(&api.ReconfigureStatusResult{Status: "ok"}) != nil {
		t.Fatal("ok reconfigure status should be accepted")
	}
	if validateDnsmasqReconfigureResult(nil) == nil || validateDnsmasqReconfigureResult(&api.ReconfigureStatusResult{Status: "failed"}) == nil {
		t.Fatal("invalid reconfigure responses must be rejected")
	}
}
