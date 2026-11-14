package haproxy

import (
	"testing"

	"github.com/biptec/opnsense-go/pkg/api"
	apihaproxy "github.com/biptec/opnsense-go/pkg/haproxy"
	"github.com/biptec/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func stringSet(values ...string) types.Set {
	items := make([]attr.Value, 0, len(values))
	for _, value := range values {
		items = append(items, types.StringValue(value))
	}
	return types.SetValueMust(types.StringType, items)
}

func TestFrontendL4PassthroughConversion(t *testing.T) {
	t.Parallel()
	model := &frontendModel{
		Enabled: types.BoolValue(true), Name: types.StringValue("tls_ingress"),
		Bind: stringSet("10.0.0.10:443"), Mode: types.StringValue("tcp"),
		SSLEnabled: types.BoolValue(false), CustomOptions: types.StringValue("tcp-request inspect-delay 5s"),
		LinkedActions: stringSet("action-id"),
	}
	got, err := frontendModelToAPI(model)
	if err != nil {
		t.Fatal(err)
	}
	if got.Mode.String() != "tcp" || got.SSLEnabled != "0" || got.Bind.String() != "10.0.0.10:443" {
		t.Fatalf("frontend is not raw TCP passthrough: %+v", got)
	}
	if got.LinkedActions.String() != "action-id" || got.CustomOptions != "tcp-request inspect-delay 5s" {
		t.Fatalf("unexpected frontend routing fields: %+v", got)
	}
}

