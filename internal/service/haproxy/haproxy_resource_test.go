package haproxy_test

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/biptec/terraform-provider-opnsense/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func haproxyPreCheck(t *testing.T) {
	t.Helper()
	acctest.AccPreCheck(t)
	req, err := http.NewRequest(http.MethodGet, os.Getenv("OPNSENSE_URI")+"/api/haproxy/settings/get", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.SetBasicAuth(os.Getenv("OPNSENSE_API_KEY"), os.Getenv("OPNSENSE_API_SECRET"))
	client := &http.Client{Timeout: 15 * time.Second, Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: os.Getenv("OPNSENSE_ALLOW_INSECURE") == "true"}}}
	resp, err := client.Do(req)
	if err != nil {
		t.Skipf("os-haproxy API is unavailable: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Skipf("os-haproxy is not installed or API access is unavailable: HTTP %d", resp.StatusCode)
	}
}

func TestAccHAProxyStatusSettingsAndConfigtestDataSources(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { haproxyPreCheck(t) }, ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{{Config: `
data "opnsense_haproxy_settings" "test" {}
data "opnsense_haproxy_status" "test" {}
data "opnsense_haproxy_configtest" "test" {}
`, Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("data.opnsense_haproxy_settings.test", "id", "haproxy_settings"),
			resource.TestCheckResourceAttrSet("data.opnsense_haproxy_settings.test", "enabled"),
			resource.TestCheckResourceAttrSet("data.opnsense_haproxy_settings.test", "show_intro"),
			resource.TestCheckResourceAttr("data.opnsense_haproxy_status.test", "id", "haproxy_status"),
			resource.TestCheckResourceAttrSet("data.opnsense_haproxy_status.test", "status"),
			resource.TestCheckResourceAttr("data.opnsense_haproxy_configtest.test", "id", "haproxy_configtest"),
			resource.TestCheckResourceAttrSet("data.opnsense_haproxy_configtest.test", "result"),
		)}},
	})
}

func TestAccHAProxySettingsResourceAdoptsSingleton(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { haproxyPreCheck(t) }, ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: `
resource "opnsense_haproxy_settings" "test" {
  show_intro = false
}
`, Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("opnsense_haproxy_settings.test", "id", "haproxy_settings"),
				resource.TestCheckResourceAttrSet("opnsense_haproxy_settings.test", "enabled"),
				resource.TestCheckResourceAttr("opnsense_haproxy_settings.test", "show_intro", "false"),
				resource.TestCheckResourceAttrSet("opnsense_haproxy_settings.test", "graceful_stop"),
				resource.TestCheckResourceAttrSet("opnsense_haproxy_settings.test", "seamless_reload"),
				checkHAProxyNoPendingDiff(),
			)},
			{ResourceName: "opnsense_haproxy_settings.test", ImportState: true, ImportStateVerify: true},
		},
	})
}

func TestAccHAProxyL4SNIResources(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { haproxyPreCheck(t) }, ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: testAccHAProxyL4SNIConfig("initial", 18443), Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttrSet("opnsense_haproxy_server.endpoint", "id"),
				resource.TestCheckResourceAttr("opnsense_haproxy_server.endpoint", "address", "127.0.0.1"),
				resource.TestCheckResourceAttr("opnsense_haproxy_server.endpoint", "port", "9443"),
				resource.TestCheckResourceAttr("opnsense_haproxy_healthcheck.tcp", "type", "tcp"),
				resource.TestCheckResourceAttr("opnsense_haproxy_backend.endpoint", "mode", "tcp"),
				resource.TestCheckResourceAttr("opnsense_haproxy_backend.endpoint", "linked_servers.#", "1"),
				resource.TestCheckResourceAttr("opnsense_haproxy_acl.sni", "expression", "ssl_sni"),
				resource.TestCheckResourceAttr("opnsense_haproxy_acl.sni", "ssl_sni", "tfacc.example.invalid"),
				resource.TestCheckResourceAttr("opnsense_haproxy_action.route", "type", "use_backend"),
				resource.TestCheckResourceAttr("opnsense_haproxy_frontend.tls", "mode", "tcp"),
				resource.TestCheckResourceAttr("opnsense_haproxy_frontend.tls", "ssl_enabled", "false"),
				resource.TestCheckTypeSetElemAttr("opnsense_haproxy_frontend.tls", "bind.*", "127.0.0.1:18443"),
				resource.TestCheckResourceAttr("data.opnsense_haproxy_acl.sni", "ssl_sni", "tfacc.example.invalid"),
				checkHAProxyNoPendingDiff(),
			)},
			{ResourceName: "opnsense_haproxy_server.endpoint", ImportState: true, ImportStateVerify: true},
			{ResourceName: "opnsense_haproxy_healthcheck.tcp", ImportState: true, ImportStateVerify: true},
			{ResourceName: "opnsense_haproxy_backend.endpoint", ImportState: true, ImportStateVerify: true},
			{ResourceName: "opnsense_haproxy_acl.sni", ImportState: true, ImportStateVerify: true},
			{ResourceName: "opnsense_haproxy_action.route", ImportState: true, ImportStateVerify: true},
			{ResourceName: "opnsense_haproxy_frontend.tls", ImportState: true, ImportStateVerify: true},
			{Config: testAccHAProxyL4SNIConfig("updated", 18444), Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("opnsense_haproxy_server.endpoint", "description", "updated"),
				resource.TestCheckResourceAttr("opnsense_haproxy_backend.endpoint", "description", "updated"),
				resource.TestCheckResourceAttr("opnsense_haproxy_acl.sni", "description", "updated"),
				resource.TestCheckResourceAttr("opnsense_haproxy_action.route", "description", "updated"),
				resource.TestCheckResourceAttr("opnsense_haproxy_frontend.tls", "description", "updated"),
				resource.TestCheckTypeSetElemAttr("opnsense_haproxy_frontend.tls", "bind.*", "127.0.0.1:18444"),
				checkHAProxyNoPendingDiff(),
			)},
		},
	})
}

