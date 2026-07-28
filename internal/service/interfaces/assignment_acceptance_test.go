package interfaces_test

import (
	"fmt"
	"testing"

	"github.com/biptec/terraform-provider-opnsense/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccInterfacesAssignmentResource(t *testing.T) {
	requireInterfaceLab(t)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAssignmentConfig("Test spare interface", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_interfaces_assignment.test", "id"),
					resource.TestCheckResourceAttrSet("opnsense_interfaces_assignment.test", "name"),
					resource.TestCheckResourceAttr("opnsense_interfaces_assignment.test", "device", spareDevice1()),
					resource.TestCheckResourceAttr("opnsense_interfaces_assignment.test", "description", "Test spare interface"),
					resource.TestCheckResourceAttr("opnsense_interfaces_assignment.test", "ipv4.mode", "none"),
					resource.TestCheckResourceAttr("opnsense_interfaces_assignment.test", "ipv6.mode", "none"),
				),
			},
			{
				ResourceName:            "opnsense_interfaces_assignment.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_readdress"},
			},
			{
				Config: testAccAssignmentConfig("Updated spare interface", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_interfaces_assignment.test", "description", "Updated spare interface"),
					resource.TestCheckResourceAttr("opnsense_interfaces_assignment.test", "block_private", "true"),
				),
			},
		},
	})
}

func testAccAssignmentConfig(description string, blockPrivate bool) string {
	return fmt.Sprintf(`
resource "opnsense_interfaces_assignment" "test" {
  device          = %[1]q
  description     = %[2]q
  enabled         = true
  block_private   = %[3]t
  allow_readdress = true

  ipv4 = {
    mode = "none"
  }

  ipv6 = {
    mode = "none"
  }
}
`, spareDevice1(), description, blockPrivate)
}

func TestAccInterfacesSettingsResource(t *testing.T) {
	requireInterfaceLab(t)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInterfaceSettingsConfig(11),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_interfaces_settings.test", "id", "interfaces_settings"),
					resource.TestCheckResourceAttr("opnsense_interfaces_settings.test", "dhcp6_ra_timeout", "11"),
				),
			},
			{
				ResourceName:      "opnsense_interfaces_settings.test",
				ImportState:       true,
				ImportStateId:     "interfaces_settings",
				ImportStateVerify: true,
			},
			{
				Config: testAccInterfaceSettingsConfig(10),
				Check:  resource.TestCheckResourceAttr("opnsense_interfaces_settings.test", "dhcp6_ra_timeout", "10"),
			},
		},
	})
}

func testAccInterfaceSettingsConfig(raTimeout int) string {
	return fmt.Sprintf(`
resource "opnsense_interfaces_settings" "test" {
  disable_checksum_offloading      = true
  disable_segmentation_offloading  = true
  disable_large_receive_offloading = true
  vlan_hardware_filtering          = "2"
  disable_ipv6                     = false
  dhcp6_no_release                 = false
  dhcp6_debug                      = false
  dhcp6_ra_timeout                 = %[1]d
}
`, raTimeout)
}
