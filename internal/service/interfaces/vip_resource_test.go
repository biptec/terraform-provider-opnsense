package interfaces_test

import (
	"fmt"
	"testing"

	"github.com/biptec/terraform-provider-opnsense/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccInterfacesVipProxyArpResource(t *testing.T) {
	requireInterfaceLab(t)
	testAccVIPMode(t, "proxyarp", "10.0.2.220/32", "10.0.2.221/32")
}

func TestAccInterfacesVipIpAliasResource(t *testing.T) {
	requireInterfaceLab(t)
	testAccVIPMode(t, "ipalias", "10.0.2.222/32", "10.0.2.223/32")
}

func testAccVIPMode(t *testing.T, mode, initialNetwork, updatedNetwork string) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVipResourceConfig(mode, "VIP test", managementInterface(), initialNetwork),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_interfaces_vip.test", "mode", mode),
					resource.TestCheckResourceAttr("opnsense_interfaces_vip.test", "interface", managementInterface()),
					resource.TestCheckResourceAttr("opnsense_interfaces_vip.test", "network", initialNetwork),
					resource.TestCheckResourceAttrSet("opnsense_interfaces_vip.test", "id"),
				),
			},
			{
				ResourceName:      "opnsense_interfaces_vip.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVipResourceConfig(mode, "Updated VIP test", managementInterface(), updatedNetwork),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_interfaces_vip.test", "description", "Updated VIP test"),
					resource.TestCheckResourceAttr("opnsense_interfaces_vip.test", "network", updatedNetwork),
				),
			},
		},
	})
}

func testAccVipResourceConfig(mode, description, interfaceName, network string) string {
	return fmt.Sprintf(`
resource "opnsense_interfaces_vip" "test" {
  mode        = %[1]q
  description = %[2]q
  interface   = %[3]q
  network     = %[4]q
}
`, mode, description, interfaceName, network)
}

func TestAccInterfacesVipIPAliasSharedCARPVHID(t *testing.T) {
	requireInterfaceLab(t)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVipSharedCARPVHIDConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_interfaces_vip.carp", "mode", "carp"),
					resource.TestCheckResourceAttr("opnsense_interfaces_vip.carp", "vhid", "221"),
					resource.TestCheckResourceAttr("opnsense_interfaces_vip.carp", "password_version", "1"),
					resource.TestCheckResourceAttr("opnsense_interfaces_vip.carp", "password_configured", "true"),
					resource.TestCheckNoResourceAttr("opnsense_interfaces_vip.carp", "password"),
					resource.TestCheckResourceAttr("opnsense_interfaces_vip.alias", "mode", "ipalias"),
					resource.TestCheckResourceAttr("opnsense_interfaces_vip.alias", "vhid", "221"),
					resource.TestCheckResourceAttr("opnsense_interfaces_vip.alias", "network", "10.0.2.225/32"),
					resource.TestCheckResourceAttrSet("opnsense_interfaces_vip.alias", "id"),
				),
			},
			{
				ResourceName:      "opnsense_interfaces_vip.alias",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccVipSharedCARPVHIDConfig() string {
	return fmt.Sprintf(`
resource "opnsense_interfaces_vip" "carp" {
  mode               = "carp"
  description        = "TF acceptance CARP parent"
  interface          = %[1]q
  network            = "10.0.2.224/24"
  password           = "tfacc-shared-vhid"
  password_version   = 1
  vhid               = 221
  advertisement_base = 1
  advertisement_skew = 0
  no_sync            = true
}

resource "opnsense_interfaces_vip" "alias" {
  mode        = "ipalias"
  description = "TF acceptance shared CARP alias"
  interface   = %[1]q
  network     = "10.0.2.225/32"
  vhid        = 221
  no_sync     = true

  depends_on = [opnsense_interfaces_vip.carp]
}
`, managementInterface())
}
