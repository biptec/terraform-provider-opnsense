package firewall

import (
	"testing"

	"github.com/biptec/opnsense-go/pkg/api"
	apifirewall "github.com/biptec/opnsense-go/pkg/firewall"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestFilterLogBoolStringConversion(t *testing.T) {
	model := &filterResourceModel{
		Enabled: types.BoolValue(true), Sequence: types.Int64Value(1), NoXMLRPCSync: types.BoolValue(false),
		Filter:    &filterFilterBlock{Quick: types.BoolValue(true), Action: types.StringValue("pass"), AllowOptions: types.BoolValue(false), Direction: types.StringValue("in"), IPProtocol: types.StringValue("inet"), Protocol: types.StringValue("CARP"), Log: types.BoolValue(true), TCPFlags: types.SetNull(types.StringType), TCPFlagsOutOf: types.SetNull(types.StringType), Schedule: types.StringValue("")},
		Interface: &filterInterfaceBlock{Invert: types.BoolValue(false), Interface: types.SetNull(types.StringType)},
	}
	remote, err := convertFilterSchemaToStruct(model)
	if err != nil {
		t.Fatalf("convertFilterSchemaToStruct(): %v", err)
	}
	if remote.Log != api.BoolString("1") {
		t.Fatalf("remote log=%q", remote.Log)
	}

	read := &apifirewall.Filter{Log: api.BoolString("1")}
	state, err := convertFilterStructToSchema(read)
	if err != nil {
		t.Fatalf("convertFilterStructToSchema(): %v", err)
	}
	if state.Filter == nil || !state.Filter.Log.ValueBool() {
		t.Fatalf("unexpected log state: %#v", state.Filter)
	}
}
