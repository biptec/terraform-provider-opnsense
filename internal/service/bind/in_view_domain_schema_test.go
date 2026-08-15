package bind

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBindInViewDomainRoundTrip(t *testing.T) {
	model := &inViewDomainResourceModel{
		ViewID:       types.StringValue("11111111-1111-4111-8111-111111111111"),
		DomainName:   types.StringValue("acme.biptec.net"),
		Enabled:      types.BoolValue(true),
		SourceViewID: types.StringValue("22222222-2222-4222-8222-222222222222"),
	}
	remote, err := inViewDomainModelToAPI(model)
	if err != nil {
		t.Fatal(err)
	}
	if remote.View.String() != model.ViewID.ValueString() || remote.SourceView.String() != model.SourceViewID.ValueString() || remote.DomainName != "acme.biptec.net" {
		t.Fatalf("unexpected API model: %+v", remote)
	}
	roundTrip, err := inViewDomainAPIToModel(remote)
	if err != nil {
		t.Fatal(err)
	}
	if roundTrip.ViewID.ValueString() != model.ViewID.ValueString() || roundTrip.SourceViewID.ValueString() != model.SourceViewID.ValueString() || !roundTrip.Enabled.ValueBool() {
		t.Fatalf("unexpected round-trip model: %+v", roundTrip)
	}
}
