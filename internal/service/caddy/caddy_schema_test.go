package caddy

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/biptec/opnsense-go/pkg/api"
	apicaddy "github.com/biptec/opnsense-go/pkg/caddy"
	"github.com/biptec/opnsense-go/pkg/opnsense"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func stringSet(values ...string) types.Set {
	items := make([]attr.Value, 0, len(values))
	for _, value := range values {
		items = append(items, types.StringValue(value))
	}
	return types.SetValueMust(types.StringType, items)
}

func TestSettingsListenerAddressesRoundTrip(t *testing.T) {
	remote := &apicaddy.SettingsResponse{Caddy: apicaddy.Settings{General: apicaddy.GeneralSettings{
		HTTPPort: "80", HTTPSPort: "443", ListenAddresses: api.SelectedMapList{"10.0.0.2", "192.0.2.10"},
	}}}
	state, err := settingsStructToSchema(remote)
	if err != nil {
		t.Fatalf("settingsStructToSchema() error = %v", err)
	}
	if state.ListenAddresses.Elements()[0].String() == "" || len(state.ListenAddresses.Elements()) != 2 {
		t.Fatalf("unexpected listen addresses state: %#v", state.ListenAddresses)
	}

	model := &settingsResourceModel{
		ListenAddresses: stringSet("192.0.2.10", "10.0.0.2"),
		HTTPVersions:    stringSet("h1", "h2"),
	}
	general := &apicaddy.GeneralSettings{}
	applySettingsModel(general, model)
	if general.ListenAddresses.String() != "10.0.0.2,192.0.2.10" {
		t.Fatalf("unexpected API listen addresses: %q", general.ListenAddresses.String())
	}
}

func TestAccessListRoundTrip(t *testing.T) {
	model := &accessListResourceModel{
		Name:                types.StringValue("management"),
		ClientIPs:           stringSet("10.0.0.0/24", "10.1.0.0/24"),
		Invert:              types.BoolValue(false),
		HTTPResponseCode:    types.Int64Value(403),
		HTTPResponseMessage: types.StringValue("denied"),
		RequestMatcher:      types.StringValue("client_ip"),
		Description:         types.StringValue("management networks"),
	}
	remote, err := convertAccessListSchemaToStruct(model)
	if err != nil {
		t.Fatalf("convertAccessListSchemaToStruct() error = %v", err)
	}
	if remote.ClientIPs.String() != "10.0.0.0/24,10.1.0.0/24" || remote.HTTPResponseCode != "403" {
		t.Fatalf("unexpected API model: %+v", remote)
	}
	state, err := convertAccessListStructToSchema(remote)
	if err != nil {
		t.Fatalf("convertAccessListStructToSchema() error = %v", err)
	}
	if state.HTTPResponseCode.ValueInt64() != 403 || state.ClientIPs.Elements()[0].String() == "" {
		t.Fatalf("unexpected state model: %+v", state)
	}
}

func TestHeaderPatternsRejectCaddyfileInjection(t *testing.T) {
	nameTests := map[string]bool{
		"Host":       true,
		"+X-Trace":   true,
		"-Server-*":  true,
		"Host Other": false,
		`Host"Other`: false,
		"Host\nX":    false,
	}
	for value, want := range nameTests {
		if got := headerNamePattern.MatchString(value); got != want {
			t.Errorf("headerNamePattern.MatchString(%q) = %t, want %t", value, got, want)
		}
	}

	valueTests := map[string]bool{
		"{http.request.host}":           true,
		"value with spaces\tand tab":    true,
		`value"quoted`:                  false,
		"value\nheader_down X injected": false,
		"value\rheader_down X injected": false,
		"value\x00injected":             false,
		"value\x0binjected":             false,
		"value\x0cinjected":             false,
	}
	for value, want := range valueTests {
		if got := headerValuePattern.MatchString(value); got != want {
			t.Errorf("headerValuePattern.MatchString(%q) = %t, want %t", value, got, want)
		}
	}
}

func TestHeaderRoundTrip(t *testing.T) {
	model := &headerResourceModel{
		Direction:   types.StringValue("header_up"),
		Name:        types.StringValue("Host"),
		Value:       types.StringValue("{host}"),
		Replace:     types.StringValue(""),
		Description: types.StringValue("preserve frontend host"),
	}
	remote, err := convertHeaderSchemaToStruct(model)
	if err != nil {
		t.Fatalf("convertHeaderSchemaToStruct() error = %v", err)
	}
	if remote.Direction.String() != "header_up" || remote.Name != "Host" || remote.Value != "{host}" {
		t.Fatalf("unexpected API model: %+v", remote)
	}
	state, err := convertHeaderStructToSchema(remote)
	if err != nil {
		t.Fatalf("convertHeaderStructToSchema() error = %v", err)
	}
	if state.Direction.ValueString() != "header_up" || state.Name.ValueString() != "Host" || state.Value.ValueString() != "{host}" {
		t.Fatalf("unexpected state model: %+v", state)
	}
}

