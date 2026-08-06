package firewall_test

import (
	"context"
	"fmt"
	"os"
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

func TestAccFirewallNATEgressAliasResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallNATEgressAliasConfig([]string{"10.200.0.0/16", "10.201.0.0/16"}, "198.51.100.230/32", "hybrid"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_firewall_nat_settings.egress", "mode", "hybrid"),
					resource.TestCheckResourceAttr("opnsense_firewall_alias.internal_networks", "name", "INTERNAL_EGRESS_NETWORKS"),
					resource.TestCheckResourceAttr("opnsense_firewall_alias.internal_networks", "content.#", "2"),
					resource.TestCheckTypeSetElemAttr("opnsense_firewall_alias.internal_networks", "content.*", "10.200.0.0/16"),
					resource.TestCheckTypeSetElemAttr("opnsense_firewall_alias.internal_networks", "content.*", "10.201.0.0/16"),
					resource.TestCheckResourceAttr("opnsense_interfaces_vip.egress", "mode", "ipalias"),
					resource.TestCheckResourceAttr("opnsense_interfaces_vip.egress", "interface", "wan"),
					resource.TestCheckResourceAttr("opnsense_interfaces_vip.egress", "network", "198.51.100.230/32"),
					resource.TestCheckResourceAttr("opnsense_firewall_nat.routed_public", "disable_nat", "true"),
					resource.TestCheckResourceAttr("opnsense_firewall_nat.routed_public", "sequence", "900000"),
					resource.TestCheckResourceAttr("opnsense_firewall_nat.routed_public", "source.net", "198.51.100.112/29"),
					resource.TestCheckResourceAttr("opnsense_firewall_nat.egress", "source.net", "INTERNAL_EGRESS_NETWORKS"),
					resource.TestCheckResourceAttr("opnsense_firewall_nat.egress", "target.ip", "198.51.100.230"),
					checkActiveIPv4Address("198.51.100.230", true),
					checkActiveAliasTable("INTERNAL_EGRESS_NETWORKS", []string{"10.200.0.0/16", "10.201.0.0/16"}),
					checkActiveNoNATRule("198.51.100.112/29", true),
					checkActiveSourceNATTableRule("INTERNAL_EGRESS_NETWORKS", "198.51.100.230", true),
					checkActiveNATRuleOrder("198.51.100.112/29", "INTERNAL_EGRESS_NETWORKS", "198.51.100.230"),
				),
			},
			{
				ResourceName:      "opnsense_firewall_nat.egress",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFirewallNATEgressAliasConfig([]string{"10.202.0.0/16", "10.203.0.0/16"}, "198.51.100.231/32", "hybrid"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_firewall_alias.internal_networks", "content.#", "2"),
					resource.TestCheckTypeSetElemAttr("opnsense_firewall_alias.internal_networks", "content.*", "10.202.0.0/16"),
					resource.TestCheckTypeSetElemAttr("opnsense_firewall_alias.internal_networks", "content.*", "10.203.0.0/16"),
					resource.TestCheckResourceAttr("opnsense_interfaces_vip.egress", "network", "198.51.100.231/32"),
					resource.TestCheckResourceAttr("opnsense_firewall_nat.egress", "source.net", "INTERNAL_EGRESS_NETWORKS"),
					resource.TestCheckResourceAttr("opnsense_firewall_nat.egress", "target.ip", "198.51.100.231"),
					checkActiveIPv4Address("198.51.100.230", false),
					checkActiveIPv4Address("198.51.100.231", true),
					checkActiveAliasTable("INTERNAL_EGRESS_NETWORKS", []string{"10.202.0.0/16", "10.203.0.0/16"}),
					checkActiveNoNATRule("198.51.100.112/29", true),
					checkActiveSourceNATTableRule("INTERNAL_EGRESS_NETWORKS", "198.51.100.230", false),
					checkActiveSourceNATTableRule("INTERNAL_EGRESS_NETWORKS", "198.51.100.231", true),
					checkActiveNATRuleOrder("198.51.100.112/29", "INTERNAL_EGRESS_NETWORKS", "198.51.100.231"),
				),
			},
			{
				Config: testAccFirewallNATEgressAliasConfig([]string{"10.202.0.0/16", "10.203.0.0/16"}, "198.51.100.231/32", "automatic"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_firewall_nat_settings.egress", "mode", "automatic"),
					checkActiveIPv4Address("198.51.100.231", true),
					checkActiveAliasTable("INTERNAL_EGRESS_NETWORKS", []string{"10.202.0.0/16", "10.203.0.0/16"}),
					checkActiveNoNATRule("198.51.100.112/29", false),
					checkActiveSourceNATTableRule("INTERNAL_EGRESS_NETWORKS", "198.51.100.231", false),
				),
			},
		},
	})
}

func checkActiveIPv4Address(address string, expected bool) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if acctest.QGASocket() == "" {
			return nil
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		device := os.Getenv("OPNSENSE_TEST_WAN_DEVICE")
		if device == "" {
			return fmt.Errorf("OPNSENSE_TEST_WAN_DEVICE must be set when QEMU runtime checks are enabled")
		}
		needle := "inet " + address + " "
		var lastOutput string
		for ctx.Err() == nil {
			output, err := acctest.QGAGuestExec(ctx, "/sbin/ifconfig", device)
			if err == nil {
				lastOutput = output
				if strings.Contains(output, needle) == expected {
					return nil
				}
			}
			time.Sleep(time.Second)
		}
		return fmt.Errorf("active IPv4 address %s expected=%t; ifconfig: %s", address, expected, strings.TrimSpace(lastOutput))
	}
}

