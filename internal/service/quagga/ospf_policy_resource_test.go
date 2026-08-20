package quagga_test

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/biptec/terraform-provider-opnsense/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func ospfPolicyPreCheck(t *testing.T) {
	t.Helper()
	acctest.AccPreCheck(t)
	req, err := http.NewRequest(http.MethodGet, os.Getenv("OPNSENSE_URI")+"/api/quagga/ospfsettings/get", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.SetBasicAuth(os.Getenv("OPNSENSE_API_KEY"), os.Getenv("OPNSENSE_API_SECRET"))
	client := &http.Client{Timeout: 15 * time.Second, Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: os.Getenv("OPNSENSE_ALLOW_INSECURE") == "true"}}}
	resp, err := client.Do(req)
	if err != nil {
		t.Skipf("os-frr API is unavailable: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Skipf("os-frr is not installed or API access is unavailable: HTTP %d", resp.StatusCode)
	}
}

func TestAccOSPFPolicyResources(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { ospfPolicyPreCheck(t) }, ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: testAccOSPFPolicyConfig("192.0.2.0/24", "initial"), Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("opnsense_quagga_ospf_prefixlist.test", "network", "192.0.2.0/24"),
				resource.TestCheckResourceAttr("opnsense_quagga_ospf_routemap.test", "prefix_lists.#", "1"),
				resource.TestCheckResourceAttr("opnsense_quagga_ospf_redistribution.test", "redistribute", "connected"),
				resource.TestCheckResourceAttr("data.opnsense_quagga_ospf_redistribution.test", "description", "initial"),
			)},
			{ResourceName: "opnsense_quagga_ospf_prefixlist.test", ImportState: true, ImportStateVerify: true},
			{ResourceName: "opnsense_quagga_ospf_routemap.test", ImportState: true, ImportStateVerify: true},
			{ResourceName: "opnsense_quagga_ospf_redistribution.test", ImportState: true, ImportStateVerify: true},
			{Config: testAccOSPFPolicyConfig("198.51.100.0/24", "updated"), Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("opnsense_quagga_ospf_prefixlist.test", "network", "198.51.100.0/24"),
				resource.TestCheckResourceAttr("opnsense_quagga_ospf_redistribution.test", "description", "updated"),
			)},
		},
	})
}

func TestAccOSPF6PolicyResources(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { ospfPolicyPreCheck(t) }, ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: testAccOSPF6PolicyConfig("2001:db8:100::/48", "initial"), Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("opnsense_quagga_ospf6_prefixlist.test", "network", "2001:db8:100::/48"),
				resource.TestCheckResourceAttr("opnsense_quagga_ospf6_routemap.test", "prefix_lists.#", "1"),
				resource.TestCheckResourceAttr("opnsense_quagga_ospf6_redistribution.test", "redistribute", "connected"),
				resource.TestCheckResourceAttr("data.opnsense_quagga_ospf6_redistribution.test", "description", "initial"),
			)},
			{ResourceName: "opnsense_quagga_ospf6_prefixlist.test", ImportState: true, ImportStateVerify: true},
			{ResourceName: "opnsense_quagga_ospf6_routemap.test", ImportState: true, ImportStateVerify: true},
			{ResourceName: "opnsense_quagga_ospf6_redistribution.test", ImportState: true, ImportStateVerify: true},
			{Config: testAccOSPF6PolicyConfig("2001:db8:200::/48", "updated"), Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("opnsense_quagga_ospf6_prefixlist.test", "network", "2001:db8:200::/48"),
				resource.TestCheckResourceAttr("opnsense_quagga_ospf6_redistribution.test", "description", "updated"),
			)},
		},
	})
}

func testAccOSPFPolicyConfig(network, description string) string {
	return fmt.Sprintf(`
resource "opnsense_quagga_ospf_prefixlist" "test" {
  name            = "tfacc-connected-v4"
  sequence_number = 10
  action          = "deny"
  network         = %[1]q
}
resource "opnsense_quagga_ospf_routemap" "test" {
  name          = "tfacc-connected-v4"
  route_map_id  = 10
  action        = "deny"
  prefix_lists  = [opnsense_quagga_ospf_prefixlist.test.id]
}
resource "opnsense_quagga_ospf_redistribution" "test" {
  description  = %[2]q
  redistribute = "connected"
  route_map    = opnsense_quagga_ospf_routemap.test.id
}
data "opnsense_quagga_ospf_redistribution" "test" { id = opnsense_quagga_ospf_redistribution.test.id }
`, network, description)
}

func testAccOSPF6PolicyConfig(network, description string) string {
	return fmt.Sprintf(`
resource "opnsense_quagga_ospf6_prefixlist" "test" {
  name            = "tfacc-connected-v6"
  sequence_number = 10
  action          = "deny"
  network         = %[1]q
}
resource "opnsense_quagga_ospf6_routemap" "test" {
  name          = "tfacc-connected-v6"
  route_map_id  = 10
  action        = "deny"
  prefix_lists  = [opnsense_quagga_ospf6_prefixlist.test.id]
}
resource "opnsense_quagga_ospf6_redistribution" "test" {
  description  = %[2]q
  redistribute = "connected"
  route_map    = opnsense_quagga_ospf6_routemap.test.id
}
data "opnsense_quagga_ospf6_redistribution" "test" { id = opnsense_quagga_ospf6_redistribution.test.id }
`, network, description)
}
