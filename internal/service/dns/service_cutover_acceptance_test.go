package dns_test

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/biptec/terraform-provider-opnsense/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDNSServiceCutoverResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { dnsCutoverPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccDNSServiceBootstrapConfig(),
				ResourceName:       "opnsense_unbound_service.legacy",
				ImportState:        true,
				ImportStateId:      "unbound_service",
				ImportStatePersist: true,
			},
			{
				Config:             testAccDNSServiceBootstrapConfig(),
				ResourceName:       "opnsense_dnsmasq_settings.legacy",
				ImportState:        true,
				ImportStateId:      "dnsmasq_settings",
				ImportStatePersist: true,
			},
			{
				Config:             testAccDNSServiceBootstrapConfig(),
				ResourceName:       "opnsense_bind_settings.resolver",
				ImportState:        true,
				ImportStateId:      "bind_settings",
				ImportStatePersist: true,
			},
			{
				Config: testAccDNSServiceBootstrapConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_unbound_service.legacy", "enabled", "true"),
					resource.TestCheckResourceAttr("opnsense_dnsmasq_settings.legacy", "dns_port", "0"),
					resource.TestCheckResourceAttr("opnsense_bind_settings.resolver", "enabled", "false"),
					checkDNSPort53Owner("unbound", []string{"*"}),
				),
			},
			{
				Config: testAccDNSServiceCutoverConfig("bind", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_dns_service_cutover.resolver", "target", "bind"),
					resource.TestCheckResourceAttr("opnsense_dns_service_cutover.resolver", "active_service", "bind"),
					resource.TestCheckResourceAttr("opnsense_dns_service_cutover.resolver", "allow_cutover", "false"),
					checkDNSPort53Owner("named", []string{"192.0.2.2", "198.51.100.231"}),
				),
			},
			{
				ResourceName:            "opnsense_dns_service_cutover.resolver",
				ImportState:             true,
				ImportStateId:           "dns_service_cutover",
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_cutover", "verify_timeout_seconds"},
			},
			{
				Config:      testAccDNSServiceCutoverConfig("unbound", false),
				ExpectError: regexp.MustCompile("DNS Cutover Requires Explicit Approval"),
			},
			{
				Config: testAccDNSServiceCutoverConfig("unbound", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_dns_service_cutover.resolver", "active_service", "unbound"),
					checkDNSPort53Owner("unbound", []string{"*"}),
				),
			},
		},
	})
}

func dnsCutoverPreCheck(t *testing.T) {
	t.Helper()
	acctest.AccPreCheck(t)
	if os.Getenv("OPNSENSE_TEST_SPARE_DEVICE_1") == "" {
		t.Fatal("OPNSENSE_TEST_SPARE_DEVICE_1 must be set for DNS cutover acceptance tests")
	}
}

func testAccDNSServiceBootstrapConfig() string {
	return fmt.Sprintf(`
resource "opnsense_unbound_service" "legacy" {
  enabled = true
}

resource "opnsense_dnsmasq_settings" "legacy" {
  enabled  = true
  dns_port = 0
}

resource "opnsense_interfaces_vlan" "dns" {
  parent      = %[1]q
  tag         = 210
  device      = "vlan210"
  protocol    = "802.1q"
  description = "DNS service VLAN"
}

resource "opnsense_interfaces_assignment" "dns" {
  device          = opnsense_interfaces_vlan.dns.device
  description     = "DNS service gateway"
  enabled         = true
  allow_readdress = true

  ipv4 = {
    mode    = "static"
    address = "192.0.2.1"
    prefix  = 30
  }

  ipv6 = {
    mode = "none"
  }
}

resource "opnsense_interfaces_vip" "dns_internal" {
  mode        = "ipalias"
  interface   = opnsense_interfaces_assignment.dns.name
  network     = "192.0.2.2/32"
  no_bind     = false
  no_expand   = true
  description = "DNS internal service address"
}

resource "opnsense_interfaces_vip" "dns_public" {
  mode        = "ipalias"
  interface   = "wan"
  network     = "198.51.100.231/32"
  description = "Public DNS test address"
}

resource "opnsense_bind_settings" "resolver" {
  enabled      = false
  disable_ipv6 = true
  listen_ipv4  = ["192.0.2.2", "198.51.100.231"]
  listen_ipv6  = ["::1"]
  port         = 53

  depends_on = [
    opnsense_interfaces_vip.dns_internal,
    opnsense_interfaces_vip.dns_public,
  ]
}
`, os.Getenv("OPNSENSE_TEST_SPARE_DEVICE_1"))
}

