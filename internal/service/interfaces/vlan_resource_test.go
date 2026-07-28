package interfaces_test

import (
	"fmt"
	"testing"

	"github.com/biptec/terraform-provider-opnsense/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccInterfacesVlanResource(t *testing.T) {
	requireInterfaceLab(t)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVlanResourceConfig(100, "VLAN 100 test", 4, "802.1q", spareDevice1(), "vlan01"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_interfaces_vlan.test", "id"),
					resource.TestCheckResourceAttr("opnsense_interfaces_vlan.test", "tag", "100"),
					resource.TestCheckResourceAttr("opnsense_interfaces_vlan.test", "description", "VLAN 100 test"),
					resource.TestCheckResourceAttr("opnsense_interfaces_vlan.test", "priority", "4"),
					resource.TestCheckResourceAttr("opnsense_interfaces_vlan.test", "protocol", "802.1q"),
					resource.TestCheckResourceAttr("opnsense_interfaces_vlan.test", "parent", spareDevice1()),
					resource.TestCheckResourceAttr("opnsense_interfaces_vlan.test", "device", "vlan01"),
				),
			},
			{
				ResourceName:      "opnsense_interfaces_vlan.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVlanResourceConfig(100, "Updated VLAN 100", 6, "802.1ad", spareDevice1(), "vlan01"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_interfaces_vlan.test", "description", "Updated VLAN 100"),
					resource.TestCheckResourceAttr("opnsense_interfaces_vlan.test", "priority", "6"),
					resource.TestCheckResourceAttr("opnsense_interfaces_vlan.test", "protocol", "802.1ad"),
				),
			},
		},
	})
}

func TestAccInterfacesVlanResource_HighVlanId(t *testing.T) {
	requireInterfaceLab(t)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: testAccVlanResourceConfig(4093, "High VLAN ID test", 6, "802.1q", spareDevice1(), "vlan02"),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("opnsense_interfaces_vlan.test", "tag", "4093"),
				resource.TestCheckResourceAttr("opnsense_interfaces_vlan.test", "priority", "6"),
				resource.TestCheckResourceAttr("opnsense_interfaces_vlan.test", "protocol", "802.1q"),
			),
		}},
	})
}

func testAccVlanResourceConfig(tag int, description string, priority int, protocol, parent, device string) string {
	return fmt.Sprintf(`
resource "opnsense_interfaces_vlan" "test" {
  tag         = %[1]d
  description = %[2]q
  priority    = %[3]d
  protocol    = %[4]q
  parent      = %[5]q
  device      = %[6]q
}
`, tag, description, priority, protocol, parent, device)
}
