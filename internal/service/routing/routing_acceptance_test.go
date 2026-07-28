package routing_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/biptec/terraform-provider-opnsense/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccGatewayStatusDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: `data "opnsense_routing_gateway_status" "test" {}`,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttrSet("data.opnsense_routing_gateway_status.test", "status"),
				resource.TestCheckResourceAttrSet("data.opnsense_routing_gateway_status.test", "items.#"),
			),
		}},
	})
}

func routingMutationPreCheck(t *testing.T) string {
	t.Helper()
	acctest.AccPreCheck(t)
	iface := os.Getenv("OPNSENSE_TEST_ROUTING_INTERFACE")
	if iface == "" {
		iface = os.Getenv("OPNSENSE_TEST_MANAGEMENT_INTERFACE")
	}
	if iface == "" {
		t.Skip("OPNSENSE_TEST_ROUTING_INTERFACE or OPNSENSE_TEST_MANAGEMENT_INTERFACE must be set for routing mutation tests")
	}
	return iface
}

func TestAccGatewayResource(t *testing.T) {
	iface := routingMutationPreCheck(t)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { _ = routingMutationPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig(iface, "initial", 250),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_routing_gateway.test", "enabled", "true"),
					resource.TestCheckResourceAttr("opnsense_routing_gateway.test", "name", "TFACC_GATEWAY"),
					resource.TestCheckResourceAttr("opnsense_routing_gateway.test", "gateway", "192.0.2.1"),
					resource.TestCheckResourceAttr("opnsense_routing_gateway.test", "far_gateway", "true"),
					resource.TestCheckResourceAttr("opnsense_routing_gateway.test", "priority", "250"),
					resource.TestCheckResourceAttrSet("opnsense_routing_gateway.test", "id"),
				),
			},
			{ResourceName: "opnsense_routing_gateway.test", ImportState: true, ImportStateVerify: true},
			{
				Config: testAccGatewayConfig(iface, "updated", 240),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_routing_gateway.test", "description", "updated"),
					resource.TestCheckResourceAttr("opnsense_routing_gateway.test", "priority", "240"),
				),
			},
		},
	})
}

func testAccGatewayConfig(iface, description string, priority int) string {
	return fmt.Sprintf(`
resource "opnsense_routing_gateway" "test" {
  name            = "TFACC_GATEWAY"
  description     = %[1]q
  interface       = %[2]q
  ip_protocol     = "inet"
  gateway         = "192.0.2.1"
  far_gateway     = true
  monitor_disable = true
  priority        = %[3]d
}
`, description, iface, priority)
}

func TestAccGatewayGroupResource(t *testing.T) {
	iface := routingMutationPreCheck(t)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { _ = routingMutationPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayGroupConfig(iface, "down"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_routing_gateway_group.test", "name", "TFACC_GROUP"),
					resource.TestCheckResourceAttr("opnsense_routing_gateway_group.test", "trigger", "down"),
					resource.TestCheckResourceAttr("opnsense_routing_gateway_group.test", "tier1.#", "1"),
					resource.TestCheckResourceAttrSet("opnsense_routing_gateway_group.test", "id"),
				),
			},
			{ResourceName: "opnsense_routing_gateway_group.test", ImportState: true, ImportStateVerify: true},
			{
				Config: testAccGatewayGroupConfig(iface, "downloss"),
				Check:  resource.TestCheckResourceAttr("opnsense_routing_gateway_group.test", "trigger", "downloss"),
			},
		},
	})
}

func testAccGatewayGroupConfig(iface, trigger string) string {
	return fmt.Sprintf(`
resource "opnsense_routing_gateway" "test" {
  name            = "TFACC_GROUP_GW"
  interface       = %[1]q
  ip_protocol     = "inet"
  gateway         = "192.0.2.2"
  far_gateway     = true
  monitor_disable = true
}

resource "opnsense_routing_gateway_group" "test" {
  name    = "TFACC_GROUP"
  tier1   = [opnsense_routing_gateway.test.name]
  trigger = %[2]q
}
`, iface, trigger)
}
