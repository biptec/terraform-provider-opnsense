package bind_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/biptec/opnsense-go/pkg/api"
	apibind "github.com/biptec/opnsense-go/pkg/bind"
	"github.com/biptec/terraform-provider-opnsense/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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
				Config:             testAccBindSettingsConfig(),
				ResourceName:       "opnsense_bind_settings.test",
				ImportState:        true,
				ImportStateId:      "bind_settings",
				ImportStatePersist: true,
			},
			{
				Config: testAccBindAuthoritativeConfig(false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_bind_settings.test", "enabled", "true"),
					resource.TestCheckResourceAttr("opnsense_bind_settings.test", "port", "53530"),
					resource.TestCheckResourceAttrSet("opnsense_bind_acl.test", "id"),
					resource.TestCheckResourceAttr("opnsense_bind_acl.test", "networks.#", "1"),
					resource.TestCheckResourceAttrSet("opnsense_bind_view.test", "id"),
					resource.TestCheckResourceAttr("opnsense_bind_view.test", "sequence", "20"),
					resource.TestCheckResourceAttr("opnsense_bind_view.test", "match_destination_acl_ids.#", "1"),
					resource.TestCheckResourceAttrSet("opnsense_bind_tsig_key.test", "id"),
					resource.TestCheckResourceAttr("opnsense_bind_tsig_key.test", "algorithm", "hmac-sha256"),
					resource.TestCheckResourceAttrSet("opnsense_bind_primary_domain.test", "id"),
					resource.TestCheckResourceAttr("opnsense_bind_primary_domain.test", "dnssec", "true"),
					resource.TestCheckResourceAttrSet("opnsense_bind_primary_domain.test", "transfer_key_id"),
					resource.TestCheckResourceAttrSet("opnsense_bind_primary_domain_update_key.test", "id"),
					resource.TestCheckResourceAttr("opnsense_bind_primary_domain.test", "also_notify.#", "1"),
					resource.TestCheckTypeSetElemAttr("opnsense_bind_primary_domain.test", "also_notify.*", "192.0.2.54"),
					resource.TestCheckResourceAttrSet("opnsense_bind_record.ns_address", "id"),
					resource.TestCheckResourceAttrSet("opnsense_bind_record.ns", "id"),
					resource.TestCheckResourceAttr("data.opnsense_bind_primary_domain.test", "domain_name", "tfacc-bind.invalid"),
					resource.TestCheckResourceAttrSet("data.opnsense_bind_primary_domain.test", "transfer_key_id"),
					resource.TestCheckResourceAttr("data.opnsense_bind_primary_domain.test", "update_key_ids.#", "1"),
					resource.TestCheckTypeSetElemAttr("data.opnsense_bind_primary_domain.test", "also_notify.*", "192.0.2.54"),
					checkBindPrimaryTransferRuntime(),
				),
			},
			{ResourceName: "opnsense_bind_acl.test", ImportState: true, ImportStateVerify: true},
			{ResourceName: "opnsense_bind_view.test", ImportState: true, ImportStateVerify: true},
			{ResourceName: "opnsense_bind_tsig_key.test", ImportState: true, ImportStateVerify: true},
			{ResourceName: "opnsense_bind_primary_domain.test", ImportState: true, ImportStateVerify: true, ImportStateVerifyIgnore: []string{"serial", "update_key_ids"}},
			{ResourceName: "opnsense_bind_record.ns_address", ImportState: true, ImportStateVerify: true},
			{
				PreConfig: func() { waitForBindDNSSEC(t) },
				Config:    testAccBindAuthoritativeConfig(true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.opnsense_bind_dnssec_status.test", "zone", "tfacc-bind.invalid"),
					resource.TestCheckResourceAttr("data.opnsense_bind_dnssec_status.test", "secure", "true"),
					resource.TestCheckResourceAttr("data.opnsense_bind_dnssec_status.test", "inline_signing", "true"),
					resource.TestCheckResourceAttrSet("data.opnsense_bind_dnssec_status.test", "ds_records.0"),
					checkBindDNSSECKSK("data.opnsense_bind_dnssec_status.test"),
				),
			},
		},
	})
}

type bindDomainSearchRow struct {
	UUID       string `json:"uuid"`
	DomainName string `json:"domainname"`
}

