package interfaces_test

import (
	"fmt"
	"testing"

	"github.com/biptec/terraform-provider-opnsense/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccInterfacesGreResource(t *testing.T) {
	requireInterfaceLab(t)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGreConfig("Test GRE"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_interfaces_gre.test", "id"),
					resource.TestCheckResourceAttrSet("opnsense_interfaces_gre.test", "device"),
					resource.TestCheckResourceAttr("opnsense_interfaces_gre.test", "local_address", managementInterface()),
					resource.TestCheckResourceAttr("opnsense_interfaces_gre.test", "remote_address", "198.51.100.10"),
					resource.TestCheckResourceAttr("opnsense_interfaces_gre.test", "tunnel_remote_prefix", "30"),
				),
			},
			{ResourceName: "opnsense_interfaces_gre.test", ImportState: true, ImportStateVerify: true},
			{
				Config: testAccGreConfig("Updated GRE"),
				Check:  resource.TestCheckResourceAttr("opnsense_interfaces_gre.test", "description", "Updated GRE"),
			},
		},
	})
}

func testAccGreConfig(description string) string {
	return fmt.Sprintf(`
resource "opnsense_interfaces_gre" "test" {
  local_address         = %[1]q
  remote_address        = "198.51.100.10"
  tunnel_local_address  = "10.10.0.1"
  tunnel_remote_address = "10.10.0.2"
  tunnel_remote_prefix  = 30
  description           = %[2]q
}
`, managementInterface(), description)
}

func TestAccInterfacesGifResource(t *testing.T) {
	requireInterfaceLab(t)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGifConfig(false, "Test GIF"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_interfaces_gif.test", "id"),
					resource.TestCheckResourceAttrSet("opnsense_interfaces_gif.test", "device"),
					resource.TestCheckResourceAttr("opnsense_interfaces_gif.test", "local_address", managementInterface()),
					resource.TestCheckResourceAttr("opnsense_interfaces_gif.test", "ecn_friendly", "true"),
				),
			},
			{ResourceName: "opnsense_interfaces_gif.test", ImportState: true, ImportStateVerify: true},
			{
				Config: testAccGifConfig(true, "Updated GIF"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_interfaces_gif.test", "disable_ingress_filtering", "true"),
					resource.TestCheckResourceAttr("opnsense_interfaces_gif.test", "description", "Updated GIF"),
				),
			},
		},
	})
}

func testAccGifConfig(disableIngress bool, description string) string {
	return fmt.Sprintf(`
resource "opnsense_interfaces_gif" "test" {
  local_address             = %[1]q
  remote_address            = "198.51.100.11"
  tunnel_local_address      = "10.20.0.1"
  tunnel_remote_address     = "10.20.0.2"
  tunnel_remote_prefix      = 30
  ecn_friendly              = true
  disable_ingress_filtering = %[2]t
  description               = %[3]q
}
`, managementInterface(), disableIngress, description)
}

func TestAccInterfacesLoopbackResource(t *testing.T) {
	requireInterfaceLab(t)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `resource "opnsense_interfaces_loopback" "test" { description = "Routing loopback" }`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_interfaces_loopback.test", "id"),
					resource.TestCheckResourceAttrSet("opnsense_interfaces_loopback.test", "device_id"),
				),
			},
			{ResourceName: "opnsense_interfaces_loopback.test", ImportState: true, ImportStateVerify: true},
			{
				Config: `resource "opnsense_interfaces_loopback" "test" { description = "Updated routing loopback" }`,
				Check:  resource.TestCheckResourceAttr("opnsense_interfaces_loopback.test", "description", "Updated routing loopback"),
			},
		},
	})
}

func TestAccInterfacesNeighborResource(t *testing.T) {
	requireInterfaceLab(t)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNeighborConfig("52:54:00:aa:bb:01", "10.0.2.230", "Test neighbor"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_interfaces_neighbor.test", "id"),
					resource.TestCheckResourceAttr("opnsense_interfaces_neighbor.test", "mac_address", "52:54:00:aa:bb:01"),
					resource.TestCheckResourceAttr("opnsense_interfaces_neighbor.test", "ip_address", "10.0.2.230"),
				),
			},
			{ResourceName: "opnsense_interfaces_neighbor.test", ImportState: true, ImportStateVerify: true},
			{
				Config: testAccNeighborConfig("52:54:00:aa:bb:02", "10.0.2.231", "Updated neighbor"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_interfaces_neighbor.test", "mac_address", "52:54:00:aa:bb:02"),
					resource.TestCheckResourceAttr("opnsense_interfaces_neighbor.test", "ip_address", "10.0.2.231"),
				),
			},
		},
	})
}

func testAccNeighborConfig(mac, ip, description string) string {
	return fmt.Sprintf(`
resource "opnsense_interfaces_neighbor" "test" {
  mac_address = %[1]q
  ip_address  = %[2]q
  description = %[3]q
}
`, mac, ip, description)
}

func TestAccInterfacesDetailsDataSource(t *testing.T) {
	requireInterfaceLab(t)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: fmt.Sprintf(`data "opnsense_interfaces_details" "management" { interface = %q }`, managementInterface()),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("data.opnsense_interfaces_details.management", "interface", managementInterface()),
				resource.TestCheckResourceAttrSet("data.opnsense_interfaces_details.management", "details_json"),
			),
		}},
	})
}
