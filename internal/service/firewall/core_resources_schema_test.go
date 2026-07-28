package firewall

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestGroupSchemaRoundTrip(t *testing.T) {
	model := &groupResourceModel{
		Name:        types.StringValue("LAB_GROUP"),
		Members:     types.SetValueMust(types.StringType, []attr.Value{types.StringValue("lan"), types.StringValue("opt1")}),
		NoGroup:     types.BoolValue(true),
		Sequence:    types.Int64Value(20),
		Description: types.StringValue("test group"),
	}
	apiModel, err := convertGroupSchemaToStruct(model)
	if err != nil {
		t.Fatalf("convertGroupSchemaToStruct() error = %v", err)
	}
	if apiModel.Name != "LAB_GROUP" || apiModel.Members.String() != "lan,opt1" || apiModel.NoGroup != "1" {
		t.Fatalf("unexpected API model: %+v", apiModel)
	}
	state, err := convertGroupStructToSchema(apiModel)
	if err != nil {
		t.Fatalf("convertGroupStructToSchema() error = %v", err)
	}
	if state.Name.ValueString() != "LAB_GROUP" || !state.NoGroup.ValueBool() || state.Sequence.ValueInt64() != 20 {
		t.Fatalf("unexpected state model: %+v", state)
	}
}

func TestNptSchemaRoundTrip(t *testing.T) {
	model := &nptResourceModel{
		Enabled:        types.BoolValue(true),
		Log:            types.BoolValue(true),
		Sequence:       types.Int64Value(100),
		Categories:     types.SetValueMust(types.StringType, []attr.Value{types.StringValue("cat-id")}),
		Description:    types.StringValue("test npt"),
		Interface:      types.StringValue("lan"),
		SourceNet:      types.StringValue("fd00:1::/48"),
		DestinationNet: types.StringValue("2001:db8:1::/48"),
		TrackInterface: types.StringNull(),
	}
	apiModel, err := convertNptSchemaToStruct(model)
	if err != nil {
		t.Fatalf("convertNptSchemaToStruct() error = %v", err)
	}
	if apiModel.Enabled != "1" || apiModel.Interface.String() != "lan" || apiModel.SourceNet != "fd00:1::/48" {
		t.Fatalf("unexpected API model: %+v", apiModel)
	}
	state, err := convertNptStructToSchema(apiModel)
	if err != nil {
		t.Fatalf("convertNptStructToSchema() error = %v", err)
	}
	if !state.Enabled.ValueBool() || state.Sequence.ValueInt64() != 100 || state.DestinationNet.ValueString() != "2001:db8:1::/48" {
		t.Fatalf("unexpected state model: %+v", state)
	}
}
