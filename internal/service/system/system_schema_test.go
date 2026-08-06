package system

import (
	"context"
	"errors"
	"testing"

	apiextensions "github.com/biptec/opnsense-go/pkg/api_extensions"
	apicore "github.com/biptec/opnsense-go/pkg/core"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestFirmwareHelpers(t *testing.T) {
	truthy := []string{"1", "true", "YES", "enabled", "locked"}
	for _, value := range truthy {
		if !firmwareFlag(value) {
			t.Errorf("firmwareFlag(%q) = false, want true", value)
		}
	}
	for _, value := range []string{"", "0", "false", "N/A", "unlocked"} {
		if firmwareFlag(value) {
			t.Errorf("firmwareFlag(%q) = true, want false", value)
		}
	}
	if got := normalizeFirmwareValue("N/A"); got != "" {
		t.Fatalf("normalizeFirmwareValue(N/A) = %q, want empty", got)
	}
}

func TestValidateFirmwareActionResult(t *testing.T) {
	if err := validateFirmwareActionResult("install", nil); err == nil {
		t.Fatal("nil firmware action result was accepted")
	}
	if err := validateFirmwareActionResult("remove", &apicore.FirmwareActionResult{Status: "failed"}); err == nil {
		t.Fatal("failed firmware action result was accepted")
	}
	if err := validateFirmwareActionResult("lock", &apicore.FirmwareActionResult{Status: "OK"}); err != nil {
		t.Fatalf("successful firmware action result rejected: %v", err)
	}
}

func TestFirmwareStatusDescription(t *testing.T) {
	if got := firmwareStatusDescription(nil, errors.New("status endpoint failed")); got != "unavailable (status endpoint failed)" {
		t.Fatalf("unexpected error status description: %q", got)
	}
	if got := firmwareStatusDescription(nil, nil); got != "unknown" {
		t.Fatalf("unexpected nil status description: %q", got)
	}
	if got := firmwareStatusDescription(&apicore.FirmwareUpgradeStatusResponse{Status: "ready"}, nil); got != `"ready"` {
		t.Fatalf("unexpected ready status description: %q", got)
	}
}

func TestWebguiRoundTripAndSafety(t *testing.T) {
	timeout := int64(600)
	model := &webguiResourceModel{
		Protocol:            types.StringValue("https"),
		Port:                types.Int64Value(443),
		Interfaces:          stringSetValue([]string{"lan"}),
		CertificateRef:      types.StringValue("cert-ref"),
		SessionTimeout:      types.Int64Value(timeout),
		HSTS:                types.BoolValue(true),
		DisableHTTPRedirect: types.BoolValue(false),
		AlternateHostnames:  stringSetValue([]string{"router.internal"}),
		AllowReaddress:      types.BoolValue(true),
	}
	remote, err := webguiToAPI(context.Background(), model)
	if err != nil {
		t.Fatalf("webguiToAPI() error = %v", err)
	}
	if remote.SessionTimeout == nil || *remote.SessionTimeout != int(timeout) {
		t.Fatalf("unexpected session timeout: %#v", remote.SessionTimeout)
	}
	state := webguiFromAPI(remote, true)
	if state.Protocol.ValueString() != "https" || state.Interfaces.Elements()[0].String() == "" {
		t.Fatalf("unexpected Web GUI state: %#v", state)
	}
	if !webguiEqual(remote, remote) {
		t.Fatal("identical Web GUI settings are not equal")
	}
	changed := *remote
	changed.Interfaces = []string{"opt1"}
	if !webguiDisruptiveChange(remote, &changed) {
		t.Fatal("listener interface change was not classified as disruptive")
	}
}

func TestSshRoundTripAndSafety(t *testing.T) {
	model := &sshResourceModel{
		Enabled:                types.BoolValue(true),
		Port:                   types.Int64Value(22),
		Interfaces:             stringSetValue([]string{"lan"}),
		PasswordAuthentication: types.BoolValue(false),
		PermitRootLogin:        types.BoolValue(false),
		AllowReaddress:         types.BoolValue(true),
	}
	remote, err := sshToAPI(context.Background(), model)
	if err != nil {
		t.Fatalf("sshToAPI() error = %v", err)
	}
	state := sshFromAPI(remote, true)
	if !state.Enabled.ValueBool() || state.Port.ValueInt64() != 22 {
		t.Fatalf("unexpected SSH state: %#v", state)
	}
	changed := *remote
	changed.Port = 2222
	if !sshDisruptiveChange(remote, &changed) {
		t.Fatal("SSH port change was not classified as disruptive")
	}
}

func TestNtpRoundTrip(t *testing.T) {
	serverSet, diagnostics := types.SetValueFrom(
		context.Background(),
		types.ObjectType{AttrTypes: ntpServerAttributeTypes},
		[]ntpServerModel{{
			Host:     types.StringValue("time.example.net"),
			NoSelect: types.BoolValue(false),
			Prefer:   types.BoolValue(true),
			IBurst:   types.BoolValue(true),
			Pool:     types.BoolValue(false),
		}},
	)
	if diagnostics.HasError() {
		t.Fatalf("create NTP server set: %v", diagnostics.Errors())
	}
	model := &ntpSettingsResourceModel{
		Enabled:              types.BoolValue(true),
		Servers:              serverSet,
		Interfaces:           stringSetValue([]string{"opt1"}),
		Orphan:               types.Int64Value(12),
		MaxClock:             types.Int64Value(10),
		ClientMode:           types.BoolValue(false),
		KissOfDeath:          types.BoolValue(true),
		RateLimiting:         types.BoolValue(true),
		DenyModifications:    types.BoolValue(true),
		DisableQueries:       types.BoolValue(true),
		DisableServing:       types.BoolValue(false),
		DenyPeerAssociations: types.BoolValue(true),
		DenyTrapService:      types.BoolValue(true),
	}
	remote, err := ntpToAPI(context.Background(), model)
	if err != nil {
		t.Fatalf("ntpToAPI() error = %v", err)
	}
	if len(remote.Servers) != 1 || remote.Servers[0].Host != "time.example.net" {
		t.Fatalf("unexpected NTP API model: %#v", remote)
	}
	state, err := ntpFromAPI(context.Background(), remote)
	if err != nil {
		t.Fatalf("ntpFromAPI() error = %v", err)
	}
	if state.Interfaces.Elements()[0].String() == "" || !ntpEqual(remote, remote) {
		t.Fatalf("unexpected NTP state: %#v", state)
	}

	changed := *remote
	changed.Servers = []apiextensions.NtpServer{{Host: "other.example.net", IBurst: true}}
	if ntpEqual(remote, &changed) {
		t.Fatal("different NTP servers compare equal")
	}
}

func TestListenerSchemasRequireExplicitInterfaces(t *testing.T) {
	webgui := webguiResourceSchema().Attributes["interfaces"]
	ssh := sshResourceSchema().Attributes["interfaces"]
	ntp := ntpSettingsResourceSchema().Attributes["interfaces"]
	if !webgui.IsRequired() || !ssh.IsRequired() || !ntp.IsRequired() {
		t.Fatal("management and NTP listener interfaces must remain required")
	}
}
