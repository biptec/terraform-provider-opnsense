package unbound

import (
	"reflect"
	"testing"

	"github.com/biptec/opnsense-go/pkg/api"
	apiunbound "github.com/biptec/opnsense-go/pkg/unbound"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestServiceAPIToModel(t *testing.T) {
	enabled := serviceAPIToModel("1")
	if enabled.ID.ValueString() != "unbound_service" || !enabled.Enabled.ValueBool() {
		t.Fatalf("unexpected enabled model: %#v", enabled)
	}
	disabled := serviceAPIToModel("0")
	if disabled.Enabled.ValueBool() {
		t.Fatalf("unexpected disabled model: %#v", disabled)
	}
}

func TestApplyUnboundServiceModelPreservesSettings(t *testing.T) {
	settings := apiunbound.Settings{
		General: apiunbound.General{
			Enabled: "1", Port: "53", DNSSEC: "1",
			ActiveInterface: api.SelectedMapList{"lan", "wan"},
		},
		ACLs:       apiunbound.ACLs{DefaultAction: api.SelectedMap("deny")},
		Forwarding: apiunbound.Forwarding{Enabled: "1"},
	}
	before := settings
	applyUnboundServiceModel(&settings, &serviceResourceModel{Enabled: types.BoolValue(false)})
	if settings.General.Enabled != "0" {
		t.Fatalf("expected disabled flag, got %q", settings.General.Enabled)
	}
	settings.General.Enabled = before.General.Enabled
	if !reflect.DeepEqual(settings, before) {
		t.Fatalf("unmanaged settings changed:\nwant %#v\n got %#v", before, settings)
	}
}
func TestValidateUnboundServiceResults(t *testing.T) {
	if validateUnboundUpdateResult(&apiunbound.Result{Result: "saved"}) != nil {
		t.Fatal("saved update result should be accepted")
	}
	if validateUnboundUpdateResult(nil) == nil || validateUnboundUpdateResult(&apiunbound.Result{Result: "failed"}) == nil {
		t.Fatal("invalid update responses must be rejected")
	}
	if validateUnboundActionResult(&apiunbound.ActionResult{Status: "ok"}) != nil {
		t.Fatal("ok action status should be accepted")
	}
	if validateUnboundActionResult(nil) == nil || validateUnboundActionResult(&apiunbound.ActionResult{Status: "failed"}) == nil {
		t.Fatal("invalid action responses must be rejected")
	}
}
