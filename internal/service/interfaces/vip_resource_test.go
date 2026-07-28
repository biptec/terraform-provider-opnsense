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