func testAccDNSServiceCutoverConfig(target string, allowCutover bool) string {
	return fmt.Sprintf(`
resource "opnsense_interfaces_vlan" "dns" {
  parent      = %[1]q
  tag         = 210
  device      = "vlan210"
  protocol    = "802.1q"
  description = "DNS service VLAN"
}

resource "opnsense_interfaces_assignment" "dns" {
  device          = opnsense_interfaces_vlan.dns.device
  description     = "DNS service gateway"
  enabled         = true
  allow_readdress = true

  ipv4 = {
    mode    = "static"
    address = "192.0.2.1"
    prefix  = 30
  }

  ipv6 = {
    mode = "none"
  }
}

resource "opnsense_interfaces_vip" "dns_internal" {
  mode        = "ipalias"
  interface   = opnsense_interfaces_assignment.dns.name
  network     = "192.0.2.2/32"
  no_bind     = false
  no_expand   = true
  description = "DNS internal service address"
}

resource "opnsense_interfaces_vip" "dns_public" {
  mode        = "ipalias"
  interface   = "wan"
  network     = "198.51.100.231/32"
  description = "Public DNS test address"
}

resource "opnsense_bind_settings" "resolver" {
  disable_ipv6 = true
  listen_ipv4  = ["192.0.2.2", "198.51.100.231"]
  listen_ipv6  = ["::1"]
  port         = 53

  depends_on = [
    opnsense_interfaces_vip.dns_internal,
    opnsense_interfaces_vip.dns_public,
  ]
}

resource "opnsense_dns_service_cutover" "resolver" {
  target                 = %[2]q
  allow_cutover          = %[3]t
  verify_timeout_seconds = 15

  depends_on = [opnsense_bind_settings.resolver]
}
`, os.Getenv("OPNSENSE_TEST_SPARE_DEVICE_1"), target, allowCutover)
}

func checkDNSPort53Owner(expectedProcess string, expectedAddresses []string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if acctest.QGASocket() == "" {
			return nil
		}
		expected := make([]string, 0, len(expectedAddresses))
		for _, address := range expectedAddresses {
			expected = append(expected, expectedProcess+"@"+address)
		}
		sort.Strings(expected)

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		var last []string
		for ctx.Err() == nil {
			output, err := acctest.QGAGuestExec(ctx, "/usr/bin/sockstat", "-4", "-6", "-l")
			if err == nil {
				last = dnsPort53Owners(output)
				if strings.Join(last, "\x00") == strings.Join(expected, "\x00") {
					return nil
				}
			}
			time.Sleep(time.Second)
		}
		return fmt.Errorf("DNS port 53 owners did not converge to %v; last observed %v", expected, last)
	}
}

func dnsPort53Owners(output string) []string {
	owners := map[string]struct{}{}
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 6 || !strings.HasPrefix(fields[4], "tcp") && !strings.HasPrefix(fields[4], "udp") {
			continue
		}
		if !strings.HasSuffix(fields[5], ":53") {
			continue
		}
		address := strings.TrimSuffix(fields[5], ":53")
		owners[fields[1]+"@"+address] = struct{}{}
	}
	result := make([]string, 0, len(owners))
	for owner := range owners {
		result = append(result, owner)
	}
	sort.Strings(result)
	return result
}
