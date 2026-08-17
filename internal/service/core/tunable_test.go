package core

import (
	"testing"

	coreapi "github.com/biptec/opnsense-go/pkg/core"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestTunableRoundTrip(t *testing.T) {
	model := &tunableModel{
		Tunable:     types.StringValue("net.inet.carp.preempt"),
		Value:       types.StringValue("1"),
		Description: types.StringValue("Enable CARP preemption"),
	}
	remote := tunableToAPI(model)
	if remote.Tunable != "net.inet.carp.preempt" || remote.Value != "1" {
		t.Fatalf("unexpected tunable API object: %#v", remote)
	}
	state := tunableFromAPI(&coreapi.Tunable{Tunable: remote.Tunable, Value: remote.Value, Description: remote.Description}, "11111111-2222-4333-8444-555555555555")
	if state.ID.ValueString() == "" || state.Tunable.ValueString() != remote.Tunable || state.Value.ValueString() != remote.Value {
		t.Fatalf("unexpected tunable state: %#v", state)
	}
}
