package caddy_test

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/biptec/terraform-provider-opnsense/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccCaddyMultipleListenerAddresses(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { caddyPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccCaddySettingsConfig(80, 443, "root"),
				ResourceName:       "opnsense_caddy_settings.test",
				ImportState:        true,
				ImportStateId:      "caddy_settings",
				ImportStateVerify:  false,
				ImportStatePersist: true,
			},
			{
				Config: testAccCaddyMultipleListenerConfig(true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_caddy_settings.test", "enabled", "true"),
					resource.TestCheckResourceAttr("opnsense_caddy_settings.test", "http_port", "18080"),
					resource.TestCheckResourceAttr("opnsense_caddy_settings.test", "listen_addresses.#", "2"),
					resource.TestCheckTypeSetElemAttr("opnsense_caddy_settings.test", "listen_addresses.*", "192.0.2.2"),
					resource.TestCheckTypeSetElemAttr("opnsense_caddy_settings.test", "listen_addresses.*", "198.51.100.2"),
					checkCaddyTCPListeners(18080, []string{"192.0.2.2", "198.51.100.2"}),
				),
			},
			{
				Config: testAccCaddyMultipleListenerConfig(false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_caddy_settings.test", "enabled", "false"),
					resource.TestCheckResourceAttr("opnsense_caddy_settings.test", "listen_addresses.#", "1"),
					resource.TestCheckTypeSetElemAttr("opnsense_caddy_settings.test", "listen_addresses.*", "127.0.0.1"),
					checkCaddyTCPListeners(18080, nil),
				),
			},
		},
	})
}

func testAccCaddyMultipleListenerConfig(enabled bool) string {
	addresses := `["127.0.0.1"]`
	if enabled {
		addresses = `["192.0.2.2", "198.51.100.2"]`
	}
	return fmt.Sprintf(`
resource "opnsense_interfaces_loopback" "frontend_a" {
  description = "Caddy frontend A"
}

resource "opnsense_interfaces_assignment" "frontend_a" {
  device          = format("lo%%d", opnsense_interfaces_loopback.frontend_a.device_id)
  description     = "Caddy frontend A"
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

resource "opnsense_interfaces_loopback" "frontend_b" {
  description = "Caddy frontend B"
}

resource "opnsense_interfaces_assignment" "frontend_b" {
  device          = format("lo%%d", opnsense_interfaces_loopback.frontend_b.device_id)
  description     = "Caddy frontend B"
  enabled         = true
  allow_readdress = true

  ipv4 = {
    mode    = "static"
    address = "198.51.100.2"
    prefix  = 30
  }

  ipv6 = {
    mode = "none"
  }
}

resource "opnsense_caddy_settings" "test" {
  enabled               = %[1]t
  enable_layer4         = false
  http_port             = 18080
  https_port            = 18443
  listen_addresses      = %[2]s
  acme_email            = ""
  auto_https            = "off"
  run_as_user            = "root"
  grace_period           = 10
  http_versions          = ["h1", "h2"]
  log_level              = ""
  plain_access_log       = false
  plain_access_log_keep  = 10

  depends_on = [
    opnsense_interfaces_assignment.frontend_a,
    opnsense_interfaces_assignment.frontend_b,
  ]
}
`, enabled, addresses)
}

func checkCaddyTCPListeners(port int, expected []string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if acctest.QGASocket() == "" {
			return nil
		}
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()
		expectedSet := make(map[string]struct{}, len(expected))
		for _, address := range expected {
			expectedSet[address] = struct{}{}
		}
		pattern := regexp.MustCompile(`(?:^|\s)(\*|[0-9.]+):` + fmt.Sprint(port) + `(?:\s|$)`)
		var lastOutput string
		for ctx.Err() == nil {
			output, err := acctest.QGAGuestExec(ctx, "/usr/bin/sockstat", "-4", "-l", "-P", "tcp")
			if err == nil {
				lastOutput = output
				actualSet := make(map[string]struct{})
				for _, line := range strings.Split(output, "\n") {
					match := pattern.FindStringSubmatch(line)
					if len(match) == 2 {
						actualSet[match[1]] = struct{}{}
					}
				}
				if sameStringSet(actualSet, expectedSet) {
					return nil
				}
			}
			time.Sleep(time.Second)
		}
		return fmt.Errorf("Caddy TCP listeners on port %d are %v, expected %v; sockstat: %s", port, sortedKeysFromSet(lastOutput, pattern), expected, strings.TrimSpace(lastOutput))
	}
}

func sameStringSet(actual, expected map[string]struct{}) bool {
	if len(actual) != len(expected) {
		return false
	}
	for value := range expected {
		if _, ok := actual[value]; !ok {
			return false
		}
	}
	return true
}

func sortedKeysFromSet(output string, pattern *regexp.Regexp) []string {
	values := make(map[string]struct{})
	for _, line := range strings.Split(output, "\n") {
		match := pattern.FindStringSubmatch(line)
		if len(match) == 2 {
			values[match[1]] = struct{}{}
		}
	}
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
