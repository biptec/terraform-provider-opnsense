package firewall_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/biptec/terraform-provider-opnsense/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func firewallGroupPreCheck(t *testing.T) string {
	t.Helper()
	acctest.AccPreCheck(t)
	member := os.Getenv("OPNSENSE_TEST_FIREWALL_GROUP_MEMBER")
	if member == "" {
		member = os.Getenv("OPNSENSE_TEST_MANAGEMENT_INTERFACE")
	}
	if member == "" {
		t.Skip("OPNSENSE_TEST_FIREWALL_GROUP_MEMBER or OPNSENSE_TEST_MANAGEMENT_INTERFACE must be set for firewall group tests")
	}
	return member
}

func TestAccFirewallGroupResource(t *testing.T) {
	member := firewallGroupPreCheck(t)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { _ = firewallGroupPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallGroupConfig(member, "initial", 10),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_firewall_group.test", "name", "TFACC_GROUP"),
					resource.TestCheckResourceAttr("opnsense_firewall_group.test", "members.#", "1"),
					resource.TestCheckResourceAttr("opnsense_firewall_group.test", "sequence", "10"),
					resource.TestCheckResourceAttrSet("opnsense_firewall_group.test", "id"),
				),
			},
			{ResourceName: "opnsense_firewall_group.test", ImportState: true, ImportStateVerify: true},
			{
				Config: testAccFirewallGroupConfig(member, "updated", 20),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_firewall_group.test", "description", "updated"),
					resource.TestCheckResourceAttr("opnsense_firewall_group.test", "sequence", "20"),
				),
			},
		},
	})
}

func testAccFirewallGroupConfig(member, description string, sequence int) string {
	return fmt.Sprintf(`
resource "opnsense_firewall_group" "test" {
  name        = "TFACC_GROUP"
  members     = [%[1]q]
  description = %[2]q
  sequence    = %[3]d
}
`, member, description, sequence)
}

func nptPreCheck(t *testing.T) (string, string, string) {
	t.Helper()
	acctest.AccPreCheck(t)
	iface := os.Getenv("OPNSENSE_TEST_NPT_INTERFACE")
	if iface == "" {
		iface = os.Getenv("OPNSENSE_TEST_MANAGEMENT_INTERFACE")
	}
	source := os.Getenv("OPNSENSE_TEST_NPT_SOURCE")
	if source == "" {
		source = "fd00:3901::/48"
	}
	destination := os.Getenv("OPNSENSE_TEST_NPT_DESTINATION")
	if destination == "" {
		destination = "2001:db8:3901::/48"
	}
	if iface == "" {
		t.Skip("OPNSENSE_TEST_NPT_INTERFACE or OPNSENSE_TEST_MANAGEMENT_INTERFACE must be set for NPT tests")
	}
	return iface, source, destination
}

func TestAccFirewallNptResource(t *testing.T) {
	iface, source, destination := nptPreCheck(t)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { _, _, _ = nptPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallNptConfig(iface, source, destination, "initial", 100),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_firewall_npt.test", "enabled", "true"),
					resource.TestCheckResourceAttr("opnsense_firewall_npt.test", "source_net", source),
					resource.TestCheckResourceAttr("opnsense_firewall_npt.test", "destination_net", destination),
					resource.TestCheckResourceAttrSet("opnsense_firewall_npt.test", "id"),
				),
			},
			{ResourceName: "opnsense_firewall_npt.test", ImportState: true, ImportStateVerify: true},
			{
				Config: testAccFirewallNptConfig(iface, source, destination, "updated", 200),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_firewall_npt.test", "description", "updated"),
					resource.TestCheckResourceAttr("opnsense_firewall_npt.test", "sequence", "200"),
				),
			},
		},
	})
}

func testAccFirewallNptConfig(iface, source, destination, description string, sequence int) string {
	return fmt.Sprintf(`
resource "opnsense_firewall_npt" "test" {
  interface       = %[1]q
  source_net      = %[2]q
  destination_net = %[3]q
  description     = %[4]q
  sequence        = %[5]d
}
`, iface, source, destination, description, sequence)
}
