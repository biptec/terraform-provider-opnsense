package bind_test

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

func bindPreCheck(t *testing.T) {
	t.Helper()
	acctest.AccPreCheck(t)
	req, err := http.NewRequest(http.MethodGet, os.Getenv("OPNSENSE_URI")+"/api/bind/general/get", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.SetBasicAuth(os.Getenv("OPNSENSE_API_KEY"), os.Getenv("OPNSENSE_API_SECRET"))
	client := &http.Client{Timeout: 15 * time.Second, Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: os.Getenv("OPNSENSE_ALLOW_INSECURE") == "true"}}}
	resp, err := client.Do(req)
	if err != nil {
		t.Skipf("os-bind API is unavailable: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Skipf("os-bind is not installed or API access is unavailable: HTTP %d", resp.StatusCode)
	}
}

func TestAccBindSettingsDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { bindPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: `
data "opnsense_bind_settings" "test" {}
data "opnsense_bind_status" "test" {}
`,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("data.opnsense_bind_settings.test", "id", "bind_settings"),
				resource.TestCheckResourceAttrSet("data.opnsense_bind_settings.test", "port"),
				resource.TestCheckResourceAttr("data.opnsense_bind_status.test", "id", "bind"),
				resource.TestCheckResourceAttrSet("data.opnsense_bind_status.test", "status"),
			),
		}},
	})
}

func TestAccBindAuthoritativeResources(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { bindPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBindAuthoritativeConfig(false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_bind_acl.test", "id"),
					resource.TestCheckResourceAttr("opnsense_bind_acl.test", "networks.#", "1"),
					resource.TestCheckResourceAttrSet("opnsense_bind_view.test", "id"),
					resource.TestCheckResourceAttr("opnsense_bind_view.test", "sequence", "20"),
					resource.TestCheckResourceAttr("opnsense_bind_view.test", "match_destination_acl_ids.#", "1"),
					resource.TestCheckResourceAttrSet("opnsense_bind_tsig_key.test", "id"),
					resource.TestCheckResourceAttr("opnsense_bind_tsig_key.test", "algorithm", "hmac-sha256"),
					resource.TestCheckResourceAttrSet("opnsense_bind_primary_domain.test", "id"),
					resource.TestCheckResourceAttr("opnsense_bind_primary_domain.test", "dnssec", "true"),
					resource.TestCheckResourceAttrSet("opnsense_bind_record.ns_address", "id"),
					resource.TestCheckResourceAttrSet("opnsense_bind_record.ns", "id"),
					resource.TestCheckResourceAttr("data.opnsense_bind_primary_domain.test", "domain_name", "tfacc-bind.invalid"),
				),
			},
			{ResourceName: "opnsense_bind_acl.test", ImportState: true, ImportStateVerify: true},
			{ResourceName: "opnsense_bind_view.test", ImportState: true, ImportStateVerify: true},
			{ResourceName: "opnsense_bind_tsig_key.test", ImportState: true, ImportStateVerify: true},
			{ResourceName: "opnsense_bind_primary_domain.test", ImportState: true, ImportStateVerify: true},
			{ResourceName: "opnsense_bind_record.ns_address", ImportState: true, ImportStateVerify: true},
			{
				PreConfig: func() { time.Sleep(7 * time.Second) },
				Config:    testAccBindAuthoritativeConfig(true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.opnsense_bind_dnssec_status.test", "zone", "tfacc-bind.invalid"),
					resource.TestCheckResourceAttr("data.opnsense_bind_dnssec_status.test", "secure", "true"),
					resource.TestCheckResourceAttr("data.opnsense_bind_dnssec_status.test", "inline_signing", "true"),
					resource.TestCheckResourceAttrSet("data.opnsense_bind_dnssec_status.test", "ds_records.0"),
					resource.TestCheckResourceAttrSet("data.opnsense_bind_dnssec_status.test", "keys.0.key_tag"),
					resource.TestCheckResourceAttr("data.opnsense_bind_dnssec_status.test", "keys.0.role", "ksk"),
				),
			},
		},
	})
}

func testAccBindAuthoritativeConfig(withDNSSECDataSource bool) string {
	dnssec := ""
	if withDNSSECDataSource {
		dnssec = `
data "opnsense_bind_dnssec_status" "test" {
  domain_id = opnsense_bind_primary_domain.test.id
  zone      = opnsense_bind_primary_domain.test.domain_name
}
`
	}
	return fmt.Sprintf(`
resource "opnsense_bind_acl" "test" {
  name     = "tfacc-bind-internal"
  networks = ["198.51.100.0/24"]
}

resource "opnsense_bind_view" "test" {
  sequence             = 20
  name                 = "tfacc_bind_internal"
  match_client_acl_ids      = [opnsense_bind_acl.test.id]
  match_destination_acl_ids = [opnsense_bind_acl.test.id]
  allow_query_acl_ids       = [opnsense_bind_acl.test.id]
  recursion            = false
  dnssec_validation    = "auto"
}

resource "opnsense_bind_tsig_key" "test" {
  name      = "_acme-challenge.tfacc-bind.invalid"
  algorithm = "hmac-sha256"
  secret    = "dGVycmFmb3JtLWFjY2VwdGFuY2UtdGVzdC1zZWNyZXQ="
}

resource "opnsense_bind_primary_domain" "test" {
  view_id          = opnsense_bind_view.test.id
  domain_name      = "tfacc-bind.invalid"
  update_key_ids   = [opnsense_bind_tsig_key.test.id]
  update_policy    = "self_txt"
  dnssec           = true
  ttl              = 60
  refresh          = 300
  retry            = 300
  expire           = 86400
  negative_ttl     = 60
  mail_admin       = "hostmaster@tfacc-bind.invalid"
  dns_server       = "ns.tfacc-bind.invalid"
}

resource "opnsense_bind_record" "ns_address" {
  domain_id = opnsense_bind_primary_domain.test.id
  name      = "ns"
  type      = "A"
  value     = "192.0.2.53"
}

resource "opnsense_bind_record" "ns" {
  domain_id = opnsense_bind_primary_domain.test.id
  name      = "@"
  type      = "NS"
  value     = "ns.tfacc-bind.invalid."
}

data "opnsense_bind_primary_domain" "test" {
  id = opnsense_bind_primary_domain.test.id
}
%s`, dnssec)
}