func TestFrontendClearsDefaultBackend(t *testing.T) {
	t.Parallel()
	got, err := frontendModelToAPI(&frontendModel{
		Enabled:        types.BoolValue(true),
		Name:           types.StringValue("shared_internal"),
		Bind:           stringSet("10.0.0.10:443"),
		Mode:           types.StringValue("http"),
		DefaultBackend: types.StringNull(),
		SSLEnabled:     types.BoolValue(true),
		LinkedActions:  stringSet(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.DefaultBackend.String() != "" {
		t.Fatalf("default backend = %q, want empty API value", got.DefaultBackend.String())
	}
}

func TestFrontendTLSTerminationConversion(t *testing.T) {
	t.Parallel()
	model := &frontendModel{
		Enabled: types.BoolValue(true), Name: types.StringValue("oidc_ingress"), HAPolicy: types.StringValue("endpoint"),
		Bind: stringSet("10.0.16.81:443"), Mode: types.StringValue("http"),
		SSLEnabled: types.BoolValue(true), SSLCertificates: stringSet("oidc-cert"),
		SSLDefaultCertificate: types.StringValue("oidc-cert"), LinkedActions: stringSet(),
	}
	got, err := frontendModelToAPI(model)
	if err != nil {
		t.Fatal(err)
	}
	if got.HAPolicy.String() != "endpoint" || got.Mode.String() != "http" || got.SSLEnabled != "1" || got.SSLCertificates.String() != "oidc-cert" || got.SSLDefaultCertificate.String() != "oidc-cert" {
		t.Fatalf("frontend TLS termination fields not preserved: %+v", got)
	}
}

func TestFrontendSSLCertificatesMayBeChildOwned(t *testing.T) {
	t.Parallel()

	attr, ok := frontendResourceSchema().Attributes["ssl_certificates"].(resourceschema.SetAttribute)
	if !ok {
		t.Fatal("ssl_certificates is not a SetAttribute")
	}
	if !attr.Optional || !attr.Computed {
		t.Fatalf("ssl_certificates must be Optional+Computed, got Optional=%v Computed=%v", attr.Optional, attr.Computed)
	}
	if attr.Default != nil {
		t.Fatal("ssl_certificates must not have a default when child TLS identities own membership")
	}

	got, err := frontendModelToAPI(&frontendModel{
		Enabled:         types.BoolValue(true),
		Name:            types.StringValue("shared"),
		Bind:            stringSet("10.0.0.10:443"),
		Mode:            types.StringValue("http"),
		SSLEnabled:      types.BoolValue(true),
		SSLCertificates: types.SetNull(types.StringType),
		LinkedActions:   stringSet(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.SSLCertificates != nil {
		t.Fatalf("omitted ssl_certificates must stay omitted in API payload, got %#v", got.SSLCertificates)
	}
}

func TestFrontendCreateSeedsDefaultForChildOwnedMembership(t *testing.T) {
	t.Parallel()

	got, err := frontendModelToAPI(&frontendModel{
		Enabled:               types.BoolValue(true),
		Name:                  types.StringValue("shared"),
		Bind:                  stringSet("10.0.0.10:443"),
		Mode:                  types.StringValue("http"),
		SSLEnabled:            types.BoolValue(true),
		SSLCertificates:       types.SetNull(types.StringType),
		SSLDefaultCertificate: types.StringValue("cert-default"),
		LinkedActions:         stringSet(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.SSLCertificates != nil {
		t.Fatalf("pre-create omitted ssl_certificates = %#v, want nil", got.SSLCertificates)
	}
	if err := prepareFrontendCreate(t.Context(), got); err != nil {
		t.Fatal(err)
	}
	if got.SSLCertificates.String() != "cert-default" {
		t.Fatalf("create certificates = %q, want default only", got.SSLCertificates.String())
	}

	explicit, err := frontendModelToAPI(&frontendModel{
		Enabled:               types.BoolValue(true),
		Name:                  types.StringValue("explicit"),
		Bind:                  stringSet("10.0.0.11:443"),
		Mode:                  types.StringValue("http"),
		SSLEnabled:            types.BoolValue(true),
		SSLCertificates:       stringSet("cert-default", "cert-consumer"),
		SSLDefaultCertificate: types.StringValue("cert-default"),
		LinkedActions:         stringSet(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := prepareFrontendCreate(t.Context(), explicit); err != nil {
		t.Fatal(err)
	}
	if explicit.SSLCertificates.String() != "cert-consumer,cert-default" {
		t.Fatalf("explicit certificate ownership changed on create: %q", explicit.SSLCertificates.String())
	}
}

func TestSNIACLAndUseBackendConversion(t *testing.T) {
	t.Parallel()
	acl, err := aclModelToAPI(&aclModel{
		Name: types.StringValue("sni_rigi"), Expression: types.StringValue("ssl_sni"),
		Negate: types.BoolValue(false), CaseSensitive: types.BoolValue(false),
		SSLSNI: types.StringValue("web.rigi.host.example.test"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if acl.Expression.String() != "ssl_sni" || acl.SSLSNI != "web.rigi.host.example.test" {
		t.Fatalf("unexpected SNI ACL: %+v", acl)
	}

	action, err := actionModelToAPI(&actionModel{
		Enabled: types.BoolValue(true), Name: types.StringValue("route_rigi"), TestType: types.StringValue("if"),
		LinkedACLs: stringSet("acl-id"), Operator: types.StringValue("and"), Type: types.StringValue("use_backend"),
		UseBackend: types.StringValue("backend-id"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if action.Type.String() != "use_backend" || action.UseBackend.String() != "backend-id" || action.LinkedACLs.String() != "acl-id" {
		t.Fatalf("unexpected use_backend action: %+v", action)
	}
}

func TestBackendStickinessPatternRoundTrip(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		apiValue string
		want     string
	}{
		{apiValue: "", want: "none"},
		{apiValue: "sourceipv4", want: "sourceipv4"},
		{apiValue: "sourceipv6", want: "sourceipv6"},
	} {
		got := backendStickinessPatternFromAPI(api.SelectedMap(tc.apiValue))
		if got.ValueString() != tc.want {
			t.Fatalf("API stickiness %q -> %q, want %q", tc.apiValue, got.ValueString(), tc.want)
		}
	}
	none := backendStickinessPatternToAPI("none")
	if got := none.String(); got != "" {
		t.Fatalf("none -> API stickiness %q, want empty", got)
	}
}

func TestBackendServerHealthcheckConversion(t *testing.T) {
	t.Parallel()
	server, err := serverModelToAPI(&serverModel{
		Enabled: types.BoolValue(true), Name: types.StringValue("endpoint"), HAPolicy: types.StringValue("endpoint"), Address: types.StringValue("10.0.0.20"),
		Port: types.Int64Value(443), Mode: types.StringValue("active"), Type: types.StringValue("static"), SSL: types.BoolValue(true),
		SSLSNI: types.StringValue("deneb.svc.sandbox.biptec.net"), SSLVerify: types.BoolValue(true),
		SSLVerifyHost: types.StringValue("deneb.svc.sandbox.biptec.net"), SSLCA: stringSet("ca-ref"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if server.Address != "10.0.0.20" || server.Port != "443" || server.SSL != "1" || server.HAPolicy.String() != "endpoint" ||
		server.SSLSNI != "deneb.svc.sandbox.biptec.net" || server.SSLVerify != "1" ||
		server.SSLVerifyHost != "deneb.svc.sandbox.biptec.net" || server.SSLCA.String() != "ca-ref" {
		t.Fatalf("unexpected TLS server: %+v", server)
	}

	backend, err := backendModelToAPI(&backendModel{
		Enabled: types.BoolValue(true), Name: types.StringValue("endpoint_backend"), HAPolicy: types.StringValue("endpoint"), Mode: types.StringValue("tcp"),
		Algorithm: types.StringValue("roundrobin"), StickinessPattern: types.StringValue("none"), ProxyProtocol: types.StringValue("v2"), LinkedServers: stringSet("server-id"), HealthCheckEnabled: types.BoolValue(true),
		HealthCheck: types.StringValue("check-id"), HealthCheckProxyProtocol: types.StringValue("backend"), HealthCheckFall: types.Int64Value(3), HealthCheckRise: types.Int64Value(2),
	})
	if err != nil {
		t.Fatal(err)
	}
	if backend.Mode.String() != "tcp" || backend.HAPolicy.String() != "endpoint" || backend.StickinessPattern.String() != "" || backend.ProxyProtocol.String() != "v2" || backend.LinkedServers.String() != "server-id" || backend.HealthCheck.String() != "check-id" || backend.HealthCheckProxyProto.String() != "backend" || backend.HealthCheckFall != "3" || backend.HealthCheckRise != "2" {
		t.Fatalf("unexpected backend: %+v", backend)
	}

	check, err := healthcheckModelToAPI(&healthcheckModel{
		Name: types.StringValue("tcp_check"), HAPolicy: types.StringValue("endpoint"), Type: types.StringValue("tcp"), Interval: types.StringValue("2s"),
		SSL: types.StringValue("nopref"), ForceSSL: types.BoolValue(false),
	})
	if err != nil {
		t.Fatal(err)
	}
	if check.Type.String() != "tcp" || check.Interval != "2s" || check.HAPolicy.String() != "endpoint" {
		t.Fatalf("unexpected check: %+v", check)
	}

	httpsCheck, err := healthcheckModelToAPI(&healthcheckModel{
		Name: types.StringValue("oidc_https"), HAPolicy: types.StringValue("endpoint"), Type: types.StringValue("http"), Interval: types.StringValue("2s"),
		SSL: types.StringValue("sslsni"), SSLSNI: types.StringValue("deneb.svc.sandbox.biptec.net"),
		HTTPMethod: types.StringValue("get"), HTTPURI: types.StringValue("/realms/sso/.well-known/openid-configuration"),
		HTTPVersion: types.StringValue("http11"), HTTPHost: types.StringValue("id.central.sandbox.biptec.net"),
		HTTPExpressionEnabled: types.BoolValue(true), HTTPExpression: types.StringValue("string"),
		HTTPNegate: types.BoolValue(false), HTTPValue: types.StringValue("https://id.central.sandbox.biptec.net/realms/sso"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if httpsCheck.Type.String() != "http" || httpsCheck.SSL.String() != "sslsni" || httpsCheck.SSLSNI != "deneb.svc.sandbox.biptec.net" ||
		httpsCheck.HTTPMethod.String() != "get" || httpsCheck.HTTPVersion.String() != "http11" ||
		httpsCheck.HTTPHost != "id.central.sandbox.biptec.net" || httpsCheck.HTTPExpressionEnabled != "1" ||
		httpsCheck.HTTPExpression.String() != "string" || httpsCheck.HTTPValue != "https://id.central.sandbox.biptec.net/realms/sso" {
		t.Fatalf("unexpected HTTPS check: %+v", httpsCheck)
	}
}

func TestHAProxyAPIToModelRoundTrip(t *testing.T) {
	t.Parallel()
	backend, err := backendAPIToModel(&apihaproxy.Backend{
		Enabled: "1", Name: "proxy_v2", HAPolicy: api.SelectedMap("endpoint"), Mode: api.SelectedMap("tcp"), Algorithm: api.SelectedMap("roundrobin"),
		ProxyProtocol: api.SelectedMap("v2"), LinkedServers: api.SelectedMapList{"srv"}, HealthCheckEnabled: "1",
		HealthCheckProxyProto: api.SelectedMap("disable"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if backend.HAPolicy.ValueString() != "endpoint" || backend.ProxyProtocol.ValueString() != "v2" || backend.HealthCheckProxyProtocol.ValueString() != "disable" {
		t.Fatalf("unexpected backend PROXY state: %+v", backend)
	}

	frontend, err := frontendAPIToModel(&apihaproxy.Frontend{
		Enabled: "1", Name: "tls", HAPolicy: api.SelectedMap("endpoint"), Bind: api.SelectedMapList{"[2001:db8::1]:443", "192.0.2.1:443"},
		Mode: api.SelectedMap("http"), SSLEnabled: "1", SSLCertificates: api.SelectedMapList{"cert-b", "cert-a"}, SSLDefaultCertificate: api.SelectedMap("cert-a"), LinkedActions: api.SelectedMapList{"a2", "a1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if frontend.HAPolicy.ValueString() != "endpoint" || frontend.Mode.ValueString() != "http" || !frontend.SSLEnabled.ValueBool() || frontend.SSLDefaultCertificate.ValueString() != "cert-a" || tools.SetToString(frontend.SSLCertificates, ",") != "cert-a,cert-b" {
		t.Fatalf("unexpected TLS frontend: %+v", frontend)
	}
	if tools.SetToString(frontend.LinkedActions, ",") != "a1,a2" {
		t.Fatalf("unexpected actions: %s", tools.SetToString(frontend.LinkedActions, ","))
	}
}

func TestHAProxyUpstreamDefaultsRoundTrip(t *testing.T) {
	t.Parallel()

	acl, err := aclAPIToModel(&apihaproxy.ACL{
		Name: "client_hello", Expression: api.SelectedMap("ssl_sni"), SSLHelloType: api.SelectedMap("x1"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if acl.SSLHelloType.ValueString() != "x1" {
		t.Fatalf("unexpected ssl_hello_type default: %q", acl.SSLHelloType.ValueString())
	}

	check, err := healthcheckAPIToModel(&apihaproxy.Healthcheck{
		Name: "tcp_check", HAPolicy: api.SelectedMap("endpoint"), Type: api.SelectedMap("tcp"), Interval: "2s", SSL: api.SelectedMap("nopref"),
		HTTPMethod: api.SelectedMap("options"), HTTPURI: "/", HTTPVersion: api.SelectedMap("http10"), HTTPHost: "localhost",
		HTTPExpressionEnabled: "0", HTTPNegate: "0", TCPMatchType: api.SelectedMap("string"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if check.HAPolicy.ValueString() != "endpoint" || check.HTTPMethod.ValueString() != "options" || check.HTTPURI.ValueString() != "/" ||
		check.HTTPVersion.ValueString() != "http10" || check.HTTPHost.ValueString() != "localhost" ||
		check.HTTPExpressionEnabled.ValueBool() || check.HTTPNegate.ValueBool() || check.TCPMatchType.ValueString() != "string" {
		t.Fatalf("unexpected OPNsense health-check defaults: %+v", check)
	}
}
