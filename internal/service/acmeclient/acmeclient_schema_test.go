package acmeclient

import "testing"

func TestRFC2136KeyIsWriteOnly(t *testing.T) {
	schema := validationResourceSchema()
	attr, ok := schema.Attributes["dns_nsupdate_key"]
	if !ok {
		t.Fatal("dns_nsupdate_key attribute missing")
	}
	if !attr.IsSensitive() || !attr.IsWriteOnly() || attr.IsComputed() {
		t.Fatal("dns_nsupdate_key must be sensitive, write-only, and non-computed")
	}
}

func TestACMEModelsDoNotPersistPrivateKeys(t *testing.T) {
	account := accountResourceSchema()
	for _, forbidden := range []string{"key", "private_key", "account_key"} {
		if _, ok := account.Attributes[forbidden]; ok {
			t.Fatalf("account schema must not expose %q", forbidden)
		}
	}
	cert := certificateResourceSchema()
	for _, forbidden := range []string{"private_key", "certificate_key", "prv", "prv_payload"} {
		if _, ok := cert.Attributes[forbidden]; ok {
			t.Fatalf("certificate schema must not expose %q", forbidden)
		}
	}
}

func TestStagingAndChallengeAliasAreExplicit(t *testing.T) {
	settings := settingsResourceSchema()
	if _, ok := settings.Attributes["environment"]; !ok {
		t.Fatal("environment attribute missing")
	}
	cert := certificateResourceSchema()
	if _, ok := cert.Attributes["challenge_alias"]; !ok {
		t.Fatal("challenge_alias attribute missing")
	}
	if _, ok := cert.Attributes["issuance_version"]; !ok {
		t.Fatal("issuance_version attribute missing")
	}
}
