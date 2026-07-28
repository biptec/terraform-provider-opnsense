package interfaces

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestConvertVipCARP(t *testing.T) {
	t.Parallel()
	model := vipResourceModel{Mode: types.StringValue("carp"), Interface: types.StringValue("lan"), Network: types.StringValue("192.0.2.10/24"), Gateway: types.StringNull(), NoExpand: types.BoolValue(false), NoBind: types.BoolValue(false), Password: types.StringValue("secret"), VHID: types.Int64Value(10), AdvertisementBase: types.Int64Value(1), AdvertisementSkew: types.Int64Value(20), PeerIPv4: types.StringNull(), PeerIPv6: types.StringNull(), NoSync: types.BoolValue(false), Description: types.StringValue("HA VIP")}
	got, err := convertVipSchemaToStruct(&model)
	if err != nil {
		t.Fatalf("convertVipSchemaToStruct() error = %v", err)
	}
	if got.Mode.String() != "carp" || got.VHID != "10" || got.AdvertisementSkew != "20" {
		t.Fatalf("vip = %#v", got)
	}
}

func TestConvertVipRequiresCARPCredentials(t *testing.T) {
	t.Parallel()
	model := vipResourceModel{Mode: types.StringValue("carp"), Interface: types.StringValue("lan"), Network: types.StringValue("192.0.2.10/24"), Password: types.StringNull(), VHID: types.Int64Null(), AdvertisementBase: types.Int64Value(1), AdvertisementSkew: types.Int64Value(0)}
	if _, err := convertVipSchemaToStruct(&model); err == nil {
		t.Fatal("expected CARP validation error")
	}
}
