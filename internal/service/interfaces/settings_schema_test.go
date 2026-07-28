package interfaces

import (
	"testing"

	"github.com/biptec/opnsense-go/pkg/api"
	apiinterfaces "github.com/biptec/opnsense-go/pkg/interfaces"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestInterfaceSettingsConversion(t *testing.T) {
	t.Parallel()
	model := &interfaceSettingsResourceModel{DisableChecksumOffloading: types.BoolValue(true), DisableSegmentationOffloading: types.BoolValue(true), DisableLargeReceiveOffloading: types.BoolValue(false), VLANHardwareFiltering: types.StringValue("2"), DisableIPv6: types.BoolValue(false), DHCP6NoRelease: types.BoolValue(true), DHCP6Debug: types.BoolValue(false), DHCP6DUID: types.StringValue("00:04:01:02:03:04"), DHCP6RATimeout: types.Int64Value(10)}
	apiModel := convertInterfaceSettingsSchemaToStruct(model)
	if apiModel.DisableChecksumOffloading != "1" || apiModel.DisableLargeReceiveOffloading != "0" || apiModel.VLANHardwareFiltering.String() != "2" || apiModel.DHCP6RATimeout != "10" {
		t.Fatalf("api model = %#v", apiModel)
	}
	back := convertInterfaceSettingsStructToSchema(&apiinterfaces.InterfaceSettings{DisableChecksumOffloading: "1", DisableSegmentationOffloading: "1", DisableLargeReceiveOffloading: "0", VLANHardwareFiltering: api.SelectedMap("2"), DisableIPv6: "0", DHCP6NoRelease: "1", DHCP6Debug: "0", DHCP6DUID: "00:04:01:02:03:04", DHCP6RATimeout: "10"})
	if back.Id.ValueString() != "interfaces_settings" || !back.DHCP6NoRelease.ValueBool() || back.DHCP6RATimeout.ValueInt64() != 10 {
		t.Fatalf("schema model = %#v", back)
	}
}