func TestHandlerProtocolRoundTrip(t *testing.T) {
	model := &handlerResourceModel{
		Enabled: types.BoolValue(true), DomainID: types.StringValue("domain-id"), SubdomainID: types.StringValue(""),
		HandlerType: types.StringValue("handle"), Path: types.StringValue(""), AccessListID: types.StringValue(""),
		BasicAuthIDs: stringSet(), HeaderIDs: stringSet(), Directive: types.StringValue("reverse_proxy"),
		UpstreamDomains: stringSet("10.0.0.10", "10.0.0.11"), UpstreamPort: types.Int64Value(8443),
		UpstreamPath: types.StringValue(""), ForwardAuth: types.BoolValue(false), UpstreamProtocol: types.StringValue("https"),
		HTTPVersion: types.StringValue("http2"), HTTPKeepalive: types.Int64Value(-1), TLSSkipVerify: types.BoolValue(false),
		TLSTrustCARefID: types.StringValue("ca-ref"), TLSServerName: types.StringValue("backend.internal"),
		LoadBalancingPolicy: types.StringValue("round_robin"), LoadBalancingRetries: types.Int64Value(-1),
		LoadBalancingTryDuration: types.Int64Value(-1), LoadBalancingTryInterval: types.Int64Value(-1),
		PassiveHealthFailDuration: types.Int64Value(-1), PassiveHealthMaxFails: types.Int64Value(-1),
		PassiveHealthUnhealthyStatus: types.StringValue(""), PassiveHealthUnhealthyLatency: types.Int64Value(-1),
		PassiveHealthUnhealthyRequestCount: types.Int64Value(-1), HealthURI: types.StringValue("/health"),
		HealthUpstream: types.StringValue(""), HealthPort: types.Int64Value(-1), HealthInterval: types.Int64Value(-1),
		HealthPasses: types.Int64Value(-1), HealthFails: types.Int64Value(-1), HealthTimeout: types.Int64Value(-1),
		HealthStatus: types.StringValue("200"), HealthBody: types.StringValue(""), HealthFollowRedirects: types.BoolValue(false),
		Description: types.StringValue("application"),
	}
	remote, err := convertHandlerSchemaToStruct(model)
	if err != nil {
		t.Fatalf("convertHandlerSchemaToStruct() error = %v", err)
	}
	if remote.UpstreamProtocol.String() != "1" || remote.TLSTrustPool.String() != "ca-ref" {
		t.Fatalf("unexpected API enum mapping: %+v", remote)
	}
	state, err := convertHandlerStructToSchema(remote)
	if err != nil {
		t.Fatalf("convertHandlerStructToSchema() error = %v", err)
	}
	if state.UpstreamProtocol.ValueString() != "https" || state.TLSServerName.ValueString() != "backend.internal" {
		t.Fatalf("unexpected state model: %+v", state)
	}
}

func baseDomainModel() *domainResourceModel {
	return &domainResourceModel{
		Enabled: types.BoolValue(true), Domain: types.StringValue("app.example.test"), Port: types.Int64Value(-1),
		Protocol: types.StringValue("https"), CertificateMode: types.StringValue("acme"), CertificateRefID: types.StringValue(""),
		InternalCAName: types.StringValue(""), InternalCertificateLifetimeDays: types.Int64Value(3650),
		InternalCertificateKeyType: types.StringValue("4096"), InternalCertificateDigest: types.StringValue("sha256"),
		GeneratedCertificateID: types.StringNull(), AccessListID: types.StringValue(""), BasicAuthIDs: stringSet(),
		Description: types.StringValue("application"), DNSChallenge: types.BoolValue(false),
		DNSChallengeOverrideDomain: types.StringValue(""), AccessLog: types.BoolValue(false), DynamicDNS: types.BoolValue(false),
		ACMEPassthrough: types.BoolValue(false), ACMEPassthroughUpstream: types.StringValue(""),
		ClientAuthMode: types.StringValue(""), ClientAuthCARefIDs: stringSet(),
	}
}