func checkBindPrimaryTransferRuntime() resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if acctest.QGASocket() == "" {
			return nil
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		script := `set -eu
config=/usr/local/etc/namedb/named.conf
named-checkconf "$config"
awk '
/zone "tfacc-bind.invalid"/ { zone = 1 }
zone && /also-notify[[:space:]]*{/ { in_notify = 1; next }
in_notify && /192[.]0[.]2[.]54 key "_acme-challenge[.]tfacc-bind[.]invalid";/ { notify_ok = 1 }
in_notify && /^[[:space:]]*};/ { in_notify = 0 }
zone && /allow-transfer[[:space:]]*{/ { in_transfer = 1; next }
in_transfer && /key "_acme-challenge[.]tfacc-bind[.]invalid";/ { transfer_ok = 1 }
in_transfer && /^[[:space:]]*};/ { in_transfer = 0 }
END { exit !(notify_ok && transfer_ok) }
' "$config"`
		output, err := acctest.QGAGuestExec(ctx, "/bin/sh", "-c", script)
		if err != nil {
			return fmt.Errorf("BIND primary transfer configuration is not authenticated: %w; output: %s", err, output)
		}
		return nil
	}
}

func waitForBindDNSSEC(t *testing.T) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	client := api.NewClient(api.Options{
		Uri:           os.Getenv("OPNSENSE_URI"),
		APIKey:        os.Getenv("OPNSENSE_API_KEY"),
		APISecret:     os.Getenv("OPNSENSE_API_SECRET"),
		AllowInsecure: os.Getenv("OPNSENSE_ALLOW_INSECURE") == "true",
	})
	controller := &apibind.Controller{Api: client}
	var domainID string
	var lastStatus string

	for {
		if domainID == "" {
			result, err := api.Search[bindDomainSearchRow](client, ctx, apibind.PrimaryDomainOpts.Search)
			if err == nil {
				for _, domain := range result.Rows {
					if domain.DomainName == "tfacc-bind.invalid" {
						domainID = domain.UUID
						break
					}
				}
			} else {
				lastStatus = "search primary domain: " + err.Error()
			}
		}

		if domainID != "" {
			status, err := controller.DNSSECStatus(ctx, "tfacc-bind.invalid", domainID)
			if err == nil {
				kskReady := false
				for _, key := range status.Keys {
					if key.Role == "ksk" && key.KeyTag != "" {
						kskReady = true
						break
					}
				}
				if status.Secure && status.InlineSigning && len(status.DSRecords) > 0 && kskReady {
					return
				}
				lastStatus = fmt.Sprintf(
					"secure=%t inline_signing=%t ds_records=%d keys=%d backend_error=%q",
					status.Secure, status.InlineSigning, len(status.DSRecords), len(status.Keys), status.Error,
				)
			} else {
				lastStatus = "read DNSSEC status: " + err.Error()
			}
		}

		select {
		case <-ctx.Done():
			t.Fatalf("BIND DNSSEC did not become ready: %s", lastStatus)
		case <-time.After(time.Second):
		}
	}
}

func checkBindDNSSECKSK(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceState, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found", resourceName)
		}
		count, err := strconv.Atoi(resourceState.Primary.Attributes["keys.#"])
		if err != nil {
			return fmt.Errorf("invalid DNSSEC key count: %w", err)
		}
		for index := 0; index < count; index++ {
			prefix := fmt.Sprintf("keys.%d.", index)
			if resourceState.Primary.Attributes[prefix+"role"] == "ksk" && resourceState.Primary.Attributes[prefix+"key_tag"] != "" {
				return nil
			}
		}
		return fmt.Errorf("no DNSSEC KSK with key_tag found in %s", resourceName)
	}
}

func testAccBindSettingsConfig() string {
	return `
resource "opnsense_bind_settings" "test" {
  enabled     = true
  listen_ipv4 = ["127.0.0.1"]
  port        = 53530
}
`
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
resource "opnsense_bind_settings" "test" {
  enabled     = true
  listen_ipv4 = ["127.0.0.1"]
  port        = 53530
}

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

  depends_on = [opnsense_bind_settings.test]
}

resource "opnsense_bind_tsig_key" "test" {
  name      = "_acme-challenge.tfacc-bind.invalid"
  algorithm = "hmac-sha256"
  secret    = "dGVycmFmb3JtLWFjY2VwdGFuY2UtdGVzdC1zZWNyZXQ="
}

resource "opnsense_bind_primary_domain" "test" {
  view_id          = opnsense_bind_view.test.id
  domain_name      = "tfacc-bind.invalid"
  transfer_key_id  = opnsense_bind_tsig_key.test.id
  also_notify      = ["192.0.2.54"]
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

resource "opnsense_bind_primary_domain_update_key" "test" {
  domain_id     = opnsense_bind_primary_domain.test.id
  update_key_id = opnsense_bind_tsig_key.test.id
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

  depends_on = [opnsense_bind_primary_domain_update_key.test]
}
%s`, dnssec)
}
