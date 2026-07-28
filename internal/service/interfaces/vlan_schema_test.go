package interfaces

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestConvertVlanProtocol(t *testing.T) {
	t.Parallel()
	model := vlanResourceModel{Description: types.StringValue("provider"), Tag: types.Int64Value(200), Priority: types.Int64Value(3), Protocol: types.StringValue("802.1ad"), Parent: types.StringValue("vtnet0"), Device: types.StringValue("qinq0.200")}
	got, err := convertVlanSchemaToStruct(&model)
	if err != nil {
		t.Fatalf("convertVlanSchemaToStruct() error = %v", err)
	}
	if got.Protocol.String() != "802.1ad" || got.Tag != "200" {
		t.Fatalf("vlan = %#v", got)
	}
}