func checkActiveAliasTable(name string, expected []string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if acctest.QGASocket() == "" {
			return nil
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		expectedSet := make(map[string]struct{}, len(expected))
		for _, value := range expected {
			expectedSet[value] = struct{}{}
		}
		var lastOutput string
		for ctx.Err() == nil {
			output, err := acctest.QGAGuestExec(ctx, "/sbin/pfctl", "-t", name, "-T", "show")
			if err == nil {
				lastOutput = output
				actualSet := make(map[string]struct{})
				for _, line := range strings.Split(output, "\n") {
					value := strings.TrimSpace(line)
					if value != "" {
						actualSet[value] = struct{}{}
					}
				}
				matches := len(actualSet) == len(expectedSet)
				if matches {
					for value := range expectedSet {
						if _, ok := actualSet[value]; !ok {
							matches = false
							break
						}
					}
				}
				if matches {
					return nil
				}
			}
			time.Sleep(time.Second)
		}
		return fmt.Errorf("active PF table %s does not match %v; table: %s", name, expected, strings.TrimSpace(lastOutput))
	}
}

func checkActiveSourceNATTableRule(alias, target string, expected bool) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if acctest.QGASocket() == "" {
			return nil
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		expectedSource := " from <" + alias + "> to any"
		expectedTarget := " -> " + target
		var lastRules string
		for ctx.Err() == nil {
			rules, err := acctest.QGAGuestExec(ctx, "/sbin/pfctl", "-sn")
			if err == nil {
				lastRules = rules
				found := false
				for _, line := range strings.Split(rules, "\n") {
					if strings.HasPrefix(line, "nat ") && strings.Contains(line, expectedSource) && strings.Contains(line, expectedTarget) {
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
		return fmt.Errorf("active source NAT rule from table %s to %s expected=%t; PF rules: %s", alias, target, expected, strings.TrimSpace(lastRules))
	}
}

func checkActiveNATRuleOrder(noNATSource, alias, target string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if acctest.QGASocket() == "" {
			return nil
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		noNATFragment := " from " + noNATSource + " to any"
		natSourceFragment := " from <" + alias + "> to any"
		natTargetFragment := " -> " + target
		var lastRules string
		for ctx.Err() == nil {
			rules, err := acctest.QGAGuestExec(ctx, "/sbin/pfctl", "-sn")
			if err == nil {
				lastRules = rules
				noNATIndex := -1
				natIndex := -1
				for index, line := range strings.Split(rules, "\n") {
					if noNATIndex < 0 && strings.HasPrefix(line, "no nat ") && strings.Contains(line, noNATFragment) {
						noNATIndex = index
					}
					if natIndex < 0 && strings.HasPrefix(line, "nat ") && strings.Contains(line, natSourceFragment) && strings.Contains(line, natTargetFragment) {
						natIndex = index
					}
				}
				if noNATIndex >= 0 && natIndex >= 0 && noNATIndex < natIndex {
					return nil
				}
			}
			time.Sleep(time.Second)
		}
		return fmt.Errorf("NO-NAT rule for %s must precede source NAT from table %s to %s; PF rules: %s", noNATSource, alias, target, strings.TrimSpace(lastRules))
	}
}

func testAccFirewallNATEgressAliasConfig(sources []string, vipNetwork, mode string) string {
	quotedSources := make([]string, 0, len(sources))
	for _, source := range sources {
		quotedSources = append(quotedSources, fmt.Sprintf("%q", source))
	}
	return fmt.Sprintf(`
resource "opnsense_firewall_nat_settings" "egress" {
  mode = %[3]q
}

resource "opnsense_firewall_alias" "internal_networks" {
  name = "INTERNAL_EGRESS_NETWORKS"
  type = "network"
  content = [
    %[1]s,
  ]
  description = "Dedicated egress NAT acceptance"
}

resource "opnsense_interfaces_vip" "egress" {
  mode        = "ipalias"
  description = "Dedicated egress NAT acceptance"
  interface   = "wan"
  network     = %[2]q
}


resource "opnsense_firewall_nat" "routed_public" {
  disable_nat = true
  sequence    = 900000
  interface   = "wan"
  protocol    = "any"

  source = {
    net = "198.51.100.112/29"
  }

  description = "Do not NAT routed public subnet"
  depends_on  = [opnsense_firewall_nat_settings.egress]
}

resource "opnsense_firewall_nat" "egress" {
  sequence  = 910000
  interface = "wan"
  protocol  = "any"

  source = {
    net = opnsense_firewall_alias.internal_networks.name
  }

  target = {
    ip = trimsuffix(opnsense_interfaces_vip.egress.network, "/32")
  }

  description = "Internal networks through dedicated egress IP"
  depends_on  = [opnsense_firewall_nat_settings.egress]
}
`, strings.Join(quotedSources, ",\n    "), vipNetwork, mode)
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
