package haproxy

import (
	"testing"

	"github.com/biptec/opnsense-go/pkg/api"
	apihaproxy "github.com/biptec/opnsense-go/pkg/haproxy"
	"github.com/biptec/terraform-provider-opnsense/internal/tools"
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

func TestBackendServerHealthcheckConversion(t *testing.T) {
	t.Parallel()
	server, err := serverModelToAPI(&serverModel{
		Enabled: types.BoolValue(true), Name: types.StringValue("endpoint"), Address: types.StringValue("10.0.0.20"),
		Port: types.Int64Value(443), Mode: types.StringValue("active"), Type: types.StringValue("static"), SSL: types.BoolValue(false),
	})
	if err != nil {
		t.Fatal(err)
	}
	if server.Address != "10.0.0.20" || server.Port != "443" || server.SSL != "0" {
		t.Fatalf("unexpected server: %+v", server)
	}

	backend, err := backendModelToAPI(&backendModel{
		Enabled: types.BoolValue(true), Name: types.StringValue("endpoint_backend"), Mode: types.StringValue("tcp"),
		Algorithm: types.StringValue("roundrobin"), LinkedServers: stringSet("server-id"), HealthCheckEnabled: types.BoolValue(true),
		HealthCheck: types.StringValue("check-id"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if backend.Mode.String() != "tcp" || backend.LinkedServers.String() != "server-id" || backend.HealthCheck.String() != "check-id" {
		t.Fatalf("unexpected backend: %+v", backend)
	}

	check, err := healthcheckModelToAPI(&healthcheckModel{
		Name: types.StringValue("tcp_check"), Type: types.StringValue("tcp"), Interval: types.StringValue("2s"),
		SSL: types.StringValue("nopref"), ForceSSL: types.BoolValue(false),
	})
	if err != nil {
		t.Fatal(err)
	}
	if check.Type.String() != "tcp" || check.Interval != "2s" {
		t.Fatalf("unexpected check: %+v", check)
	}
}

func TestHAProxyAPIToModelRoundTrip(t *testing.T) {
	t.Parallel()
	frontend, err := frontendAPIToModel(&apihaproxy.Frontend{
		Enabled: "1", Name: "tls", Bind: api.SelectedMapList{"[2001:db8::1]:443", "192.0.2.1:443"},
		Mode: api.SelectedMap("tcp"), SSLEnabled: "0", LinkedActions: api.SelectedMapList{"a2", "a1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if frontend.Mode.ValueString() != "tcp" || frontend.SSLEnabled.ValueBool() {
		t.Fatalf("unexpected frontend: %+v", frontend)
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
		Name: "tcp_check", Type: api.SelectedMap("tcp"), Interval: "2s", SSL: api.SelectedMap("nopref"),
		HTTPMethod: api.SelectedMap("options"), HTTPURI: "/", HTTPHost: "localhost", TCPMatchType: api.SelectedMap("string"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if check.HTTPMethod.ValueString() != "options" || check.HTTPURI.ValueString() != "/" ||
		check.HTTPHost.ValueString() != "localhost" || check.TCPMatchType.ValueString() != "string" {
		t.Fatalf("unexpected OPNsense health-check defaults: %+v", check)
	}
}