func TestDomainACMEPassthroughUsesExplicitUpstream(t *testing.T) {
	model := baseDomainModel()
	remote, err := buildDomainAPI(model, "")
	if err != nil {
		t.Fatalf("buildDomainAPI() default error = %v", err)
	}
	if remote.ACMEPassthrough != "" {
		t.Fatalf("default passthrough = %q, want empty", remote.ACMEPassthrough)
	}

	model.ACMEPassthrough = types.BoolValue(true)
	if _, err = buildDomainAPI(model, ""); err == nil {
		t.Fatal("legacy boolean passthrough should require an explicit upstream")
	}

	model.ACMEPassthroughUpstream = types.StringValue("acme-backend.internal")
	remote, err = buildDomainAPI(model, "")
	if err != nil {
		t.Fatalf("buildDomainAPI() upstream error = %v", err)
	}
	if remote.ACMEPassthrough != "acme-backend.internal" {
		t.Fatalf("passthrough upstream = %q", remote.ACMEPassthrough)
	}

	state, err := domainStructToSchema(&apicaddy.Domain{ACMEPassthrough: "0", DisableTLS: api.SelectedMap("1")}, nil)
	if err != nil {
		t.Fatalf("domainStructToSchema() error = %v", err)
	}
	if state.ACMEPassthrough.ValueBool() || state.ACMEPassthroughUpstream.ValueString() != "" {
		t.Fatalf("legacy API value was not normalized: %#v", state)
	}
}

func TestDomainCertificateModes(t *testing.T) {
	model := baseDomainModel()
	remote, err := buildDomainAPI(model, "")
	if err != nil || remote.DisableTLS.String() != "0" || remote.CustomCertificate.String() != "" {
		t.Fatalf("ACME domain = %+v, %v", remote, err)
	}

	model.CertificateMode = types.StringValue("custom")
	model.CertificateRefID = types.StringValue("cert-ref")
	remote, err = buildDomainAPI(model, model.CertificateRefID.ValueString())
	if err != nil || remote.CustomCertificate.String() != "cert-ref" {
		t.Fatalf("custom domain = %+v, %v", remote, err)
	}

	model.Protocol = types.StringValue("http")
	model.CertificateMode = types.StringValue("none")
	model.CertificateRefID = types.StringValue("")
	remote, err = buildDomainAPI(model, "")
	if err != nil || remote.DisableTLS.String() != "1" {
		t.Fatalf("HTTP domain = %+v, %v", remote, err)
	}

	model.CertificateMode = types.StringValue("acme")
	if err := validateDomainModel(model); err == nil {
		t.Fatal("expected protocol/certificate mode validation error")
	}
}

func TestInternalCertificateUsesExistingCA(t *testing.T) {
	var certBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch req.URL.Path {
		case "/api/trust/ca/search":
			_ = json.NewEncoder(w).Encode(map[string]any{"rows": []map[string]any{{
				"uuid": "ca-uuid", "refid": "ca-ref", "commonname": "example.test", "descr": "internal CA",
			}}, "total": 1, "rowCount": 1, "current": 1})
		case "/api/trust/ca/get/ca-uuid":
			_ = json.NewEncoder(w).Encode(map[string]any{"ca": map[string]any{
				"uuid": "ca-uuid", "refid": "ca-ref", "commonname": "example.test", "descr": "internal CA",
				"country": map[string]any{"NL": map[string]any{"selected": 1, "value": "Netherlands"}},
			}})
		case "/api/trust/cert/add":
			if err := json.NewDecoder(req.Body).Decode(&certBody); err != nil {
				t.Fatalf("decode certificate body: %v", err)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": "saved", "uuid": "cert-uuid"})
		case "/api/trust/cert/get/cert-uuid":
			_ = json.NewEncoder(w).Encode(map[string]any{"cert": map[string]any{
				"refid": "cert-ref", "commonname": "app.example.test", "caref": map[string]any{"ca-ref": map[string]any{"selected": 1, "value": "example.test"}},
			}})
		default:
			http.NotFound(w, req)
		}
	}))
	defer server.Close()

	client := api.NewClient(api.Options{Uri: server.URL})
	resource := &domainResource{resourceClient: resourceClient{client: opnsense.NewClient(client)}}
	model := baseDomainModel()
	model.CertificateMode = types.StringValue("internal")
	model.InternalCAName = types.StringValue("example.test")
	id, ref, err := resource.createInternalCertificate(context.Background(), model)
	if err != nil {
		t.Fatalf("createInternalCertificate() error = %v", err)
	}
	if id != "cert-uuid" || ref != "cert-ref" {
		t.Fatalf("createInternalCertificate() = %q, %q", id, ref)
	}
	cert := certBody["cert"].(map[string]any)
	if cert["caref"] != "ca-ref" || cert["country"] != "NL" || cert["lifetime"] != "3650" || cert["altnames_dns"] != "app.example.test" {
		t.Fatalf("unexpected certificate request: %#v", cert)
	}
}
