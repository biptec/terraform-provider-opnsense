package system

import (
	"errors"
	"strings"
	"testing"

	"github.com/biptec/opnsense-go/pkg/errs"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestExactHAProxyTLSIdentityReplacement(t *testing.T) {
	state := &haproxyTLSIdentityResourceModel{
		FrontendName: types.StringValue("arcturus-shared-https"),
		DNSName:      types.StringValue("Router.Example.Invalid."),
		PolicyID:     types.StringValue("endpoint"),
		ID:           types.StringValue("old-local-uuid"),
	}
	match := map[string]string{
		"uuid":          "new-local-uuid",
		"frontend_name": "arcturus-shared-https",
		"dns_name":      "router.example.invalid",
		"policy_id":     "endpoint",
	}
	got, err := exactHAProxyTLSIdentityReplacement(state, []map[string]string{match})
	if got != "new-local-uuid" || err != nil {
		t.Fatalf("expected exact identity reattach, got %q %v", got, err)
	}

	table := []struct {
		name     string
		rows     []map[string]string
		state    *haproxyTLSIdentityResourceModel
		want     string
		notFound bool
	}{
		{"missing", nil, state, "", true},
		{"ambiguous", []map[string]string{match, match}, state, "found 2", false},
		{"wrong frontend", []map[string]string{{"uuid": "foreign", "frontend_name": "other", "dns_name": "router.example.invalid"}}, state, "", true},
		{"wrong DNS", []map[string]string{{"uuid": "foreign", "frontend_name": "arcturus-shared-https", "dns_name": "id.example.invalid"}}, state, "", true},
		{"missing UUID", []map[string]string{{"frontend_name": "arcturus-shared-https", "dns_name": "router.example.invalid"}}, state, "no UUID", false},
		{"unknown DNS", []map[string]string{match}, &haproxyTLSIdentityResourceModel{FrontendName: types.StringValue("arcturus-shared-https"), DNSName: types.StringUnknown()}, "lacks stable", false},
	}
	for _, tc := range table {
		t.Run(tc.name, func(t *testing.T) {
			_, e := exactHAProxyTLSIdentityReplacement(tc.state, tc.rows)
			if tc.notFound {
				var notFound *errs.NotFoundError
				if !errors.As(e, &notFound) {
					t.Fatalf("expected remote-not-found, got %v", e)
				}
				return
			}
			if e == nil || !strings.Contains(e.Error(), tc.want) {
				t.Fatalf("wanted error containing %q, got %v", tc.want, e)
			}
		})
	}
}
