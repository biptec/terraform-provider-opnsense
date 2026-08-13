package system_test

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/biptec/opnsense-go/pkg/api"
	apiextensions "github.com/biptec/opnsense-go/pkg/api_extensions"
	"github.com/biptec/opnsense-go/pkg/opnsense"
	"github.com/biptec/terraform-provider-opnsense/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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
	result, err := systemClient().ApiExtensions().WebguiGet(context.Background())
	if err != nil {
		t.Skipf("os-api-extensions is unavailable: %v", err)
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

func TestAccPluginResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { systemPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             checkPluginUninstalled("os-acme-client"),
		Steps: []resource.TestStep{
			{
				Config: testAccPluginConfig(false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_plugin.test", "id", "os-acme-client"),
					resource.TestCheckResourceAttr("opnsense_plugin.test", "installed", "true"),
					resource.TestCheckResourceAttrSet("opnsense_plugin.test", "version"),
				),
			},
			{
				ResourceName:            "opnsense_plugin.test",
				ImportState:             true,
				ImportStateId:           "os-acme-client",
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
  name                 = "os-acme-client"
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

func checkPluginUninstalled(name string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		info, err := systemClient().Core().FirmwareInfo(context.Background())
		if err != nil {
			return err
		}
		for _, plugin := range info.Plugins {
			if plugin.Name == name && plugin.Installed == "1" {
				return fmt.Errorf("plugin %s remains installed", name)
			}
		}
		return nil
	}
}
