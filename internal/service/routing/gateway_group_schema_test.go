package routing

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestGatewayGroupSchemaRoundTrip(t *testing.T) {
	model := &gatewayGroupResourceModel{
		Name:        types.StringValue("GROUP_TEST"),
		Tier1:       types.SetValueMust(types.StringType, []attr.Value{types.StringValue("GW_A"), types.StringValue("GW_B")}),
		Tier2:       types.SetValueMust(types.StringType, []attr.Value{types.StringValue("GW_C")}),
		Tier3:       types.SetValueMust(types.StringType, nil),
		Tier4:       types.SetValueMust(types.StringType, nil),
		Tier5:       types.SetValueMust(types.StringType, nil),
		Trigger:     types.StringValue("downloss"),
		PoolOptions: types.StringValue("round-robin"),
		Description: types.StringValue("test group"),
	}

	apiModel, err := convertGatewayGroupSchemaToStruct(model)
	if err != nil {
		t.Fatalf("convertGatewayGroupSchemaToStruct() error = %v", err)
	}
	if apiModel.Name != "GROUP_TEST" || len(apiModel.Tier1) != 2 || apiModel.Trigger.String() != "downloss" {
		t.Fatalf("unexpected API model: %+v", apiModel)
	}

	state, err := convertGatewayGroupStructToSchema(apiModel)
	if err != nil {
		t.Fatalf("convertGatewayGroupStructToSchema() error = %v", err)
	}
	if state.Name.ValueString() != "GROUP_TEST" || state.Trigger.ValueString() != "downloss" {
		t.Fatalf("unexpected state model: %+v", state)
	}
	var tier1 []string
	state.Tier1.ElementsAs(t.Context(), &tier1, false)
	if len(tier1) != 2 {
		t.Fatalf("unexpected tier1 values: %#v", tier1)
	}
}
