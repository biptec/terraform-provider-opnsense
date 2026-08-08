package bind

import (
	"testing"

	"github.com/biptec/opnsense-go/pkg/api"
	apibind "github.com/biptec/opnsense-go/pkg/bind"
	"github.com/biptec/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func transferModel(key string, notify ...string) *primaryDomainTransferResourceModel {
	return &primaryDomainTransferResourceModel{
		TransferKeyID: types.StringValue(key),
		AlsoNotify:    tools.StringSliceToSet(notify),
	}
}

func TestPrimaryDomainTransferHelpers(t *testing.T) {
	domain := &apibind.PrimaryDomain{}
	if !transferEmpty(domain) {
		t.Fatal("new domain transfer must be empty")
	}

	model := transferModel("11111111-1111-4111-8111-111111111111", "203.0.113.2", "10.0.0.2")
	applyTransfer(domain, model)

	if transferEmpty(domain) {
		t.Fatal("applied transfer must not be empty")
	}
	if !transferMatches(domain, model) {
		t.Fatal("applied transfer must match model regardless of notify order")
	}

	changed := transferModel("22222222-2222-4222-8222-222222222222", "10.0.0.2")
	if transferMatches(domain, changed) {
		t.Fatal("different owner configuration must not match")
	}

	clearTransfer(domain)
	if !transferEmpty(domain) {
		t.Fatalf("cleared transfer must be empty: key=%q notify=%v", domain.TransferKey.String(), []string(domain.AlsoNotify))
	}
}

func TestRemoteTransferValuesAreCanonical(t *testing.T) {
	domain := &apibind.PrimaryDomain{
		TransferKey: api.SelectedMap(" 11111111-1111-4111-8111-111111111111 "),
		AlsoNotify:  api.SelectedMapList{"", "203.0.113.2", " 10.0.0.2 ", "   "},
	}
	key, notify := remoteTransferValues(domain)
	if key != "11111111-1111-4111-8111-111111111111" {
		t.Fatalf("unexpected key %q", key)
	}
	if len(notify) != 2 || notify[0] != "10.0.0.2" || notify[1] != "203.0.113.2" {
		t.Fatalf("notify values are not canonical: %v", notify)
	}
}

func TestRemoteEmptyNotifyArtifactsAreIgnored(t *testing.T) {
	domain := &apibind.PrimaryDomain{AlsoNotify: api.SelectedMapList{"", "   "}}
	if !transferEmpty(domain) {
		key, notify := remoteTransferValues(domain)
		t.Fatalf("empty API artifacts must not create ownership: key=%q notify=%q", key, notify)
	}
}
