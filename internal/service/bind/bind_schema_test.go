package bind

import (
	"context"
	"testing"

	"github.com/biptec/opnsense-go/pkg/api"
	apibind "github.com/biptec/opnsense-go/pkg/bind"
	"github.com/biptec/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	frameworkvalidator "github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBindViewRoundTrip(t *testing.T) {
	model := &viewResourceModel{
		Enabled: types.BoolValue(true), Sequence: types.Int64Value(10), Name: types.StringValue("internal"),
		MatchAny: types.BoolValue(false), MatchClientACLs: tools.StringSliceToSet([]string{"acl-a"}),
		MatchClientTSIGKeyIDs:        tools.StringSliceToSet([]string{"tsig-internal"}),
		ExcludeMatchClientTSIGKeyIDs: tools.StringSliceToSet([]string{"tsig-public"}),
		MatchDestinationACLs:         tools.StringSliceToSet([]string{"acl-destination"}),
		Recursion:                    types.BoolValue(true), AllowRecursion: tools.StringSliceToSet([]string{"acl-a"}),
		AllowQueryAny: types.BoolValue(false), AllowQuery: tools.StringSliceToSet([]string{"acl-a"}),
		AllowTransfer: tools.StringSliceToSet([]string{"acl-secondary"}), Forwarders: tools.StringSliceToSet([]string{"1.1.1.1"}),
		DNSSECValidation: types.StringValue("auto"),
	}
	remote, err := viewModelToAPI(model)
	if err != nil {
		t.Fatalf("viewModelToAPI() error = %v", err)
	}
	if remote.Sequence != "10" || remote.MatchClients.String() != "acl-a" || remote.MatchClientTSIGKeys.String() != "tsig-internal" || remote.ExcludeMatchClientTSIGKeys.String() != "tsig-public" || remote.MatchDestinations.String() != "acl-destination" || remote.DNSSECValidation.String() != "auto" {
		t.Fatalf("unexpected API view: %+v", remote)
	}
	state, err := viewAPIToModel(remote)
	if err != nil {
		t.Fatalf("viewAPIToModel() error = %v", err)
	}
	if !state.Recursion.ValueBool() || state.Sequence.ValueInt64() != 10 || state.MatchClientTSIGKeyIDs.Elements()[0].String() == "" || state.ExcludeMatchClientTSIGKeyIDs.Elements()[0].String() == "" || state.Forwarders.Elements()[0].String() == "" {
		t.Fatalf("unexpected view state: %+v", state)
	}
}

func TestBindViewTSIGSelectorsRejectOverlap(t *testing.T) {
	included := tools.StringSliceToSet([]string{"key-a", "key-b"})
	excluded := tools.StringSliceToSet([]string{"key-c", "key-b"})
	if err := validateViewTSIGSelectors(included, excluded); err == nil {
		t.Fatal("expected overlapping TSIG view selectors to fail validation")
	}
	excluded = tools.StringSliceToSet([]string{"key-c"})
	if err := validateViewTSIGSelectors(included, excluded); err != nil {
		t.Fatalf("non-overlapping TSIG view selectors failed validation: %v", err)
	}
}

func TestBindPrimaryDomainRoundTrip(t *testing.T) {
	model := &primaryDomainResourceModel{
		ViewID: types.StringValue("view-id"), DomainName: types.StringValue("example.test"), Enabled: types.BoolValue(true),
		AllowTransferACLs: tools.StringSliceToSet([]string{"secondary-acl"}), AllowRndcTransfer: types.BoolValue(false),
		TransferKeyID: types.StringValue("transfer-key-id"), AlsoNotify: tools.StringSliceToSet([]string{"192.0.2.54"}),
		AllowQueryACLs: tools.StringSliceToSet([]string{"public-acl"}), AllowRndcUpdate: types.BoolValue(false),
		UpdateKeyIDs: tools.StringSliceToSet([]string{"key-id"}), UpdatePolicy: types.StringValue("self_txt"), DNSSEC: types.BoolValue(true),
		Serial: types.StringUnknown(), TTL: types.Int64Value(300), Refresh: types.Int64Value(600), Retry: types.Int64Value(300),
		Expire: types.Int64Value(86400), NegativeTTL: types.Int64Value(300), MailAdmin: types.StringValue("hostmaster@example.test"), DNSServer: types.StringValue("ns1.example.test"),
	}
	remote, err := primaryDomainModelToAPI(model)
	if err != nil {
		t.Fatalf("primaryDomainModelToAPI() error = %v", err)
	}
	if remote.View.String() != "view-id" || remote.TransferKey.String() != "transfer-key-id" || remote.AlsoNotify.String() != "192.0.2.54" || remote.UpdateKeys.String() != "key-id" || remote.UpdatePolicy.String() != "self_txt" || remote.DNSSEC != "1" || remote.Serial != "" {
		t.Fatalf("unexpected API primary domain: %+v", remote)
	}
	remote.Serial = "2026080501"
	state, err := primaryDomainAPIToModel(remote)
	if err != nil {
		t.Fatalf("primaryDomainAPIToModel() error = %v", err)
	}
	if state.Serial.ValueString() != "2026080501" || !state.DNSSEC.ValueBool() || state.TransferKeyID.ValueString() != "transfer-key-id" || state.AlsoNotify.Elements()[0].String() == "" || state.UpdateKeyIDs.Elements()[0].String() == "" {
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
	if !tsig.IsSensitive() || !tsig.IsWriteOnly() || tsig.IsComputed() {
		t.Fatal("TSIG secret must be sensitive, write-only, and non-computed")
	}
	if _, ok := tsigKeyDataSourceSchema().Attributes["secret"]; ok {
		t.Fatal("TSIG data source must not expose secret")
	}
	secondary := secondaryDomainResourceSchema().Attributes["transfer_key"]
	if !secondary.IsSensitive() || !secondary.IsWriteOnly() || secondary.IsComputed() {
		t.Fatal("secondary transfer key must be sensitive, write-only, and non-computed")
	}
}

func TestTsigKeyRemoteSecretDoesNotEnterState(t *testing.T) {
	remote := &apibind.TsigKey{
		Enabled: "1", Name: "_acme-challenge.web.test.invalid",
		Algorithm: api.SelectedMap("hmac-sha256"), Secret: "must-not-enter-state",
	}
	state := tsigKeyAPIToModel(remote)
	if !state.Secret.IsNull() {
		t.Fatal("remote TSIG secret must be null in Terraform state model")
	}
	if !state.SecretConfigured.ValueBool() {
		t.Fatal("secret_configured must report the remote credential presence")
	}
	data, err := tsigKeyAPIToDataSourceModel(remote)
	if err != nil {
		t.Fatalf("tsigKeyAPIToDataSourceModel() error = %v", err)
	}
	if !data.SecretConfigured.ValueBool() {
		t.Fatal("TSIG metadata data source lost credential presence flag")
	}
}

func TestSecondaryDomainRemoteSecretDoesNotEnterState(t *testing.T) {
	remote := &apibind.SecondaryDomain{View: api.SelectedMap("view-id"), DomainName: "example.test", Enabled: "1", PrimaryIP: api.SelectedMapList{"192.0.2.53"}, TransferKeyAlgorithm: api.SelectedMap("hmac-sha256"), TransferKeyName: "xfr.example.test", TransferKey: "must-not-enter-state"}
	state := secondaryDomainAPIToModel(remote)
	if !state.TransferKey.IsNull() || !state.TransferKeyConfigured.ValueBool() {
		t.Fatal("secondary transfer secret leaked into resource state or lost presence metadata")
	}
	data, err := secondaryDomainAPIToDataSourceModel(remote)
	if err != nil {
		t.Fatalf("secondaryDomainAPIToDataSourceModel() error = %v", err)
	}
	if !data.TransferKey.IsNull() || !data.TransferKeyConfigured.ValueBool() {
		t.Fatal("secondary data source must expose only transfer-key presence metadata")
	}
}

func TestApplySecondaryDomainModelPreservesOrRotatesTransferSecret(t *testing.T) {
	remote := &apibind.SecondaryDomain{View: api.SelectedMap("view-old"), DomainName: "old.test", Enabled: "1", PrimaryIP: api.SelectedMapList{"192.0.2.53"}, TransferKeyAlgorithm: api.SelectedMap("hmac-sha256"), TransferKeyName: "xfr.old.test", TransferKey: "old-secret"}
	plan := &secondaryDomainResourceModel{ViewID: types.StringValue("view-new"), DomainName: types.StringValue("new.test"), Enabled: types.BoolValue(true), PrimaryIPs: tools.StringSliceToSet([]string{"198.51.100.53"}), AllowNotify: tools.StringSliceToSet(nil), TransferKeyAlgorithm: types.StringValue("hmac-sha256"), TransferKeyName: types.StringValue("xfr.new.test"), AllowTransferACLs: tools.StringSliceToSet(nil), AllowQueryACLs: tools.StringSliceToSet(nil)}
	applySecondaryDomainModel(remote, plan, types.StringNull())
	if remote.TransferKey != "old-secret" {
		t.Fatal("metadata update must preserve the existing transfer secret")
	}
	applySecondaryDomainModel(remote, plan, types.StringValue("new-secret"))
	if remote.TransferKey != "new-secret" {
		t.Fatal("write-only rotated transfer secret was not applied")
	}
	plan.TransferKeyAlgorithm = types.StringValue("")
	plan.TransferKeyName = types.StringValue("")
	applySecondaryDomainModel(remote, plan, types.StringNull())
	if remote.TransferKey != "" {
		t.Fatal("clearing transfer metadata must clear the remote transfer secret")
	}
}

func TestApplyTsigKeyModelPreservesOrRotatesSecret(t *testing.T) {
	remote := &apibind.TsigKey{Enabled: "1", Name: "old", Algorithm: api.SelectedMap("hmac-sha256"), Secret: "old-secret"}
	plan := &tsigKeyResourceModel{Enabled: types.BoolValue(true), Name: types.StringValue("new"), Algorithm: types.StringValue("hmac-sha512")}
	applyTsigKeyModel(remote, plan, types.StringNull())
	if remote.Secret != "old-secret" || remote.Name != "new" || remote.Algorithm.String() != "hmac-sha512" {
		t.Fatalf("metadata update failed to preserve secret: %+v", remote)
	}
	applyTsigKeyModel(remote, plan, types.StringValue("new-secret"))
	if remote.Secret != "new-secret" {
		t.Fatal("write-only TSIG rotation secret was not applied")
	}
}

func TestBindSettingsPartialUpdatePreservesRemoteValues(t *testing.T) {
	remote := apibind.GeneralSettings{
		Enabled: "0", DisableIPv6: "0", EnableRPZ: "1",
		ListenIPv4: api.SelectedMapList{"0.0.0.0"}, ListenIPv6: api.SelectedMapList{"::"},
		Port: "53530", MaxCacheSize: "80", DNSSECValidation: api.SelectedMap("no"),
		RateLimitExcept: api.SelectedMapList{"0.0.0.0", "::"},
	}
	plan := &settingsResourceModel{
		Enabled:    types.BoolValue(true),
		ListenIPv4: tools.StringSliceToSet([]string{"127.0.0.1"}),
		Port:       types.Int64Value(53530),
	}

	applySettingsModel(&remote, plan)

	if remote.Enabled != "1" || remote.ListenIPv4.String() != "127.0.0.1" || remote.Port != "53530" {
		t.Fatalf("explicit settings not applied: %+v", remote)
	}
	if remote.EnableRPZ != "1" || remote.ListenIPv6.String() != "::" || remote.MaxCacheSize != "80" || remote.DNSSECValidation.String() != "no" || remote.RateLimitExcept.String() != "0.0.0.0,::" {
		t.Fatalf("omitted settings were not preserved: %+v", remote)
	}
}

func TestBindSettingsSchemaDoesNotForceDefaults(t *testing.T) {
	for name, attribute := range settingsResourceSchema().Attributes {
		if name == "id" {
			continue
		}
		var hasDefault bool
		switch typed := attribute.(type) {
		case rschema.BoolAttribute:
			hasDefault = typed.Default != nil
		case rschema.Int64Attribute:
			hasDefault = typed.Default != nil
		case rschema.StringAttribute:
			hasDefault = typed.Default != nil
		case rschema.SetAttribute:
			hasDefault = typed.Default != nil
		default:
			t.Fatalf("unexpected settings attribute type %T for %s", attribute, name)
		}
		if hasDefault {
			t.Errorf("settings attribute %s must preserve imported state, not force a Terraform default", name)
		}
	}
}

func TestValidateBindSettingsSetResult(t *testing.T) {
	if err := validateSettingsSetResult(&api.ActionResult{Result: "saved"}); err != nil {
		t.Fatalf("saved result rejected: %v", err)
	}
	if err := validateSettingsSetResult(&api.ActionResult{Result: "failed"}); err == nil {
		t.Fatal("failed result must return an error")
	}
	if err := validateSettingsSetResult(nil); err == nil {
		t.Fatal("nil result must return an error")
	}
}

func TestBindListenerSetsRejectEmptyValues(t *testing.T) {
	attributes := settingsResourceSchema().Attributes
	for _, name := range []string{"listen_ipv4", "listen_ipv6"} {
		attribute, ok := attributes[name].(rschema.SetAttribute)
		if !ok {
			t.Fatalf("%s schema type = %T, want SetAttribute", name, attributes[name])
		}
		response := &frameworkvalidator.SetResponse{}
		request := frameworkvalidator.SetRequest{
			ConfigValue: types.SetValueMust(types.StringType, []attr.Value{}),
		}
		for _, configuredValidator := range attribute.Validators {
			configuredValidator.ValidateSet(context.Background(), request, response)
		}
		if !response.Diagnostics.HasError() {
			t.Fatalf("%s accepted an explicit empty listener set", name)
		}
	}

	forwarders, ok := attributes["forwarders"].(rschema.SetAttribute)
	if !ok {
		t.Fatalf("forwarders schema type = %T, want SetAttribute", attributes["forwarders"])
	}
	response := &frameworkvalidator.SetResponse{}
	request := frameworkvalidator.SetRequest{
		ConfigValue: types.SetValueMust(types.StringType, []attr.Value{}),
	}
	for _, configuredValidator := range forwarders.Validators {
		configuredValidator.ValidateSet(context.Background(), request, response)
	}
	if response.Diagnostics.HasError() {
		t.Fatalf("forwarders rejected an optional empty set: %v", response.Diagnostics)
	}
}

func TestPrimaryDomainUpdateKeysCanBeOwnedAdditively(t *testing.T) {
	attribute, ok := primaryDomainResourceSchema().Attributes["update_key_ids"].(rschema.SetAttribute)
	if !ok {
		t.Fatalf("update_key_ids schema type = %T, want SetAttribute", primaryDomainResourceSchema().Attributes["update_key_ids"])
	}
	if attribute.Default != nil {
		t.Fatal("update_key_ids must not force a default when additive attachment resources own memberships")
	}
	if !attribute.IsOptional() || !attribute.IsComputed() {
		t.Fatal("update_key_ids must remain Optional+Computed so omitted configuration preserves remote memberships")
	}
}

func TestPrimaryDomainUpdateKeyAttachmentPreservesOtherKeys(t *testing.T) {
	domain := &apibind.PrimaryDomain{UpdateKeys: api.SelectedMapList{"key-b", "key-a"}}
	addUpdateKey(domain, "key-c")
	if got := domain.UpdateKeys.String(); got != "key-a,key-b,key-c" {
		t.Fatalf("addUpdateKey() = %q, want sorted additive membership", got)
	}
	addUpdateKey(domain, "key-b")
	if got := domain.UpdateKeys.String(); got != "key-a,key-b,key-c" {
		t.Fatalf("duplicate add changed membership: %q", got)
	}
	removeUpdateKey(domain, "key-b")
	if got := domain.UpdateKeys.String(); got != "key-a,key-c" {
		t.Fatalf("removeUpdateKey() removed unrelated memberships: %q", got)
	}
}
