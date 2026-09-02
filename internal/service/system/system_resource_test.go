package system_test

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/biptec/opnsense-go/pkg/api"
	apiextensions "github.com/biptec/opnsense-go/pkg/api_extensions"
	"github.com/biptec/opnsense-go/pkg/opnsense"
	"github.com/biptec/terraform-provider-opnsense/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var originalWebgui apiextensions.WebguiSettings

func systemClient() opnsense.Client {
	return opnsense.NewClient(api.NewClient(api.Options{
		Uri:           os.Getenv("OPNSENSE_URI"),
		APIKey:        os.Getenv("OPNSENSE_API_KEY"),
		APISecret:     os.Getenv("OPNSENSE_API_SECRET"),
		AllowInsecure: os.Getenv("OPNSENSE_ALLOW_INSECURE") == "true",
	}))
}

func systemPreCheck(t *testing.T) {
	t.Helper()
	acctest.AccPreCheck(t)
	client := systemClient().ApiExtensions()
	result, err := client.WebguiGet(context.Background())
	if err != nil {
		t.Fatalf("os-api-extensions Web GUI API is unavailable: %v", err)
	}
	packageState, err := client.PackageGet(context.Background(), "os-api-extensions")
	if err != nil {
		t.Fatalf("os-api-extensions local package status API is unavailable: %v", err)
	}
	if packageState.Status != "ok" || !packageState.Package.Installed || packageState.Package.Name != "os-api-extensions" {
		t.Fatalf("unexpected os-api-extensions local package state: %#v", packageState)
	}
	originalWebgui = result.Webgui
}

func TestAccSystemWebguiResource(t *testing.T) {
	systemPreCheck(t)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWebguiConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_system_webgui.test", "id", "system_webgui"),
					resource.TestCheckResourceAttr("opnsense_system_webgui.test", "interfaces.#", "1"),
					resource.TestCheckResourceAttr("opnsense_system_webgui.test", "interfaces.0", "lan"),
				),
			},
			{
				ResourceName:            "opnsense_system_webgui.test",
				ImportState:             true,
				ImportStateId:           "system_webgui",
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_readdress"},
			},
		},
	})
}

func TestAccSystemSshResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { systemPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSshConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_system_ssh.test", "id", "system_ssh"),
					resource.TestCheckResourceAttr("opnsense_system_ssh.test", "enabled", "true"),
					resource.TestCheckResourceAttr("opnsense_system_ssh.test", "interfaces.#", "1"),
					resource.TestCheckResourceAttr("opnsense_system_ssh.test", "password_authentication", "false"),
					resource.TestCheckResourceAttr("opnsense_system_ssh.test", "permit_root_login", "false"),
				),
			},
			{
				ResourceName:            "opnsense_system_ssh.test",
				ImportState:             true,
				ImportStateId:           "system_ssh",
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_readdress"},
			},
		},
	})
}

func TestAccSystemDnsResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { systemPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSystemDnsConfig("1.1.1.1", "8.8.8.8"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_system_dns.test", "id", "system_dns"),
					resource.TestCheckResourceAttr("opnsense_system_dns.test", "servers.#", "2"),
					resource.TestCheckResourceAttr("opnsense_system_dns.test", "servers.0", "1.1.1.1"),
					resource.TestCheckResourceAttr("opnsense_system_dns.test", "servers.1", "8.8.8.8"),
					resource.TestCheckResourceAttr("opnsense_system_dns.test", "allow_override", "false"),
					resource.TestCheckResourceAttr("opnsense_system_dns.test", "use_local_service", "false"),
					checkSystemDnsRuntime("1.1.1.1", "8.8.8.8"),
				),
			},
			{
				Config: testAccSystemDnsConfig("9.9.9.9", "1.1.1.1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_system_dns.test", "servers.0", "9.9.9.9"),
					resource.TestCheckResourceAttr("opnsense_system_dns.test", "servers.1", "1.1.1.1"),
					checkSystemDnsRuntime("9.9.9.9", "1.1.1.1"),
				),
			},
			{
				ResourceName:      "opnsense_system_dns.test",
				ImportState:       true,
				ImportStateId:     "system_dns",
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNtpSettingsResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { systemPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNtpConfig(13, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_ntp_settings.test", "id", "ntp_settings"),
					resource.TestCheckResourceAttr("opnsense_ntp_settings.test", "enabled", "true"),
					resource.TestCheckResourceAttr("opnsense_ntp_settings.test", "servers.#", "1"),
					resource.TestCheckResourceAttr("opnsense_ntp_settings.test", "interfaces.#", "1"),
					resource.TestCheckResourceAttr("opnsense_ntp_settings.test", "orphan", "13"),
					resource.TestCheckResourceAttrSet("opnsense_interfaces_assignment.ntp", "name"),
					checkNtpRuntimeListener("192.0.2.2"),
				),
			},
			{
				ResourceName:      "opnsense_ntp_settings.test",
				ImportState:       true,
				ImportStateId:     "ntp_settings",
				ImportStateVerify: true,
			},
			{
				Config: testAccNtpConfig(12, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_ntp_settings.test", "orphan", "12"),
					checkNtpRuntimeListener("192.0.2.2"),
				),
			},
			{
				Config: testAccNtpConfig(12, false),
				Check:  resource.TestCheckResourceAttr("opnsense_ntp_settings.test", "interfaces.0", "lan"),
			},
		},
	})
}

