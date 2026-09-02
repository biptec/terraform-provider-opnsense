package system

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/biptec/opnsense-go/pkg/api"
	apiextensions "github.com/biptec/opnsense-go/pkg/api_extensions"
	apicore "github.com/biptec/opnsense-go/pkg/core"
	"github.com/biptec/opnsense-go/pkg/opnsense"
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

func TestPluginRefreshUsesOnlyLocalPackageAPI(t *testing.T) {
	var packageCalls atomic.Int32
	var firmwareInfoCalls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/api_extensions/package/get/os-api-extensions":
			packageCalls.Add(1)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status": "ok",
				"package": map[string]any{
					"name":       "os-api-extensions",
					"installed":  true,
					"provided":   false,
					"version":    "0.12",
					"locked":     false,
					"repository": "unknown-repository",
					"origin":     "opnsense/os-api-extensions",
				},
			})
		case "/api/core/firmware/info":
			firmwareInfoCalls.Add(1)
			http.Error(w, "firmware info must not be called during refresh", http.StatusInternalServerError)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	r := &pluginResource{client: opnsense.NewClient(api.NewClient(api.Options{Uri: server.URL, MaxRetries: -1}))}
	started := time.Now()
	state, err := r.pluginState(context.Background(), "os-api-extensions", false)
	elapsed := time.Since(started)
	if err != nil {
		t.Fatalf("pluginState() error = %v", err)
	}
	if elapsed > time.Second {
		t.Fatalf("local plugin refresh took %s, want under 1s in the local API test", elapsed)
	}
	if got := packageCalls.Load(); got != 1 {
		t.Fatalf("local package API calls = %d, want 1", got)
	}
	if got := firmwareInfoCalls.Load(); got != 0 {
		t.Fatalf("firmware info calls during refresh = %d, want 0", got)
	}
	if !state.Installed.ValueBool() || state.Version.ValueString() != "0.12" || state.Locked.ValueBool() {
		t.Fatalf("unexpected plugin state: %#v", state)
	}
}

func TestPluginCreateAdoptsInstalledPackageWithoutFirmwareInfo(t *testing.T) {
	var firmwareInfoCalls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/api_extensions/package/get/os-api-extensions":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status": "ok",
				"package": map[string]any{
					"name": "os-api-extensions", "installed": true, "provided": false,
					"version": "0.12", "locked": false, "repository": "unknown-repository", "origin": "opnsense/os-api-extensions",
				},
			})
		case "/api/core/firmware/info":
			firmwareInfoCalls.Add(1)
			http.Error(w, "firmware info must not be called for an already installed package", http.StatusInternalServerError)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	r := &pluginResource{client: opnsense.NewClient(api.NewClient(api.Options{Uri: server.URL, MaxRetries: -1}))}
	if err := r.ensurePluginInstalled(context.Background(), "os-api-extensions"); err != nil {
		t.Fatalf("ensurePluginInstalled() error = %v", err)
	}
	if got := firmwareInfoCalls.Load(); got != 0 {
		t.Fatalf("firmware info calls while adopting installed plugin = %d, want 0", got)
	}
}

func TestValidateAPIExtensionActions(t *testing.T) {
	if err := validateWebguiAction("settings update", nil); err == nil {
		t.Fatal("nil Web GUI action result was accepted")
	}
	if err := validateSSHAction("reconfigure", &apiextensions.SshActionResult{Status: "failed"}); err == nil {
		t.Fatal("failed SSH action result was accepted")
	}
	if err := validateNTPAction("reconfigure", &apiextensions.NtpActionResult{Status: "OK"}); err != nil {
		t.Fatalf("successful NTP action result rejected: %v", err)
	}
	validation := &apiextensions.WebguiActionResult{
		Status:      "failed",
		Validations: map[string]string{"interfaces": "required"},
	}
	if err := validateWebguiAction("settings update", validation); err == nil || !strings.Contains(err.Error(), "interfaces: required") {
		t.Fatalf("validation details missing from error: %v", err)
	}
}

