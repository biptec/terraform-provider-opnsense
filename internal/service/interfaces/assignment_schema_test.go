package interfaces

import (
	"strings"
	"testing"

	"github.com/biptec/opnsense-go/pkg/api"
	apiinterfaces "github.com/biptec/opnsense-go/pkg/interfaces"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestConvertAssignmentSchemaToStruct(t *testing.T) {
	t.Parallel()
	model := assignmentResourceModel{
		Description: types.StringValue("Transit"), Device: types.StringValue("vtnet2"),
		Locked: types.BoolValue(false), Enabled: types.BoolValue(true), BlockPrivate: types.BoolValue(true),
		BlockBogons: types.BoolValue(false), GatewayInterface: types.BoolValue(false), Promiscuous: types.BoolValue(false),
		SpoofMAC: types.StringNull(), MTU: types.Int64Value(1500), MSS: types.Int64Null(),
		IPv4: &assignmentIPv4Model{Mode: types.StringValue("static"), Address: types.StringValue("192.0.2.1"), Prefix: types.Int64Value(24), Gateway: types.StringNull(), DHCPHostname: types.StringNull(), AliasAddress: types.StringNull(), AliasPrefix: types.Int64Null(), RejectFrom: types.StringNull()},
		IPv6: &assignmentIPv6Model{Mode: types.StringValue("track6"), Address: types.StringNull(), Prefix: types.Int64Null(), Gateway: types.StringNull(), IAPDLength: types.Int64Null(), IAPDSendHint: types.BoolValue(false), PrefixOnly: types.BoolValue(false), UseIPv4Interface: types.BoolValue(false), VLANPriority: types.Int64Null(), TrackInterface: types.StringValue("wan"), TrackPrefixID: types.Int64Value(1), TrackAssociatedPD: types.Int64Value(0)},
	}
	got, err := convertAssignmentSchemaToStruct(&model)
	if err != nil {
		t.Fatalf("convertAssignmentSchemaToStruct() error = %v", err)
	}
	if got.Device.String() != "vtnet2" || got.IPv4Mode.String() != "static" || got.IPv4Address != "192.0.2.1" || got.IPv4Prefix != "24" || got.IPv6Mode.String() != "track6" || got.Track6Interface != "wan" {
		t.Fatalf("converted assignment = %#v", got)
	}
}

func TestConvertAssignmentSchemaRejectsInvalidModes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		ipv4 *assignmentIPv4Model
		want string
	}{
		{"static without address", &assignmentIPv4Model{Mode: types.StringValue("static"), Address: types.StringNull(), Prefix: types.Int64Value(24)}, "requires address and prefix"},
		{"dhcp with static address", &assignmentIPv4Model{Mode: types.StringValue("dhcp"), Address: types.StringValue("192.0.2.1"), Prefix: types.Int64Null()}, "only be set in static mode"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			model := assignmentResourceModel{Device: types.StringValue("vtnet2"), Locked: types.BoolValue(false), Enabled: types.BoolValue(true), BlockPrivate: types.BoolValue(false), BlockBogons: types.BoolValue(false), GatewayInterface: types.BoolValue(false), Promiscuous: types.BoolValue(false), IPv4: test.ipv4, IPv6: &assignmentIPv6Model{Mode: types.StringValue("none")}}
			_, err := convertAssignmentSchemaToStruct(&model)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want text %q", err, test.want)
			}
		})
	}
}

func TestAssignmentStructRoundTrip(t *testing.T) {
	t.Parallel()
	input := &apiinterfaces.Assignment{Identifier: "opt1", Description: "Transit", Device: api.SelectedMap("vtnet2"), Lock: "0", Enabled: "1", IPv4Mode: api.SelectedMap("dhcp"), DHCPHostname: "edge", IPv6Mode: api.SelectedMap("none")}
	model := convertAssignmentStructToResourceSchema(input, "opt1", types.BoolValue(false))
	if model.Name.ValueString() != "opt1" || model.Device.ValueString() != "vtnet2" || model.IPv4.Mode.ValueString() != "dhcp" || model.IPv4.DHCPHostname.ValueString() != "edge" {
		t.Fatalf("model = %#v", model)
	}
}

func TestAssignmentAddressingChanged(t *testing.T) {
	t.Parallel()
	base := assignmentResourceModel{Device: types.StringValue("vtnet1"), IPv4: &assignmentIPv4Model{Mode: types.StringValue("none"), Address: types.StringNull(), Prefix: types.Int64Null(), Gateway: types.StringNull()}, IPv6: &assignmentIPv6Model{Mode: types.StringValue("none"), Address: types.StringNull(), Prefix: types.Int64Null(), Gateway: types.StringNull(), TrackInterface: types.StringNull(), TrackPrefixID: types.Int64Null()}}
	same := base
	if assignmentAddressingChanged(&base, &same) {
		t.Fatal("identical addressing reported as changed")
	}
	changed := base
	changed.Device = types.StringValue("vtnet2")
	if !assignmentAddressingChanged(&base, &changed) {
		t.Fatal("device change was not detected")
	}
}
