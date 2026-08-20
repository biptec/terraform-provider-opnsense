package bind

import (
	"context"
	"strings"
	"testing"

	"github.com/biptec/opnsense-go/pkg/api"
	apibind "github.com/biptec/opnsense-go/pkg/bind"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestNormalizeBindSemanticNames(t *testing.T) {
	for _, value := range []string{"biptec.com", "biptec.com.", " BIPTEC.COM. "} {
		if got := normalizeBindDomainName(value); got != "biptec.com" {
			t.Fatalf("normalizeBindDomainName(%q) = %q", value, got)
		}
	}
	for _, value := range []string{"internal", " INTERNAL ", "Internal"} {
		if got := normalizeBindViewName(value); got != "internal" {
			t.Fatalf("normalizeBindViewName(%q) = %q", value, got)
		}
	}
}

func TestSelectBindViewByName(t *testing.T) {
	rows := []apibind.View{{UUID: "view-public", Name: "public"}, {UUID: "view-internal", Name: "Internal"}}
	resolved, err := selectBindViewByName(rows, " internal ")
	if err != nil || resolved.ID != "view-internal" || resolved.Name != "internal" {
		t.Fatalf("selectBindViewByName() = %+v, %v", resolved, err)
	}
	if _, err := selectBindViewByName(rows, "missing"); err == nil || !strings.Contains(err.Error(), "was not found") {
		t.Fatalf("missing view error = %v", err)
	}
	duplicate := append(rows, apibind.View{UUID: "view-duplicate", Name: "INTERNAL"})
	if _, err := selectBindViewByName(duplicate, "internal"); err == nil || !strings.Contains(err.Error(), "multiple BIND views") {
		t.Fatalf("ambiguous view error = %v", err)
	}
}

func TestSelectPrimaryDomainInView(t *testing.T) {
	rows := []apibind.PrimaryDomain{
		{UUID: "zone-public", DomainName: "BIPTEC.COM."},
		{UUID: "zone-internal", DomainName: "biptec.com"},
	}
	remote := map[string]*apibind.PrimaryDomain{
		"zone-public":   {View: api.SelectedMap("view-public"), DomainName: "biptec.com"},
		"zone-internal": {View: api.SelectedMap("view-internal"), DomainName: "biptec.com"},
	}
	get := func(_ context.Context, id string) (*apibind.PrimaryDomain, error) { return remote[id], nil }
	resolved, err := selectPrimaryDomainInView(context.Background(), rows, " BIPTEC.COM. ", "view-internal", "internal", get)
	if err != nil || resolved.ID != "zone-internal" {
		t.Fatalf("selectPrimaryDomainInView() = %+v, %v", resolved, err)
	}
	if _, err := selectPrimaryDomainInView(context.Background(), rows, "missing.test", "view-internal", "internal", get); err == nil || !strings.Contains(err.Error(), "was not found in view") {
		t.Fatalf("missing zone error = %v", err)
	}
	duplicateRows := append(rows, apibind.PrimaryDomain{UUID: "zone-internal-2", DomainName: "biptec.com."})
	remote["zone-internal-2"] = &apibind.PrimaryDomain{View: api.SelectedMap("view-internal"), DomainName: "biptec.com"}
	if _, err := selectPrimaryDomainInView(context.Background(), duplicateRows, "biptec.com", "view-internal", "internal", get); err == nil || !strings.Contains(err.Error(), "multiple BIND primary domains") {
		t.Fatalf("ambiguous zone error = %v", err)
	}
}

func TestPrimaryDomainLookupSelectorModes(t *testing.T) {
	null := types.StringNull()
	value := func(s string) types.String { return types.StringValue(s) }
	tests := []struct {
		name                         string
		id, domain, viewName, viewID types.String
		valid                        bool
	}{
		{"id", value("id"), null, null, null, true},
		{"domain+view-name", null, value("example.test"), value("internal"), null, true},
		{"domain+view-id", null, value("example.test"), null, value("view-id"), true},
		{"domain-only", null, value("example.test"), null, null, false},
		{"view-name-only", null, null, value("internal"), null, false},
		{"id+domain", value("id"), value("example.test"), null, null, false},
		{"both-view-selectors", null, value("example.test"), value("internal"), value("view-id"), false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := validPrimaryDomainLookupSelectors(test.id, test.domain, test.viewName, test.viewID); got != test.valid {
				t.Fatalf("validPrimaryDomainLookupSelectors() = %t, want %t", got, test.valid)
			}
		})
	}
}

func TestSemanticLookupDataSourceSchemas(t *testing.T) {
	viewSchema := viewDataSourceSchema()
	if !viewSchema.Attributes["id"].IsOptional() || !viewSchema.Attributes["id"].IsComputed() ||
		!viewSchema.Attributes["name"].IsOptional() || !viewSchema.Attributes["name"].IsComputed() {
		t.Fatal("BIND view id/name selectors must be Optional+Computed")
	}
	zoneSchema := primaryDomainDataSourceSchema()
	for _, name := range []string{"id", "domain_name", "view_name", "view_id"} {
		attribute := zoneSchema.Attributes[name]
		if !attribute.IsOptional() || !attribute.IsComputed() {
			t.Fatalf("BIND primary-domain %s must be Optional+Computed", name)
		}
	}
}
