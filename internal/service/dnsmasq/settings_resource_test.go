package dnsmasq_test

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/biptec/terraform-provider-opnsense/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDnsmasqSettingsResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccDnsmasqSettingsConfig(53532),
				ResourceName:       "opnsense_dnsmasq_settings.test",
				ImportState:        true,
				ImportStateId:      "dnsmasq_settings",
				ImportStatePersist: true,
			},
			{
				Config: testAccDnsmasqSettingsConfig(53532),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_dnsmasq_settings.test", "enabled", "true"),
					resource.TestCheckResourceAttr("opnsense_dnsmasq_settings.test", "strict_interface_binding", "true"),
					resource.TestCheckResourceAttr("opnsense_dnsmasq_settings.test", "interfaces.#", "1"),
					resource.TestCheckTypeSetElemAttr("opnsense_dnsmasq_settings.test", "interfaces.*", "lan"),
					resource.TestCheckResourceAttr("opnsense_dnsmasq_settings.test", "dns_port", "53532"),
					resource.TestCheckResourceAttr("data.opnsense_dnsmasq_settings.test", "dns_port", "53532"),
					checkDnsmasqPort("tcp", 53532, true),
					checkDnsmasqPort("udp", 53532, true),
					checkDnsmasqPort("tcp", 53, false),
					checkDnsmasqPort("udp", 53, false),
				),
			},
			{
				Config: testAccDnsmasqSettingsConfig(0),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_dnsmasq_settings.test", "enabled", "true"),
					resource.TestCheckResourceAttr("opnsense_dnsmasq_settings.test", "dns_port", "0"),
					resource.TestCheckResourceAttr("data.opnsense_dnsmasq_settings.test", "dns_port", "0"),
					checkDnsmasqPort("tcp", 53532, false),
					checkDnsmasqPort("udp", 53532, false),
					checkDnsmasqPort("tcp", 53053, false),
					checkDnsmasqPort("udp", 53053, false),
					checkDnsmasqPort("tcp", 53, false),
					checkDnsmasqPort("udp", 53, false),
				),
			},
		},
	})
}
func testAccDnsmasqSettingsConfig(port int) string {
	return fmt.Sprintf(`
resource "opnsense_dnsmasq_settings" "test" {
  enabled                  = true
  interfaces               = ["lan"]
  strict_interface_binding = true
  dns_port                 = %d
}

data "opnsense_dnsmasq_settings" "test" {
  depends_on = [opnsense_dnsmasq_settings.test]
}
`, port)
}

func checkDnsmasqPort(protocol string, port int, expected bool) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if acctest.QGASocket() == "" {
			return nil
		}
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()
		pattern := regexp.MustCompile(`(?:^|\s)(\*|[0-9.]+):` + fmt.Sprint(port) + `(?:\s|$)`)
		var lastOutput string
		for ctx.Err() == nil {
			output, err := acctest.QGAGuestExec(ctx, "/usr/bin/sockstat", "-4", "-l", "-P", protocol)
			if err == nil {
				lastOutput = output
				listeners := make([]string, 0)
				for _, line := range strings.Split(output, "\n") {
					if !strings.Contains(line, "dnsmasq") {
						continue
					}
					match := pattern.FindStringSubmatch(line)
					if len(match) == 2 {
						listeners = append(listeners, match[1])
					}
				}
				found := len(listeners) > 0
				wildcard := false
				for _, listener := range listeners {
					wildcard = wildcard || listener == "*"
				}
				if found == expected && (!expected || !wildcard) {
					return nil
				}
			}
			time.Sleep(time.Second)
		}
		return fmt.Errorf(
			"dnsmasq %s listener on port %d expected=%t; sockstat: %s",
			protocol,
			port,
			expected,
			strings.TrimSpace(lastOutput),
		)
	}
}
