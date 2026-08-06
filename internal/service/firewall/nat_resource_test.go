package firewall_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/biptec/terraform-provider-opnsense/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccFirewallNATSettingsResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallNATSettingsConfig("hybrid"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_firewall_nat_settings.test", "id", "firewall_nat_settings"),
					resource.TestCheckResourceAttr("opnsense_firewall_nat_settings.test", "mode", "hybrid"),
				),
			},
			{
				ResourceName:      "opnsense_firewall_nat_settings.test",
				ImportState:       true,
				ImportStateId:     "firewall_nat_settings",
				ImportStateVerify: true,
			},
			{
				Config: testAccFirewallNATSettingsConfig("automatic"),
				Check:  resource.TestCheckResourceAttr("opnsense_firewall_nat_settings.test", "mode", "automatic"),
			},
		},
	})
}

func TestAccFirewallNATNoNATResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallNATNoNATConfig("198.51.100.64/29", "Initial routed public subnet", 900000, "hybrid"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_firewall_nat.test", "disable_nat", "true"),
					resource.TestCheckResourceAttr("opnsense_firewall_nat.test", "interface", "wan"),
					resource.TestCheckResourceAttr("opnsense_firewall_nat.test", "protocol", "any"),
					resource.TestCheckResourceAttr("opnsense_firewall_nat.test", "source.net", "198.51.100.64/29"),
					resource.TestCheckNoResourceAttr("opnsense_firewall_nat.test", "target.ip"),
					resource.TestCheckResourceAttrSet("opnsense_firewall_nat.test", "id"),
					checkActiveNoNATRule("198.51.100.64/29", true),
				),
			},
			{
				ResourceName:      "opnsense_firewall_nat.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFirewallNATNoNATConfig("203.0.113.64/29", "Updated routed public subnet", 900001, "hybrid"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_firewall_nat.test", "source.net", "203.0.113.64/29"),
					resource.TestCheckResourceAttr("opnsense_firewall_nat.test", "sequence", "900001"),
					resource.TestCheckNoResourceAttr("opnsense_firewall_nat.test", "target.ip"),
					checkActiveNoNATRule("198.51.100.64/29", false),
					checkActiveNoNATRule("203.0.113.64/29", true),
				),
			},
			{
				Config: testAccFirewallNATNoNATConfig("203.0.113.64/29", "Updated routed public subnet", 900001, "automatic"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_firewall_nat_settings.test", "mode", "automatic"),
					resource.TestCheckResourceAttr("opnsense_firewall_nat.test", "source.net", "203.0.113.64/29"),
					checkActiveNoNATRule("203.0.113.64/29", false),
				),
			},
		},
	})
}

func checkActiveNoNATRule(source string, expected bool) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if acctest.QGASocket() == "" {
			return nil
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		expectedFragment := " from " + source + " to any"
		var lastRules string
		for ctx.Err() == nil {
			rules, err := acctest.QGAGuestExec(ctx, "/sbin/pfctl", "-sn")
			if err == nil {
				lastRules = rules
				found := false
				for _, line := range strings.Split(rules, "\n") {
					if strings.HasPrefix(line, "no nat ") && strings.Contains(line, expectedFragment) {
						found = true
						break
					}
				}
				if found == expected {
					return nil
				}
			}
			time.Sleep(time.Second)
		}
		return fmt.Errorf("active NO-NAT rule for %s expected=%t; PF rules: %s", source, expected, strings.TrimSpace(lastRules))
	}
}

func testAccFirewallNATSettingsConfig(mode string) string {
	return fmt.Sprintf(`
resource "opnsense_firewall_nat_settings" "test" {
  mode = %q
}
`, mode)
}

func testAccFirewallNATNoNATConfig(source, description string, sequence int, mode string) string {
	return fmt.Sprintf(`
resource "opnsense_firewall_nat_settings" "test" {
  mode = %[4]q
}

resource "opnsense_firewall_nat" "test" {
  disable_nat = true
  sequence    = %[3]d
  interface   = "wan"
  protocol    = "any"

  source = {
    net = %[1]q
  }

  description = %[2]q

  depends_on = [opnsense_firewall_nat_settings.test]
}
`, source, description, sequence, mode)
}
