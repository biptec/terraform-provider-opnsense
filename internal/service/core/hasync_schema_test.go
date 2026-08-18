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
	if got.ID.ValueString() != hasyncID || !got.Password.IsNull() || !got.PasswordConfigured.ValueBool() {
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
	applyHasyncModel(&remote, model, types.StringNull())
	if remote.Password != "existing-secret" {
		t.Fatalf("password overwritten: %q", remote.Password)
	}
	if remote.PfsyncInterface.String() != "opt9" || remote.SyncItems.String() != "nat,rules" {
		t.Fatalf("unexpected API overlay: %+v", remote)
	}
}

func TestPreserveHasyncConfiguredEmptyStrings(t *testing.T) {
	t.Parallel()
	state := &hasyncModel{
		PfsyncPeerIP: types.StringNull(), SynchronizeToIP: types.StringNull(), Username: types.StringNull(),
	}
	configured := &hasyncModel{
		PfsyncPeerIP: types.StringValue(""), SynchronizeToIP: types.StringValue(""), Username: types.StringValue(""),
	}
	preserveHasyncConfiguredEmptyStrings(state, configured)
	for name, value := range map[string]types.String{
		"pfsync_peer_ip": state.PfsyncPeerIP, "synchronize_to_ip": state.SynchronizeToIP, "username": state.Username,
	} {
		if value.IsNull() || value.IsUnknown() || value.ValueString() != "" {
			t.Fatalf("%s did not preserve explicit empty string: %#v", name, value)
		}
	}
}

func TestCompleteHasyncModelPreservesExistingSingletonDefaults(t *testing.T) {
	t.Parallel()
	plan := &hasyncModel{
		DisablePreempt: types.BoolUnknown(), DisconnectPPPs: types.BoolUnknown(),
		PfsyncInterface: types.StringUnknown(), PfsyncPeerIP: types.StringUnknown(),
		PfsyncVersion: types.StringUnknown(), PfsyncDefer: types.BoolUnknown(),
		SynchronizeToIP: types.StringUnknown(), VerifyPeer: types.BoolUnknown(),
		Username: types.StringUnknown(), SyncItems: types.SetUnknown(types.StringType),
		PasswordVersion: types.Int64Unknown(),
	}
	current := hasyncAPIToModel(&apicore.HasyncSettings{
		DisablePreempt: "1", DisconnectPPPs: "0", PfsyncInterface: api.SelectedMap("opt9"),
		PfsyncPeerIP: "10.0.0.2", PfsyncVersion: api.SelectedMap("1400"), PfsyncDefer: "1",
		SynchronizeToIP: "", VerifyPeer: "0", Username: "", SyncItems: api.SelectedMapList{},
	})
	completeHasyncModel(plan, current)
	if !plan.DisablePreempt.ValueBool() || plan.PfsyncInterface.ValueString() != "opt9" || plan.PfsyncPeerIP.ValueString() != "10.0.0.2" {
		t.Fatalf("existing singleton values were not preserved: %+v", plan)
	}
	if plan.PasswordVersion.ValueInt64() != 0 {
		t.Fatalf("unexpected password version: %d", plan.PasswordVersion.ValueInt64())
	}
}

func TestHasyncPasswordIsWriteOnly(t *testing.T) {
	t.Parallel()
	schema := hasyncResourceSchema()
	password := schema.Attributes["password"]
	if !password.IsSensitive() || !password.IsWriteOnly() || password.IsComputed() {
		t.Fatal("HAsync password must be sensitive write-only and not computed")
	}
	if _, ok := schema.Attributes["password_version"]; !ok {
		t.Fatal("password_version attribute missing")
	}
	if _, ok := schema.Attributes["password_configured"]; !ok {
		t.Fatal("password_configured attribute missing")
	}
	if _, ok := hasyncDataSourceSchema().Attributes["password"]; ok {
		t.Fatal("HAsync data source must not expose password")
	}
}
