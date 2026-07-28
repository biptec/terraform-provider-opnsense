package interfaces

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestConvertGreTunnel(t *testing.T) {
	t.Parallel()
	model := greResourceModel{LocalAddress: types.StringValue("wan"), RemoteAddress: types.StringValue("198.51.100.2"), TunnelLocalAddress: types.StringValue("10.0.0.1"), TunnelRemoteAddress: types.StringValue("10.0.0.2"), TunnelRemotePrefix: types.Int64Value(30), Description: types.StringValue("peer")}
	got, err := convertGreSchemaToStruct(&model)
	if err != nil {
		t.Fatalf("convertGreSchemaToStruct() error = %v", err)
	}
	if got.LocalAddress != "wan" || got.TunnelRemotePrefix != "30" {
		t.Fatalf("gre = %#v", got)
	}
}

func TestTunnelAddressValidation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		local  types.String
		remote types.String
		prefix types.Int64
		want   string
	}{
		{"missing peer", types.StringValue("10.0.0.1"), types.StringNull(), types.Int64Value(30), "configured together"},
		{"family mismatch", types.StringValue("10.0.0.1"), types.StringValue("2001:db8::2"), types.Int64Value(64), "same address family"},
		{"ipv4 prefix too long", types.StringValue("10.0.0.1"), types.StringValue("10.0.0.2"), types.Int64Value(64), "must not exceed 32"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateTunnelAddresses(types.StringValue("198.51.100.2"), test.local, test.remote, test.prefix)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestConvertGifFlags(t *testing.T) {
	t.Parallel()
	model := gifResourceModel{LocalAddress: types.StringValue("wan"), RemoteAddress: types.StringValue("2001:db8::2"), TunnelLocalAddress: types.StringValue("2001:db8:1::1"), TunnelRemoteAddress: types.StringValue("2001:db8:1::2"), TunnelRemotePrefix: types.Int64Value(64), ECNFriendly: types.BoolValue(true), DisableIngressFiltering: types.BoolValue(true), Description: types.StringNull()}
	got, err := convertGifSchemaToStruct(&model)
	if err != nil {
		t.Fatalf("convertGifSchemaToStruct() error = %v", err)
	}
	if got.ECNFriendly != "1" || got.IngressFiltering != "1" {
		t.Fatalf("gif = %#v", got)
	}
}

func TestConvertNeighbor(t *testing.T) {
	t.Parallel()
	model := neighborResourceModel{MACAddress: types.StringValue("AA:BB:CC:DD:EE:FF"), IPAddress: types.StringValue("192.0.2.20"), Description: types.StringValue("server")}
	got, err := convertNeighborSchemaToStruct(&model)
	if err != nil {
		t.Fatalf("convertNeighborSchemaToStruct() error = %v", err)
	}
	if got.MACAddress != "aa:bb:cc:dd:ee:ff" || got.IPAddress != "192.0.2.20" {
		t.Fatalf("neighbor = %#v", got)
	}
	model.IPAddress = types.StringValue("invalid")
	if _, err := convertNeighborSchemaToStruct(&model); err == nil {
		t.Fatal("expected invalid IP error")
	}
	model.IPAddress = types.StringValue("192.0.2.20")
	model.MACAddress = types.StringValue("00:11:22:33:44:55:66:77")
	if _, err := convertNeighborSchemaToStruct(&model); err == nil {
		t.Fatal("expected EUI-64 rejection")
	}
}

func TestConvertLoopback(t *testing.T) {
	t.Parallel()
	model := loopbackResourceModel{Description: types.StringValue("routing ID")}
	got, err := convertLoopbackSchemaToStruct(&model)
	if err != nil {
		t.Fatalf("convertLoopbackSchemaToStruct() error = %v", err)
	}
	if got.Description != "routing ID" {
		t.Fatalf("loopback = %#v", got)
	}
}
