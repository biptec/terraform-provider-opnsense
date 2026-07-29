package routing

import (
	"testing"

	apirouting "github.com/biptec/opnsense-go/pkg/routing"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestGatewaySchemaRoundTrip(t *testing.T) {
	model := &gatewayResourceModel{
		Enabled:                   types.BoolValue(true),
		Name:                      types.StringValue("GW_TEST"),
		Description:               types.StringValue("test gateway"),
		Interface:                 types.StringValue("lan"),
		IPProtocol:                types.StringValue("inet"),
		Gateway:                   types.StringValue("192.0.2.1"),
		DefaultGateway:            types.BoolValue(false),
		FarGateway:                types.BoolValue(true),
		MonitorDisable:            types.BoolValue(false),
		MonitorNoRoute:            types.BoolValue(true),
		MonitorKillStates:         types.BoolValue(true),
		MonitorKillStatesPriority: types.BoolValue(false),
		Monitor:                   types.StringValue("192.0.2.2"),
		ForceDown:                 types.BoolValue(false),
		NoSync:                    types.BoolValue(true),
		Priority:                  types.Int64Value(200),
		Weight:                    types.Int64Value(3),
		LatencyLow:                types.Int64Value(100),
		LatencyHigh:               types.Int64Value(500),
		LossLow:                   types.Int64Value(10),
		LossHigh:                  types.Int64Value(20),
		Interval:                  types.Int64Value(1000),
		TimePeriod:                types.Int64Value(60000),
		LossInterval:              types.Int64Value(2000),
		DataLength:                types.Int64Value(1),
	}

	apiModel, err := convertGatewaySchemaToStruct(model)
	if err != nil {
		t.Fatalf("convertGatewaySchemaToStruct() error = %v", err)
	}
	if apiModel.Disabled != "0" || apiModel.Interface.String() != "lan" || apiModel.FarGateway != "1" {
		t.Fatalf("unexpected API model: %+v", apiModel)
	}
	if apiModel.Priority != "200" || apiModel.LatencyHigh != "500" || apiModel.DataLength != "1" {
		t.Fatalf("unexpected numeric conversion: %+v", apiModel)
	}

	state, err := convertGatewayStructToSchema(apiModel)
	if err != nil {
		t.Fatalf("convertGatewayStructToSchema() error = %v", err)
	}
	if !state.Enabled.ValueBool() || state.Name.ValueString() != "GW_TEST" || state.Interface.ValueString() != "lan" {
		t.Fatalf("unexpected state model: %+v", state)
	}
	if state.Priority.ValueInt64() != 200 || state.Weight.ValueInt64() != 3 || state.LossHigh.ValueInt64() != 20 {
		t.Fatalf("unexpected numeric state: %+v", state)
	}
}

func TestGatewayOptionalThresholdsRemainUnset(t *testing.T) {
	state, err := convertGatewayStructToSchema(&apirouting.Gateway{Priority: "255", Weight: "1"})
	if err != nil {
		t.Fatalf("convertGatewayStructToSchema() error = %v", err)
	}
	if !state.LatencyLow.IsNull() || !state.LossHigh.IsNull() || !state.Interval.IsNull() {
		t.Fatalf("optional thresholds should remain null: %+v", state)
	}
}
