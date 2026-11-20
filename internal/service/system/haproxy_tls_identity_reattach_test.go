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
		{"wrong policy", []map[string]string{{"uuid": "foreign", "frontend_name": "arcturus-shared-https", "dns_name": "router.example.invalid", "policy_id": "local"}}, state, "different owner policy", false},
		{"missing UUID", []map[string]string{{"frontend_name": "arcturus-shared-https", "dns_name": "router.example.invalid", "policy_id": "endpoint"}}, state, "no UUID", false},
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

func TestExactHAProxyTLSIdentityCreateMatch(t *testing.T) {
	wanted := &haproxyTLSIdentityResourceModel{
		FrontendName: types.StringValue("arcturus-shared-https"),
		DNSName:      types.StringValue("s3.central.sandbox.biptec.net."),
		PolicyID:     types.StringValue("endpoint"),
	}
	match := map[string]string{
		"uuid":          "persisted-local-uuid",
		"frontend_name": "arcturus-shared-https",
		"dns_name":      "s3.central.sandbox.biptec.net",
		"policy_id":     "endpoint",
	}
	got, err := exactHAProxyTLSIdentityCreateMatch(wanted, []map[string]string{match})
	if got != "persisted-local-uuid" || err != nil {
		t.Fatalf("expected exact create reattach, got %q %v", got, err)
	}

	tests := []struct {
		name     string
		rows     []map[string]string
		want     string
		notFound bool
	}{
		{name: "not found", rows: nil, notFound: true},
		{name: "ambiguous", rows: []map[string]string{match, match}, want: "found 2"},
		{name: "wrong policy", rows: []map[string]string{{
			"uuid": "foreign", "frontend_name": "arcturus-shared-https",
			"dns_name": "s3.central.sandbox.biptec.net", "policy_id": "local",
		}}, want: "different owner policy"},
		{name: "missing UUID", rows: []map[string]string{{
			"frontend_name": "arcturus-shared-https", "dns_name": "s3.central.sandbox.biptec.net",
			"policy_id": "endpoint",
		}}, want: "no UUID"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := exactHAProxyTLSIdentityCreateMatch(wanted, test.rows)
			if test.notFound {
				var notFound *errs.NotFoundError
				if !errors.As(err, &notFound) {
					t.Fatalf("expected remote-not-found, got %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("wanted error containing %q, got %v", test.want, err)
			}
		})
	}
}

func TestValidateHAProxyTLSIdentitySearchResponse(t *testing.T) {
	totalZero := 0
	totalOne := 1
	totalTwo := 2
	tests := []struct {
		name    string
		result  *haproxyTLSIdentitySearchResponse
		wantErr string
	}{
		{name: "nil response", result: nil, wantErr: "empty search response"},
		{name: "search error message is not not-found", result: &haproxyTLSIdentitySearchResponse{Message: "unable to read HAProxy TLS identity configuration"}, wantErr: "unable to read HAProxy TLS identity configuration"},
		{name: "missing total is rejected", result: &haproxyTLSIdentitySearchResponse{Rows: &[]map[string]string{}}, wantErr: "missing total or rows"},
		{name: "missing rows is rejected", result: &haproxyTLSIdentitySearchResponse{Total: &totalZero}, wantErr: "missing total or rows"},
		{name: "truncated search is rejected", result: &haproxyTLSIdentitySearchResponse{Total: &totalTwo, Rows: &[]map[string]string{{"uuid": "one"}}}, wantErr: "incomplete search response"},
		{name: "empty successful search is not an error", result: &haproxyTLSIdentitySearchResponse{Total: &totalZero, Rows: &[]map[string]string{}}, wantErr: ""},
		{name: "complete search is accepted", result: &haproxyTLSIdentitySearchResponse{Total: &totalOne, Rows: &[]map[string]string{{"uuid": "one"}}}, wantErr: ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateHAProxyTLSIdentitySearchResponse(test.result)
			if test.wantErr == "" {
				if err != nil {
					t.Fatalf("expected valid search response, got %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("wanted error containing %q, got %v", test.wantErr, err)
			}
		})
	}
}
