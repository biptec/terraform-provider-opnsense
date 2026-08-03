package caddy_test

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

func caddyPreCheck(t *testing.T) {
	t.Helper()
	acctest.AccPreCheck(t)
	req, err := http.NewRequest(http.MethodGet, os.Getenv("OPNSENSE_URI")+"/api/caddy/general/get", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.SetBasicAuth(os.Getenv("OPNSENSE_API_KEY"), os.Getenv("OPNSENSE_API_SECRET"))
	client := &http.Client{Timeout: 15 * time.Second, Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: os.Getenv("OPNSENSE_ALLOW_INSECURE") == "true"}}}
	resp, err := client.Do(req)
	if err != nil {
		t.Skipf("os-caddy API is unavailable: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Skipf("os-caddy is not installed or API access is unavailable: HTTP %d", resp.StatusCode)
	}
}

func TestAccCaddyStatusAndSettingsDataSources(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { caddyPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: `
data "opnsense_caddy_settings" "test" {}
data "opnsense_caddy_status" "test" {}
`,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("data.opnsense_caddy_settings.test", "id", "caddy_settings"),
				resource.TestCheckResourceAttrSet("data.opnsense_caddy_settings.test", "http_port"),
				resource.TestCheckResourceAttr("data.opnsense_caddy_status.test", "id", "caddy"),
				resource.TestCheckResourceAttrSet("data.opnsense_caddy_status.test", "status"),
			),
		}},
	})
}

func TestAccCaddyReverseProxyResources(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { caddyPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCaddyReverseProxyConfig("initial", 8080),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_caddy_access_list.test", "id"),
					resource.TestCheckResourceAttr("opnsense_caddy_access_list.test", "client_ips.#", "2"),
					resource.TestCheckResourceAttrSet("opnsense_caddy_domain.test", "id"),
					resource.TestCheckResourceAttr("opnsense_caddy_domain.test", "protocol", "http"),
					resource.TestCheckResourceAttr("opnsense_caddy_domain.test", "certificate_mode", "none"),
					resource.TestCheckResourceAttrSet("opnsense_caddy_handler.test", "id"),
					resource.TestCheckResourceAttr("opnsense_caddy_handler.test", "upstream_port", "8080"),
					resource.TestCheckResourceAttr("data.opnsense_caddy_domain.test", "domain", "tfacc-http.invalid"),
				),
			},
			{ResourceName: "opnsense_caddy_access_list.test", ImportState: true, ImportStateVerify: true},
			{ResourceName: "opnsense_caddy_domain.test", ImportState: true, ImportStateVerify: true, ImportStateVerifyIgnore: []string{"generated_certificate_id"}},
			{ResourceName: "opnsense_caddy_handler.test", ImportState: true, ImportStateVerify: true},
			{
				Config: testAccCaddyReverseProxyConfig("updated", 8081),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_caddy_access_list.test", "description", "updated"),
					resource.TestCheckResourceAttr("opnsense_caddy_domain.test", "description", "updated"),
					resource.TestCheckResourceAttr("opnsense_caddy_handler.test", "description", "updated"),
					resource.TestCheckResourceAttr("opnsense_caddy_handler.test", "upstream_port", "8081"),
				),
			},
		},
	})
}

func testAccCaddyReverseProxyConfig(description string, port int) string {
	return fmt.Sprintf(`
resource "opnsense_caddy_access_list" "test" {
  name            = "tfacc-management"
  client_ips      = ["192.0.2.0/24", "198.51.100.10"]
  request_matcher = "client_ip"
  description     = %[1]q
}

resource "opnsense_caddy_domain" "test" {
  domain           = "tfacc-http.invalid"
  protocol         = "http"
  certificate_mode = "none"
  access_list_id   = opnsense_caddy_access_list.test.id
  description      = %[1]q
}

resource "opnsense_caddy_handler" "test" {
  domain_id         = opnsense_caddy_domain.test.id
  upstream_domains = ["127.0.0.1"]
  upstream_port     = %[2]d
  upstream_protocol = "http"
  description       = %[1]q
}

data "opnsense_caddy_domain" "test" {
  id = opnsense_caddy_domain.test.id
}
`, description, port)
}

func TestAccCaddyACMEDomain(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { caddyPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: `
resource "opnsense_caddy_domain" "test" {
  domain           = "tfacc-acme.invalid"
  protocol         = "https"
  certificate_mode = "acme"
  description      = "automatic public ACME"
}
`,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("opnsense_caddy_domain.test", "certificate_mode", "acme"),
				resource.TestCheckResourceAttr("opnsense_caddy_domain.test", "certificate_ref_id", ""),
			),
		}},
	})
}