func TestAccHAProxyServerRejectsDuplicateName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { haproxyPreCheck(t) }, ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: testAccHAProxyUniqueServerConfig(false)},
			{Config: testAccHAProxyUniqueServerConfig(true), ExpectError: regexp.MustCompile(`HAProxy server name "tfacc-unique-server" is already used`)},
		},
	})
}

func TestAccHAProxyBackendRejectsDuplicateName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { haproxyPreCheck(t) }, ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: testAccHAProxyUniqueBackendConfig(false)},
			{Config: testAccHAProxyUniqueBackendConfig(true), ExpectError: regexp.MustCompile(`HAProxy backend name "tfacc-unique-backend" is already used`)},
		},
	})
}

func testAccHAProxyUniqueServerConfig(withDuplicate bool) string {
	duplicate := ""
	if withDuplicate {
		duplicate = `
resource "opnsense_haproxy_server" "duplicate" {
  name    = "tfacc-unique-server"
  address = "127.0.0.2"
  port    = 9443
}`
	}
	return `
resource "opnsense_haproxy_server" "original" {
  name    = "tfacc-unique-server"
  address = "127.0.0.1"
  port    = 9443
}
` + duplicate
}

func testAccHAProxyUniqueBackendConfig(withDuplicate bool) string {
	duplicate := ""
	if withDuplicate {
		duplicate = `
resource "opnsense_haproxy_server" "duplicate" {
  name    = "tfacc-unique-backend-server-2"
  address = "127.0.0.2"
  port    = 9443
}
resource "opnsense_haproxy_backend" "duplicate" {
  name                 = "tfacc-unique-backend"
  mode                 = "tcp"
  linked_servers       = [opnsense_haproxy_server.duplicate.id]
  health_check_enabled = false
}`
	}
	return `
resource "opnsense_haproxy_server" "original" {
  name    = "tfacc-unique-backend-server-1"
  address = "127.0.0.1"
  port    = 9443
}
resource "opnsense_haproxy_backend" "original" {
  name                 = "tfacc-unique-backend"
  mode                 = "tcp"
  linked_servers       = [opnsense_haproxy_server.original.id]
  health_check_enabled = false
}
` + duplicate
}

func testAccHAProxyL4SNIConfig(description string, listenPort int) string {
	return fmt.Sprintf(`
resource "opnsense_haproxy_server" "endpoint" {
  name        = "tfacc-endpoint"
  description = %[1]q
  address     = "127.0.0.1"
  port        = 9443
  mode        = "active"
  type        = "static"
  ssl         = false
}
resource "opnsense_haproxy_healthcheck" "tcp" {
  name        = "tfacc-tcp"
  description = %[1]q
  type        = "tcp"
  interval    = "2s"
}
resource "opnsense_haproxy_backend" "endpoint" {
  name                 = "tfacc-backend"
  description          = %[1]q
  mode                 = "tcp"
  algorithm            = "roundrobin"
  linked_servers       = [opnsense_haproxy_server.endpoint.id]
  health_check_enabled = true
  health_check         = opnsense_haproxy_healthcheck.tcp.id
}
resource "opnsense_haproxy_acl" "sni" {
  name        = "tfacc-sni"
  description = %[1]q
  expression  = "ssl_sni"
  ssl_sni     = "tfacc.example.invalid"
}
resource "opnsense_haproxy_action" "route" {
  name         = "tfacc-route"
  description  = %[1]q
  type         = "use_backend"
  test_type    = "if"
  operator     = "and"
  linked_acls  = [opnsense_haproxy_acl.sni.id]
  use_backend  = opnsense_haproxy_backend.endpoint.id
}
resource "opnsense_haproxy_frontend" "tls" {
  name           = "tfacc-tls-ingress"
  description    = %[1]q
  bind           = ["127.0.0.1:%[2]d"]
  mode           = "tcp"
  ssl_enabled    = false
  custom_options = "tcp-request inspect-delay 5s"
  linked_actions = [opnsense_haproxy_action.route.id]
}
data "opnsense_haproxy_acl" "sni" { id = opnsense_haproxy_acl.sni.id }
`, description, listenPort)
}
