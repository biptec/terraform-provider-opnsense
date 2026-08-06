package bind_test

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

func TestAccBindMultipleListenerAddresses(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { bindPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccBindSettingsConfig(),
				ResourceName:       "opnsense_bind_settings.test",
				ImportState:        true,
				ImportStateId:      "bind_settings",
				ImportStateVerify:  false,
				ImportStatePersist: true,
			},
			{
				Config: testAccBindMultipleListenerConfig(true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_bind_settings.test", "enabled", "true"),
					resource.TestCheckResourceAttr("opnsense_bind_settings.test", "port", "53531"),
					resource.TestCheckResourceAttr("opnsense_bind_settings.test", "listen_ipv4.#", "2"),
					resource.TestCheckTypeSetElemAttr("opnsense_bind_settings.test", "listen_ipv4.*", "192.0.2.2"),
					resource.TestCheckTypeSetElemAttr("opnsense_bind_settings.test", "listen_ipv4.*", "198.51.100.231"),
					resource.TestCheckResourceAttr("opnsense_interfaces_vip.public", "mode", "ipalias"),
					resource.TestCheckResourceAttr("opnsense_interfaces_vip.public", "interface", "wan"),
					resource.TestCheckResourceAttr("opnsense_interfaces_vip.public", "network", "198.51.100.231/32"),
					checkBindSocketListeners("tcp", 53531, []string{"192.0.2.2", "198.51.100.231"}),
					checkBindSocketListeners("udp", 53531, []string{"192.0.2.2", "198.51.100.231"}),
				),
			},
			{
				Config: testAccBindMultipleListenerConfig(false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_bind_settings.test", "enabled", "false"),
					resource.TestCheckResourceAttr("opnsense_bind_settings.test", "listen_ipv4.#", "1"),
					resource.TestCheckTypeSetElemAttr("opnsense_bind_settings.test", "listen_ipv4.*", "127.0.0.1"),
					checkBindSocketListeners("tcp", 53531, nil),
					checkBindSocketListeners("udp", 53531, nil),
				),
			},
		},
	})
}

func testAccBindMultipleListenerConfig(enabled bool) string {
	addresses := `["127.0.0.1"]`
	if enabled {
		addresses = `["192.0.2.2", "198.51.100.231"]`
	}
	return fmt.Sprintf(`
resource "opnsense_interfaces_loopback" "dns_a" {
  description = "DNS frontend A"
}

resource "opnsense_interfaces_assignment" "dns_a" {
  device          = format("lo%%d", opnsense_interfaces_loopback.dns_a.device_id)
  description     = "DNS frontend A"
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

resource "opnsense_interfaces_vip" "public" {
  mode        = "ipalias"
  interface   = "wan"
  network     = "198.51.100.231/32"
  description = "DNS public frontend"
}

resource "opnsense_bind_settings" "test" {
  enabled      = %[1]t
  disable_ipv6 = true
  listen_ipv4  = %[2]s
  listen_ipv6  = ["::1"]
  port         = 53531

  depends_on = [
    opnsense_interfaces_assignment.dns_a,
    opnsense_interfaces_vip.public,
  ]
}
`, enabled, addresses)
}

func checkBindSocketListeners(protocol string, port int, expected []string) resource.TestCheckFunc {
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
			output, err := acctest.QGAGuestExec(ctx, "/usr/bin/sockstat", "-4", "-l", "-P", protocol)
			if err == nil {
				lastOutput = output
				actualSet := bindListenerSet(output, pattern)
				if bindSameStringSet(actualSet, expectedSet) {
					return nil
				}
			}
			time.Sleep(time.Second)
		}
		return fmt.Errorf("BIND %s listeners on port %d are %v, expected %v; sockstat: %s", protocol, port, bindSortedSet(lastOutput, pattern), expected, strings.TrimSpace(lastOutput))
	}
}

func bindListenerSet(output string, pattern *regexp.Regexp) map[string]struct{} {
	values := make(map[string]struct{})
	for _, line := range strings.Split(output, "\n") {
		match := pattern.FindStringSubmatch(line)
		if len(match) == 2 {
			values[match[1]] = struct{}{}
		}
	}
	return values
}

func bindSameStringSet(actual, expected map[string]struct{}) bool {
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

func bindSortedSet(output string, pattern *regexp.Regexp) []string {
	values := bindListenerSet(output, pattern)
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