func TestAccCaddyInternalDomain(t *testing.T) {
	caName := os.Getenv("OPNSENSE_TEST_CADDY_CA_NAME")
	if caName == "" {
		t.Skip("OPNSENSE_TEST_CADDY_CA_NAME is required for internal certificate acceptance testing")
	}
	var firstCertificateID string
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { caddyPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCaddyInternalDomainConfig(caName, "tfacc-internal.invalid", "internal"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_caddy_domain.test", "certificate_mode", "internal"),
					resource.TestCheckResourceAttr("opnsense_caddy_domain.test", "internal_certificate_lifetime_days", "3650"),
					resource.TestCheckResourceAttrSet("opnsense_caddy_domain.test", "certificate_ref_id"),
					resource.TestCheckResourceAttrWith("opnsense_caddy_domain.test", "generated_certificate_id", func(value string) error {
						firstCertificateID = value
						if value == "" {
							return fmt.Errorf("generated certificate ID is empty")
						}
						return nil
					}),
				),
			},
			{
				Config: testAccCaddyInternalDomainConfig(caName, "tfacc-internal-updated.invalid", "internal"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_caddy_domain.test", "domain", "tfacc-internal-updated.invalid"),
					resource.TestCheckResourceAttrWith("opnsense_caddy_domain.test", "generated_certificate_id", func(value string) error {
						if value == "" || value == firstCertificateID {
							return fmt.Errorf("expected replacement certificate, got %q", value)
						}
						return nil
					}),
				),
			},
			{
				Config: testAccCaddyInternalDomainConfig(caName, "tfacc-internal-updated.invalid", "acme"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_caddy_domain.test", "certificate_mode", "acme"),
					resource.TestCheckResourceAttr("opnsense_caddy_domain.test", "certificate_ref_id", ""),
					resource.TestCheckNoResourceAttr("opnsense_caddy_domain.test", "generated_certificate_id"),
				),
			},
		},
	})
}

func testAccCaddyInternalDomainConfig(caName, domain, mode string) string {
	internal := ""
	if mode == "internal" {
		internal = fmt.Sprintf(`
  internal_ca_name                    = %q
  internal_certificate_lifetime_days = 3650`, caName)
	}
	return fmt.Sprintf(`
resource "opnsense_caddy_domain" "test" {
  domain           = %q
  protocol         = "https"
  certificate_mode = %q
  description      = "dynamic internal certificate"%s
}
`, domain, mode, internal)
}

func TestAccCaddySettingsResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { caddyPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccCaddySettingsConfig(80, 443, "root"),
				ResourceName:       "opnsense_caddy_settings.test",
				ImportState:        true,
				ImportStateId:      "caddy_settings",
				ImportStateVerify:  false,
				ImportStatePersist: true,
			},
			{
				Config: testAccCaddySettingsConfig(8080, 8443, "www"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_caddy_settings.test", "http_port", "8080"),
					resource.TestCheckResourceAttr("opnsense_caddy_settings.test", "https_port", "8443"),
					resource.TestCheckResourceAttr("opnsense_caddy_settings.test", "run_as_user", "www"),
				),
			},
			{
				Config: testAccCaddySettingsConfig(80, 443, "root"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_caddy_settings.test", "http_port", "80"),
					resource.TestCheckResourceAttr("opnsense_caddy_settings.test", "https_port", "443"),
					resource.TestCheckResourceAttr("opnsense_caddy_settings.test", "run_as_user", "root"),
				),
			},
		},
	})
}

func testAccCaddySettingsConfig(httpPort, httpsPort int, runAs string) string {
	return fmt.Sprintf(`
resource "opnsense_caddy_settings" "test" {
  enabled               = false
  enable_layer4         = false
  http_port              = %d
  https_port             = %d
  acme_email             = ""
  auto_https             = ""
  run_as_user            = %q
  grace_period           = 10
  http_versions          = ["h1", "h2"]
  log_level              = ""
  plain_access_log       = false
  plain_access_log_keep  = 10
}
`, httpPort, httpsPort, runAs)
}

func TestAccCaddyCustomCertificateDomain(t *testing.T) {
	certificateRefID := os.Getenv("OPNSENSE_TEST_CADDY_CERT_REF_ID")
	if certificateRefID == "" {
		t.Skip("OPNSENSE_TEST_CADDY_CERT_REF_ID is required for custom certificate acceptance testing")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { caddyPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: fmt.Sprintf(`
resource "opnsense_caddy_domain" "test" {
  domain             = "tfacc-custom.invalid"
  protocol           = "https"
  certificate_mode   = "custom"
  certificate_ref_id = %q
  description        = "existing custom certificate"
}
`, certificateRefID),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("opnsense_caddy_domain.test", "certificate_mode", "custom"),
				resource.TestCheckResourceAttr("opnsense_caddy_domain.test", "certificate_ref_id", certificateRefID),
				resource.TestCheckNoResourceAttr("opnsense_caddy_domain.test", "generated_certificate_id"),
			),
		}},
	})
}