func TestAccPluginLocalRefreshLatency(t *testing.T) {
	acctest.AccPreCheck(t)
	client := systemClient().ApiExtensions()
	started := time.Now()
	for i := 0; i < 5; i++ {
		state, err := client.PackageGet(context.Background(), "os-api-extensions")
		if err != nil {
			t.Fatalf("local package status request %d failed: %v", i+1, err)
		}
		if state.Status != "ok" || !state.Package.Installed {
			t.Fatalf("local package status request %d returned unexpected state: %#v", i+1, state)
		}
	}
	if elapsed := time.Since(started); elapsed > 5*time.Second {
		t.Fatalf("five local package refreshes took %s, want <= 5s", elapsed)
	}
}

func TestAccPluginResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { systemPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPluginConfig(false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_plugin.test", "id", "os-api-extensions"),
					resource.TestCheckResourceAttr("opnsense_plugin.test", "installed", "true"),
					resource.TestCheckResourceAttrSet("opnsense_plugin.test", "version"),
				),
			},
			{
				ResourceName:            "opnsense_plugin.test",
				ImportState:             true,
				ImportStateId:           "os-api-extensions",
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"uninstall_on_destroy"},
			},
			{
				Config: testAccPluginConfig(true, false),
				Check:  resource.TestCheckResourceAttr("opnsense_plugin.test", "locked", "true"),
			},
			{
				Config: testAccPluginConfig(false, true),
				Check:  resource.TestCheckResourceAttr("opnsense_plugin.test", "locked", "false"),
			},
		},
	})
}

func testAccWebguiConfig() string {
	certificate := strconv.Quote(originalWebgui.CertificateRef)
	timeout := ""
	if originalWebgui.SessionTimeout != nil {
		timeout = fmt.Sprintf("  session_timeout_minutes = %d\n", *originalWebgui.SessionTimeout)
	}
	return fmt.Sprintf(`
resource "opnsense_system_webgui" "test" {
  protocol              = %[1]q
  port                  = %[2]d
  interfaces            = ["lan"]
  certificate_ref       = %[3]s
%[4]s  hsts                  = %[5]t
  disable_http_redirect = %[6]t
  alternate_hostnames   = %[7]s
  allow_readdress       = true
}
`, originalWebgui.Protocol, originalWebgui.Port, certificate, timeout,
		originalWebgui.HSTS, originalWebgui.DisableHTTPRedirect, hclStringSet(originalWebgui.AlternateHostnames))
}

func testAccSshConfig() string {
	return `
resource "opnsense_system_ssh" "test" {
  enabled                 = true
  port                    = 22
  interfaces              = ["lan"]
  password_authentication = false
  permit_root_login       = false
  allow_readdress         = true
}
`
}

func testAccSystemDnsConfig(primary, secondary string) string {
	return fmt.Sprintf(`
resource "opnsense_system_dns" "test" {
  servers           = [%q, %q]
  allow_override    = false
  use_local_service = false
}
`, primary, secondary)
}

func testAccNtpConfig(orphan int, useLoopback bool) string {
	listener := `"lan"`
	if useLoopback {
		listener = "opnsense_interfaces_assignment.ntp.name"
	}
	return fmt.Sprintf(`
resource "opnsense_interfaces_loopback" "ntp" {
  description = "NTP service loopback"
}

resource "opnsense_interfaces_assignment" "ntp" {
  device          = format("lo%%d", opnsense_interfaces_loopback.ntp.device_id)
  description     = "NTP service"
  enabled         = true
  allow_readdress = true

  ipv4 = {
    mode    = "static"
    address = "192.0.2.2"
    prefix  = 30
  }

  ipv6 = {
    mode = "none"
  }
}

resource "opnsense_ntp_settings" "test" {
  enabled    = true
  interfaces = [%[2]s]
  orphan     = %[1]d
  max_clock  = 10

  servers = [{
    host     = "0.opnsense.pool.ntp.org"
    noselect = false
    prefer   = true
    iburst   = true
    pool     = true
  }]
}
`, orphan, listener)
}

func testAccPluginConfig(locked, uninstall bool) string {
	return fmt.Sprintf(`
resource "opnsense_plugin" "test" {
  name                 = "os-api-extensions"
  locked               = %[1]t
  uninstall_on_destroy = %[2]t
}
`, locked, uninstall)
}

func hclStringSet(values []string) string {
	sorted := append([]string(nil), values...)
	sort.Strings(sorted)
	quoted := make([]string, 0, len(sorted))
	for _, value := range sorted {
		quoted = append(quoted, strconv.Quote(value))
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}
