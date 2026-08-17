package quagga

import (
	"testing"

	"github.com/biptec/opnsense-go/pkg/api"
	quaggaapi "github.com/biptec/opnsense-go/pkg/quagga"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestOptionalRoutingConversions(t *testing.T) {
	if got := optionalBoolToAPI(types.BoolNull()); got != "" {
		t.Fatalf("null bool encoded as %q", got)
	}
	if got := optionalBoolToAPI(types.BoolValue(true)); got != "1" {
		t.Fatalf("true bool encoded as %q", got)
	}
	if got := optionalBoolFromAPI(""); !got.IsNull() {
		t.Fatalf("empty bool must remain null: %v", got)
	}
	if got := optionalBoolFromAPI("1"); got.IsNull() || !got.ValueBool() {
		t.Fatalf("bool decode failed: %v", got)
	}
	if got := optionalIntToAPI(types.Int64Null()); got != "" {
		t.Fatalf("null int encoded as %q", got)
	}
	if got := optionalIntToAPI(types.Int64Value(10)); got != "10" {
		t.Fatalf("int encoded as %q", got)
	}
	if got := optionalIntFromAPI(""); !got.IsNull() {
		t.Fatalf("empty int must remain null: %v", got)
	}
}

func TestRoutingActionValidation(t *testing.T) {
	if err := validateRoutingSet(&api.ActionResult{Result: "saved"}); err != nil {
		t.Fatalf("saved set rejected: %v", err)
	}
	if err := validateRoutingSet(&api.ActionResult{Result: "failed"}); err == nil {
		t.Fatal("failed set accepted")
	}
	if err := validateRoutingSet(nil); err == nil {
		t.Fatal("nil set result accepted")
	}
	if err := validateRoutingReconfigure(&api.ReconfigureStatusResult{Status: "ok"}); err != nil {
		t.Fatalf("ok reconfigure rejected: %v", err)
	}
	if err := validateRoutingReconfigure(&api.ReconfigureStatusResult{Status: "failed"}); err == nil {
		t.Fatal("failed reconfigure accepted")
	}
	if err := validateRoutingReconfigure(nil); err == nil {
		t.Fatal("nil reconfigure result accepted")
	}
}

func TestOSPFNetworkConversion(t *testing.T) {
	model := &ospfNetworkModel{
		Enabled: types.BoolValue(true), IPAddress: types.StringValue("10.16.0.0"),
		Area: types.StringValue("0.0.0.0"), Netmask: types.Int64Value(16),
		PrefixListIn: types.StringValue("internal-in"), PrefixListOut: types.StringValue("internal-out"),
	}
	remote := ospfNetworkToAPI(model)
	if remote.Enabled != "1" || remote.IPAddress != "10.16.0.0" || remote.Netmask != "16" {
		t.Fatalf("unexpected OSPF network API object: %#v", remote)
	}
	if remote.PrefixListIn.String() != "internal-in" || remote.PrefixListOut.String() != "internal-out" {
		t.Fatalf("prefix-list conversion failed: %#v", remote)
	}
	state := ospfNetworkFromAPI(remote, "11111111-2222-4333-8444-555555555555")
	if state.Netmask.ValueInt64() != 16 || state.ID.ValueString() == "" {
		t.Fatalf("unexpected OSPF network state: %#v", state)
	}
}

func TestOSPF6InterfacePreservesOptionalNulls(t *testing.T) {
	remote := &quaggaapi.OSPF6Interface{
		Enabled: "1", InterfaceName: api.SelectedMap("opt1"), Area: "0.0.0.0",
		Passive: "0", Cost: "", CostDemoted: "100", BFD: "",
	}
	state := ospf6InterfaceFromAPI(remote, "11111111-2222-4333-8444-555555555555")
	if !state.Cost.IsNull() {
		t.Fatalf("empty cost must remain null: %v", state.Cost)
	}
	if state.Passive.IsNull() || state.Passive.ValueBool() {
		t.Fatalf("passive=false conversion failed: %v", state.Passive)
	}
	if state.CostDemoted.ValueInt64() != 100 {
		t.Fatalf("cost_demoted conversion failed: %v", state.CostDemoted)
	}
	if !state.BFD.IsNull() {
		t.Fatalf("empty bfd must remain null: %v", state.BFD)
	}
}
