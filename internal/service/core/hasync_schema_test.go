package core

import (
	"testing"

	"github.com/biptec/opnsense-go/pkg/api"
	apicore "github.com/biptec/opnsense-go/pkg/core"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestHasyncAPIToModel(t *testing.T) {
	t.Parallel()
	got := hasyncAPIToModel(&apicore.HasyncSettings{
		DisablePreempt: "1", DisconnectPPPs: "0", PfsyncInterface: api.SelectedMap("opt9"),
		PfsyncPeerIP: "10.0.0.2", PfsyncVersion: api.SelectedMap("1400"), PfsyncDefer: "1",
		SynchronizeToIP: "10.0.0.2", VerifyPeer: "1", Username: "ha-sync", Password: "secret",
		SyncItems: api.SelectedMapList{"nat", "rules", "virtualip"},
	})
	if !got.DisablePreempt.ValueBool() || got.DisconnectPPPs.ValueBool() {
		t.Fatalf("unexpected bool conversion: %+v", got)
	}
	if got.PfsyncInterface.ValueString() != "opt9" || got.PfsyncVersion.ValueString() != "1400" || got.PfsyncPeerIP.ValueString() != "10.0.0.2" {
		t.Fatalf("unexpected pfsync conversion: %+v", got)
	}
	if got.ID.ValueString() != hasyncID || got.Password.ValueString() != "secret" {
		t.Fatalf("unexpected singleton/credential conversion: %+v", got)
	}
	values := map[string]bool{}
	for _, item := range got.SyncItems.Elements() {
		values[item.(types.String).ValueString()] = true
	}
	for _, want := range []string{"nat", "rules", "virtualip"} {
		if !values[want] {
			t.Fatalf("sync_items missing %q: %#v", want, values)
		}
	}
}

func TestApplyHasyncModelPreservesUnsetPassword(t *testing.T) {
	t.Parallel()
	remote := apicore.HasyncSettings{Password: "existing-secret"}
	model := &hasyncModel{
		DisablePreempt: types.BoolValue(false), DisconnectPPPs: types.BoolValue(false),
		PfsyncInterface: types.StringValue("opt9"), PfsyncPeerIP: types.StringValue("10.0.0.2"),
		PfsyncVersion: types.StringValue("1400"), PfsyncDefer: types.BoolValue(false),
		SynchronizeToIP: types.StringValue("10.0.0.2"), VerifyPeer: types.BoolValue(false),
		Username: types.StringValue("ha-sync"), Password: types.StringNull(),
		SyncItems: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("nat"), types.StringValue("rules")}),
	}
	applyHasyncModel(&remote, model)
	if remote.Password != "existing-secret" {
		t.Fatalf("password overwritten: %q", remote.Password)
	}
	if remote.PfsyncInterface.String() != "opt9" || remote.SyncItems.String() != "nat,rules" {
		t.Fatalf("unexpected API overlay: %+v", remote)
	}
}
