package routes_test

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

func testRouteGateway() string {
	if gateway := os.Getenv("OPNSENSE_TEST_ROUTE_GATEWAY"); gateway != "" {
		return gateway
	}
	return "WAN_DHCP"
}

func testRouteGatewayV6() string {
	if gateway := os.Getenv("OPNSENSE_TEST_ROUTE_GATEWAY_V6"); gateway != "" {
		return gateway
	}
	return "WAN_DHCP6"
}

func TestAccRouteResource(t *testing.T) {
	initialNetwork := "198.51.100.200/32"
	updatedNetwork := "198.51.100.201/32"
	gatewayIP := os.Getenv("OPNSENSE_TEST_ROUTE_GATEWAY_IP")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			checkActiveIPv4HostRoute(initialNetwork, gatewayIP, false),
			checkActiveIPv4HostRoute(updatedNetwork, gatewayIP, false),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccRouteResourceConfig(testRouteGateway(), initialNetwork),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_route.test", "enabled", "true"),
					resource.TestCheckResourceAttr("opnsense_route.test", "gateway", testRouteGateway()),
					resource.TestCheckResourceAttr("opnsense_route.test", "network", initialNetwork),
					resource.TestCheckResourceAttrSet("opnsense_route.test", "id"),
					checkActiveIPv4HostRoute(initialNetwork, gatewayIP, true),
				),
			},
			{
				ResourceName:      "opnsense_route.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRouteResourceConfig(testRouteGateway(), updatedNetwork),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_route.test", "network", updatedNetwork),
					checkActiveIPv4HostRoute(initialNetwork, gatewayIP, false),
					checkActiveIPv4HostRoute(updatedNetwork, gatewayIP, true),
				),
			},
		},
	})
}

func checkActiveIPv4HostRoute(network, gateway string, expected bool) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if acctest.QGASocket() == "" {
			return nil
		}
		if gateway == "" {
			return fmt.Errorf("OPNSENSE_TEST_ROUTE_GATEWAY_IP must be set when QEMU runtime checks are enabled")
		}
		destination := strings.TrimSuffix(network, "/32")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		var lastTable string
		for ctx.Err() == nil {
			table, err := acctest.QGAGuestExec(ctx, "/usr/bin/netstat", "-rn", "-f", "inet")
			if err == nil {
				lastTable = table
				found := false
				for _, line := range strings.Split(table, "\n") {
					fields := strings.Fields(line)
					if len(fields) >= 2 && fields[0] == destination && fields[1] == gateway {
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
		return fmt.Errorf("active host route %s via %s expected=%t; routing table: %s", network, gateway, expected, strings.TrimSpace(lastTable))
	}
}

func TestAccRouteResource_Disabled(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRouteResourceConfigDisabled(testRouteGateway(), "192.0.2.0/24"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_route.test", "enabled", "false"),
					resource.TestCheckResourceAttrSet("opnsense_route.test", "id"),
				),
			},
			{
				ResourceName:      "opnsense_route.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccRouteResource_WithDescription(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRouteResourceConfigWithDescription(testRouteGateway(), "192.0.2.0/24", "Test static route"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_route.test", "description", "Test static route"),
					resource.TestCheckResourceAttrSet("opnsense_route.test", "id"),
				),
			},
			{
				ResourceName:      "opnsense_route.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccRouteResource_IPv6(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRouteResourceConfig(testRouteGatewayV6(), "2001:db8::/32"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_route.test", "network", "2001:db8::/32"),
					resource.TestCheckResourceAttrSet("opnsense_route.test", "id"),
				),
			},
			{
				ResourceName:      "opnsense_route.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccRouteResourceConfig(gateway, network string) string {
	return fmt.Sprintf(`
resource "opnsense_route" "test" {
  gateway = %[1]q
  network = %[2]q
}
`, gateway, network)
}

func testAccRouteResourceConfigDisabled(gateway, network string) string {
	return fmt.Sprintf(`
resource "opnsense_route" "test" {
  enabled = false
  gateway = %[1]q
  network = %[2]q
}
`, gateway, network)
}

func testAccRouteResourceConfigWithDescription(gateway, network, description string) string {
	return fmt.Sprintf(`
resource "opnsense_route" "test" {
  gateway     = %[1]q
  network     = %[2]q
  description = %[3]q
}
`, gateway, network, description)
}