func TestWebguiTimeoutSchemaUsesMinutes(t *testing.T) {
	attributes := webguiResourceSchema().Attributes
	if _, ok := attributes["session_timeout_minutes"]; !ok {
		t.Fatal("session_timeout_minutes attribute is missing")
	}
	if _, ok := attributes["session_timeout"]; ok {
		t.Fatal("ambiguous session_timeout attribute must not be exposed")
	}
}

func TestWebguiRoundTripAndSafety(t *testing.T) {
	timeout := int64(15)
	model := &webguiResourceModel{
		Protocol:              types.StringValue("https"),
		Port:                  types.Int64Value(443),
		Interfaces:            stringSetValue([]string{"lan"}),
		CertificateRef:        types.StringValue("cert-ref"),
		SessionTimeoutMinutes: types.Int64Value(timeout),
		HSTS:                  types.BoolValue(true),
		DisableHTTPRedirect:   types.BoolValue(false),
		AlternateHostnames:    stringSetValue([]string{"router.internal"}),
		AllowReaddress:        types.BoolValue(true),
	}
	remote, err := webguiToAPI(context.Background(), model, &apiextensions.WebguiSettings{Protocol: "http", Port: 80, CertificateRef: "old-cert"})
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

func TestWebguiOmittedSettingsPreserveCurrent(t *testing.T) {
	current := &apiextensions.WebguiSettings{Protocol: "https", Port: 8443, Interfaces: []string{"opt1"}, CertificateRef: "cert-current", HSTS: false, DisableHTTPRedirect: true, AlternateHostnames: []string{"router.example"}}
	model := &webguiResourceModel{Interfaces: stringSetValue([]string{"lan"}), Protocol: types.StringNull(), Port: types.Int64Null(), CertificateRef: types.StringNull(), SessionTimeoutMinutes: types.Int64Null(), HSTS: types.BoolNull(), DisableHTTPRedirect: types.BoolNull(), AlternateHostnames: types.SetNull(types.StringType)}
	desired, err := webguiToAPI(context.Background(), model, current)
	if err != nil {
		t.Fatalf("webguiToAPI() error = %v", err)
	}
	if desired.Protocol != current.Protocol || desired.Port != current.Port || desired.CertificateRef != current.CertificateRef || desired.HSTS != current.HSTS || desired.DisableHTTPRedirect != current.DisableHTTPRedirect || !sameStrings(desired.AlternateHostnames, current.AlternateHostnames) || !sameStrings(desired.Interfaces, []string{"lan"}) {
		t.Fatalf("omitted Web GUI settings were not preserved: %#v", desired)
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
	remote, err := sshToAPI(context.Background(), model, &apiextensions.SshSettings{Enabled: false, Port: 2222, PasswordAuthentication: true, PermitRootLogin: true})
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

func TestSshOmittedSettingsPreserveCurrent(t *testing.T) {
	current := &apiextensions.SshSettings{Enabled: true, Port: 2222, Interfaces: []string{"opt1"}, PasswordAuthentication: true, PermitRootLogin: true}
	model := &sshResourceModel{Interfaces: stringSetValue([]string{"lan"}), Enabled: types.BoolNull(), Port: types.Int64Null(), PasswordAuthentication: types.BoolNull(), PermitRootLogin: types.BoolNull()}
	desired, err := sshToAPI(context.Background(), model, current)
	if err != nil {
		t.Fatalf("sshToAPI() error = %v", err)
	}
	if desired.Enabled != current.Enabled || desired.Port != current.Port || desired.PasswordAuthentication != current.PasswordAuthentication || desired.PermitRootLogin != current.PermitRootLogin || !sameStrings(desired.Interfaces, []string{"lan"}) {
		t.Fatalf("omitted SSH settings were not preserved: %#v", desired)
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
