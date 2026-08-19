package quagga

import (
	"testing"

	"github.com/biptec/opnsense-go/pkg/api"
	quaggaapi "github.com/biptec/opnsense-go/pkg/quagga"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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

func TestOSPFPolicyConversions(t *testing.T) {
	prefix := ospfPrefixListToAPI(&ospfPrefixListModel{
		Enabled: types.BoolValue(true), Name: types.StringValue("nc-connected-v4"),
		SequenceNumber: types.Int64Value(10), Action: types.StringValue("deny"),
		Network: types.StringValue("10.200.0.0/24"),
	})
	if prefix.Name != "nc-connected-v4" || prefix.SequenceNumber != "10" || prefix.Action.String() != "deny" || prefix.Network != "10.200.0.0/24" {
		t.Fatalf("unexpected OSPF prefix list: %#v", prefix)
	}

	routeMap := ospfRouteMapToAPI(&ospfRouteMapModel{
		Enabled: types.BoolValue(true), Name: types.StringValue("nc-connected-v4"),
		Action: types.StringValue("deny"), RouteMapID: types.Int64Value(10),
		PrefixLists: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("prefix-id")}),
		Set:         types.StringValue(""),
	})
	if routeMap.Name != "nc-connected-v4" || routeMap.RouteMapID != "10" || len(routeMap.PrefixList) != 1 || routeMap.PrefixList[0] != "prefix-id" {
		t.Fatalf("unexpected OSPF route map: %#v", routeMap)
	}

	redistribution := ospfRedistributionToAPI(&ospfRedistributionModel{
		Enabled: types.BoolValue(true), Description: types.StringValue("redistribute endpoint connected routes"),
		Redistribute: types.StringValue("connected"), RouteMap: types.StringValue("route-map-id"),
	})
	if redistribution.Redistribute.String() != "connected" || redistribution.RouteMap.String() != "route-map-id" {
		t.Fatalf("unexpected OSPF redistribution: %#v", redistribution)
	}
}

func TestOSPF6PolicyConversions(t *testing.T) {
	prefix := ospf6PrefixListToAPI(&ospf6PrefixListModel{
		Enabled: types.BoolValue(true), Name: types.StringValue("nc-connected-v6"),
		SequenceNumber: types.Int64Value(10), Action: types.StringValue("deny"),
		Network: types.StringValue("fd00:200::/64"),
	})
	if prefix.Name != "nc-connected-v6" || prefix.SequenceNumber != "10" || prefix.Action.String() != "deny" {
		t.Fatalf("unexpected OSPFv3 prefix list: %#v", prefix)
	}

	state := ospf6RedistributionFromAPI(&quaggaapi.OSPF6Redistribution{
		Enabled: "1", Description: "connected", Redistribute: api.SelectedMap("connected"), RouteMap: api.SelectedMap("route-map-id"),
	}, "11111111-2222-4333-8444-555555555555")
	if !state.Enabled.ValueBool() || state.Redistribute.ValueString() != "connected" || state.RouteMap.ValueString() != "route-map-id" {
		t.Fatalf("unexpected OSPFv3 redistribution state: %#v", state)
	}
}
