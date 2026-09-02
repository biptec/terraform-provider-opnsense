package system

import (
	"context"
	"testing"

	apiextensions "github.com/biptec/opnsense-go/pkg/api_extensions"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestDnsConversionPreservesServerOrder(t *testing.T) {
	ctx := context.Background()
	model := &dnsSettingsResourceModel{
		Servers: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("10.16.18.53"),
			types.StringValue("10.16.16.53"),
		}),
		AllowOverride:   types.BoolValue(false),
		UseLocalService: types.BoolValue(false),
	}
	got, err := dnsToAPI(ctx, model)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Servers) != 2 || got.Servers[0] != "10.16.18.53" || got.Servers[1] != "10.16.16.53" {
		t.Fatalf("unexpected server order: %#v", got.Servers)
	}

	state, err := dnsFromAPI(ctx, got)
	if err != nil {
		t.Fatal(err)
	}
	var servers []string
	if diagnostics := state.Servers.ElementsAs(ctx, &servers, false); diagnostics.HasError() {
		t.Fatal(diagnostics.Errors())
	}
	if len(servers) != 2 || servers[0] != "10.16.18.53" || servers[1] != "10.16.16.53" {
		t.Fatalf("state lost server order: %#v", servers)
	}
}

func TestDnsEqualIsOrderSensitive(t *testing.T) {
	left := &apiextensions.DnsSettings{
		Servers:         []string{"10.16.18.53", "10.16.16.53"},
		AllowOverride:   false,
		UseLocalService: false,
	}
	right := &apiextensions.DnsSettings{
		Servers:         []string{"10.16.18.53", "10.16.16.53"},
		AllowOverride:   false,
		UseLocalService: false,
	}
	if !dnsEqual(left, right) {
		t.Fatal("identical DNS settings should compare equal")
	}
	right.Servers = []string{"10.16.16.53", "10.16.18.53"}
	if dnsEqual(left, right) {
		t.Fatal("DNS server order must be significant")
	}
}
