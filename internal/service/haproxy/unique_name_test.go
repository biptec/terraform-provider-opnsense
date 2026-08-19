package haproxy

import (
	"strings"
	"testing"
)

func TestValidateUniqueRemoteName(t *testing.T) {
	t.Parallel()

	items := []namedRemoteItem{
		{ID: "uuid-a", Name: "nc-sni-a.example.invalid"},
		{ID: "uuid-b", Name: "nc-sni-b.example.invalid"},
	}

	if err := validateUniqueRemoteName("HAProxy backend", "nc-sni-new.example.invalid", "", items); err != nil {
		t.Fatalf("unused name rejected: %v", err)
	}
	if err := validateUniqueRemoteName("HAProxy backend", "nc-sni-a.example.invalid", "uuid-a", items); err != nil {
		t.Fatalf("resource rejected its own unchanged name: %v", err)
	}
	if err := validateUniqueRemoteName("HAProxy backend", "nc-sni-a.example.invalid", "uuid-b", items); err == nil {
		t.Fatal("duplicate remote name accepted")
	} else if !strings.Contains(err.Error(), `name "nc-sni-a.example.invalid" is already used`) {
		t.Fatalf("unexpected duplicate error: %v", err)
	}
}
