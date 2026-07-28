package interfaces_test

import (
	"fmt"
	"testing"

	"github.com/biptec/terraform-provider-opnsense/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccInterfacesVxlanResource(t *testing.T) {
	requireInterfaceLab(t)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVxlanConfig(100, "198.51.100.2", 4789),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.opnsense_interfaces_overview.management", "addr4"),
					resource.TestCheckResourceAttrSet("opnsense_interfaces_vxlan.test", "id"),
					resource.TestCheckResourceAttrSet("opnsense_interfaces_vxlan.test", "device_id"),
					resource.TestCheckResourceAttr("opnsense_interfaces_vxlan.test", "vni", "100"),
					resource.TestCheckResourceAttr("opnsense_interfaces_vxlan.test", "remote_address", "198.51.100.2"),
				),
			},
			{
				ResourceName:      "opnsense_interfaces_vxlan.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVxlanConfig(100, "198.51.100.3", 4790),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_interfaces_vxlan.test", "remote_address", "198.51.100.3"),
					resource.TestCheckResourceAttr("opnsense_interfaces_vxlan.test", "remote_port", "4790"),
				),
			},
		},
	})
}

func testAccVxlanConfig(vni int, remoteAddress string, remotePort int) string {
	return providerDataSourceConfig() + fmt.Sprintf(`
resource "opnsense_interfaces_vxlan" "test" {
  vni            = %[1]d
  source_address = split("/", data.opnsense_interfaces_overview.management.addr4)[0]
  remote_address = %[2]q
  remote_port    = %[3]d
}
`, vni, remoteAddress, remotePort)
}

func TestAccInterfacesBridgeResource(t *testing.T) {
	requireInterfaceLab(t)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBridgeConfig("Test bridge"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_interfaces_bridge.test", "id"),
					resource.TestCheckResourceAttrSet("opnsense_interfaces_bridge.test", "device"),
					resource.TestCheckResourceAttr("opnsense_interfaces_bridge.test", "members.#", "2"),
					resource.TestCheckResourceAttr("opnsense_interfaces_assignment.bridge_member_1", "device", spareDevice1()),
					resource.TestCheckResourceAttr("opnsense_interfaces_assignment.bridge_member_2", "device", spareDevice2()),
				),
			},
			{
				ResourceName:      "opnsense_interfaces_bridge.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBridgeConfig("Updated bridge"),
				Check:  resource.TestCheckResourceAttr("opnsense_interfaces_bridge.test", "description", "Updated bridge"),
			},
		},
	})
}

func testAccBridgeConfig(description string) string {
	return fmt.Sprintf(`
resource "opnsense_interfaces_assignment" "bridge_member_1" {
  device          = %[1]q
  description     = "Bridge member 1"
  enabled         = true
  allow_readdress = true
  ipv4 = { mode = "none" }
  ipv6 = { mode = "none" }
}

resource "opnsense_interfaces_assignment" "bridge_member_2" {
  device          = %[2]q
  description     = "Bridge member 2"
  enabled         = true
  allow_readdress = true
  ipv4 = { mode = "none" }
  ipv6 = { mode = "none" }
}

resource "opnsense_interfaces_bridge" "test" {
  members = [
    opnsense_interfaces_assignment.bridge_member_1.name,
    opnsense_interfaces_assignment.bridge_member_2.name,
  ]
  description = %[3]q
}
`, spareDevice1(), spareDevice2(), description)
}

func TestAccInterfacesLaggResource(t *testing.T) {
	requireInterfaceLab(t)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLaggConfig(spareDevice1(), "Test failover LAGG"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_interfaces_lagg.test", "id"),
					resource.TestCheckResourceAttrSet("opnsense_interfaces_lagg.test", "device"),
					resource.TestCheckResourceAttr("opnsense_interfaces_lagg.test", "protocol", "failover"),
					resource.TestCheckResourceAttr("opnsense_interfaces_lagg.test", "primary_member", spareDevice1()),
				),
			},
			{
				ResourceName:      "opnsense_interfaces_lagg.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaggConfig(spareDevice2(), "Updated failover LAGG"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_interfaces_lagg.test", "primary_member", spareDevice2()),
					resource.TestCheckResourceAttr("opnsense_interfaces_lagg.test", "description", "Updated failover LAGG"),
				),
			},
		},
	})
}

func testAccLaggConfig(primary, description string) string {
	return fmt.Sprintf(`
resource "opnsense_interfaces_lagg" "test" {
  members        = [%[1]q, %[2]q]
  protocol       = "failover"
  primary_member = %[3]q
  description    = %[4]q
}
`, spareDevice1(), spareDevice2(), primary, description)
}
