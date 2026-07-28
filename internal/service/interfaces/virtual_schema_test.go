package interfaces

import (
	"strings"
	"testing"

	"github.com/biptec/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestConvertVxlanUnicast(t *testing.T) {
	t.Parallel()
	model := vxlanResourceModel{VNI: types.Int64Value(100), SourceAddress: types.StringValue("192.0.2.1"), SourcePort: types.Int64Value(4789), RemoteAddress: types.StringValue("192.0.2.2"), RemotePort: types.Int64Value(4789), MulticastGroup: types.StringNull(), MulticastInterface: types.StringNull()}
	got, err := convertVxlanSchemaToStruct(&model)
	if err != nil {
		t.Fatalf("convertVxlanSchemaToStruct() error = %v", err)
	}
	if got.VNI != "100" || got.LocalAddress != "192.0.2.1" || got.RemoteAddress != "192.0.2.2" {
		t.Fatalf("vxlan = %#v", got)
	}
}

func TestConvertVxlanMulticastValidation(t *testing.T) {
	t.Parallel()
	model := vxlanResourceModel{VNI: types.Int64Value(100), SourceAddress: types.StringValue("192.0.2.1"), SourcePort: types.Int64Value(4789), RemoteAddress: types.StringNull(), RemotePort: types.Int64Value(4789), MulticastGroup: types.StringValue("239.1.1.1"), MulticastInterface: types.StringNull()}
	_, err := convertVxlanSchemaToStruct(&model)
	if err == nil || !strings.Contains(err.Error(), "multicast_interface is required") {
		t.Fatalf("error = %v", err)
	}
	model.MulticastInterface = types.StringValue("vtnet0")
	if _, err := convertVxlanSchemaToStruct(&model); err != nil {
		t.Fatalf("valid multicast VXLAN rejected: %v", err)
	}
}

func TestConvertBridgeValidatesMemberOptions(t *testing.T) {
	t.Parallel()
	empty := tools.EmptySetValue(types.StringType)
	model := bridgeResourceModel{Members: tools.StringSliceToSet([]string{"lan", "opt1"}), LinkLocal: types.BoolValue(false), EnableSTP: types.BoolValue(true), Protocol: types.StringValue("rstp"), STPMembers: tools.StringSliceToSet([]string{"lan"}), Edge: empty, AutoEdge: empty, PointToPoint: empty, AutoPointToPoint: empty, Static: empty, Private: empty, Span: types.StringNull()}
	got, err := convertBridgeSchemaToStruct(&model)
	if err != nil {
		t.Fatalf("convertBridgeSchemaToStruct() error = %v", err)
	}
	if len(got.Members) != 2 || len(got.STPMembers) != 1 {
		t.Fatalf("bridge = %#v", got)
	}
	model.Edge = tools.StringSliceToSet([]string{"wan"})
	if _, err := convertBridgeSchemaToStruct(&model); err == nil {
		t.Fatal("expected non-member validation error")
	}
}

func TestConvertLaggValidation(t *testing.T) {
	t.Parallel()
	model := laggResourceModel{Members: tools.StringSliceToSet([]string{"vtnet1", "vtnet2"}), PrimaryMember: types.StringNull(), Protocol: types.StringValue("lacp"), LACPFastTimeout: types.BoolValue(true), UseFlowID: types.StringValue("1"), HashLayers: tools.StringSliceToSet([]string{"l2", "l3"}), LACPStrict: types.StringValue("1"), MTU: types.Int64Null(), Description: types.StringValue("uplink")}
	got, err := convertLaggSchemaToStruct(&model)
	if err != nil {
		t.Fatalf("convertLaggSchemaToStruct() error = %v", err)
	}
	if got.Protocol.String() != "lacp" || len(got.Members) != 2 {
		t.Fatalf("lagg = %#v", got)
	}
	model.Protocol = types.StringValue("failover")
	model.PrimaryMember = types.StringValue("vtnet9")
	model.LACPFastTimeout = types.BoolValue(false)
	model.UseFlowID = types.StringValue("")
	model.HashLayers = tools.EmptySetValue(types.StringType)
	model.LACPStrict = types.StringValue("")
	if _, err := convertLaggSchemaToStruct(&model); err == nil {
		t.Fatal("expected primary member validation error")
	}
}
