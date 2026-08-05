package bind

import (
	"testing"

	"github.com/biptec/opnsense-go/pkg/api"
	apibind "github.com/biptec/opnsense-go/pkg/bind"
	"github.com/biptec/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBindViewRoundTrip(t *testing.T) {
	model := &viewResourceModel{
		Enabled: types.BoolValue(true), Sequence: types.Int64Value(10), Name: types.StringValue("internal"),
		MatchAny: types.BoolValue(false), MatchClientACLs: tools.StringSliceToSet([]string{"acl-a"}),
		Recursion: types.BoolValue(true), AllowRecursion: tools.StringSliceToSet([]string{"acl-a"}),
		AllowQueryAny: types.BoolValue(false), AllowQuery: tools.StringSliceToSet([]string{"acl-a"}),
		AllowTransfer: tools.StringSliceToSet([]string{"acl-secondary"}), Forwarders: tools.StringSliceToSet([]string{"1.1.1.1"}),
		DNSSECValidation: types.StringValue("auto"),
	}
	remote, err := viewModelToAPI(model)
	if err != nil {
		t.Fatalf("viewModelToAPI() error = %v", err)
	}
	if remote.Sequence != "10" || remote.MatchClients.String() != "acl-a" || remote.DNSSECValidation.String() != "auto" {
		t.Fatalf("unexpected API view: %+v", remote)
	}
	state, err := viewAPIToModel(remote)
	if err != nil {
		t.Fatalf("viewAPIToModel() error = %v", err)
	}
	if !state.Recursion.ValueBool() || state.Sequence.ValueInt64() != 10 || state.Forwarders.Elements()[0].String() == "" {
		t.Fatalf("unexpected view state: %+v", state)
	}
}

func TestBindPrimaryDomainRoundTrip(t *testing.T) {
	model := &primaryDomainResourceModel{
		ViewID: types.StringValue("view-id"), DomainName: types.StringValue("example.test"), Enabled: types.BoolValue(true),
		AllowTransferACLs: tools.StringSliceToSet([]string{"secondary-acl"}), AllowRndcTransfer: types.BoolValue(false),
		AllowQueryACLs: tools.StringSliceToSet([]string{"public-acl"}), AllowRndcUpdate: types.BoolValue(false),
		UpdateKeyIDs: tools.StringSliceToSet([]string{"key-id"}), UpdatePolicy: types.StringValue("zonesub_txt"), DNSSEC: types.BoolValue(true),
		Serial: types.StringUnknown(), TTL: types.Int64Value(300), Refresh: types.Int64Value(600), Retry: types.Int64Value(300),
		Expire: types.Int64Value(86400), NegativeTTL: types.Int64Value(300), MailAdmin: types.StringValue("hostmaster@example.test"), DNSServer: types.StringValue("ns1.example.test"),
	}
	remote, err := primaryDomainModelToAPI(model)
	if err != nil {
		t.Fatalf("primaryDomainModelToAPI() error = %v", err)
	}
	if remote.View.String() != "view-id" || remote.UpdateKeys.String() != "key-id" || remote.UpdatePolicy.String() != "zonesub_txt" || remote.DNSSEC != "1" || remote.Serial != "" {
		t.Fatalf("unexpected API primary domain: %+v", remote)
	}
	remote.Serial = "2026080501"
	state, err := primaryDomainAPIToModel(remote)
	if err != nil {
		t.Fatalf("primaryDomainAPIToModel() error = %v", err)
	}
	if state.Serial.ValueString() != "2026080501" || !state.DNSSEC.ValueBool() || state.UpdateKeyIDs.Elements()[0].String() == "" {
		t.Fatalf("unexpected primary domain state: %+v", state)
	}
}

func TestBindSettingsRoundTrip(t *testing.T) {
	remote := &apibind.SettingsResponse{General: apibind.GeneralSettings{
		Enabled: "1", ListenIPv4: api.SelectedMapList{"10.0.0.1", "192.0.2.53"}, ListenIPv6: api.SelectedMapList{"::1"},
		Port: "53", LogSize: "10", LogLevel: api.SelectedMap("info"), MaxCacheSize: "25", DNSSECValidation: api.SelectedMap("auto"),
		HideHostname: "1", HideVersion: "1", EnableRateLimiting: "1", RateLimitCount: "20",
	}}
	model, err := settingsAPIToModel(remote)
	if err != nil {
		t.Fatalf("settingsAPIToModel() error = %v", err)
	}
	if model.ID.ValueString() != "bind_settings" || model.Port.ValueInt64() != 53 || model.ListenIPv4.Elements()[0].String() == "" {
		t.Fatalf("unexpected settings state: %+v", model)
	}
	model.Port = types.Int64Value(53530)
	model.Forwarders = tools.StringSliceToSet([]string{"1.1.1.1"})
	applySettingsModel(&remote.General, model)
	if remote.General.Port != "53530" || remote.General.Forwarders.String() != "1.1.1.1" || remote.General.HideVersion != "1" {
		t.Fatalf("unexpected settings API model: %+v", remote.General)
	}
}

func TestBindSecretsAreSensitive(t *testing.T) {
	tsig := tsigKeyResourceSchema().Attributes["secret"]
	if !tsig.IsSensitive() {
		t.Fatal("TSIG secret must be sensitive")
	}
	secondary := secondaryDomainResourceSchema().Attributes["transfer_key"]
	if !secondary.IsSensitive() {
		t.Fatal("secondary transfer key must be sensitive")
	}
}
