package bind_test

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/biptec/opnsense-go/pkg/api"
	apibind "github.com/biptec/opnsense-go/pkg/bind"
	"github.com/biptec/terraform-provider-opnsense/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccBindPrimaryDomainTransferAttachmentLifecycle(t *testing.T) {
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
				Config: testAccBindTransferAttachmentConfig(false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_bind_primary_domain.transfer", "id"),
					checkBindTransferAttachmentRemote(false),
				),
			},
			{
				Config: testAccBindTransferAttachmentConfig(true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_bind_primary_domain_transfer.test", "id"),
					resource.TestCheckResourceAttrSet("opnsense_bind_primary_domain_transfer.test", "transfer_key_id"),
					resource.TestCheckTypeSetElemAttr("opnsense_bind_primary_domain_transfer.test", "also_notify.*", "192.0.2.54"),
					checkBindTransferAttachmentRemote(true),
					checkBindTransferAttachmentNamedConfig(true),
				),
			},
			{
				ResourceName:      "opnsense_bind_primary_domain_transfer.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBindTransferAttachmentConfig(false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_bind_primary_domain.transfer", "id"),
					checkBindTransferAttachmentRemote(false),
					checkBindTransferAttachmentNamedConfig(false),
				),
			},
			{
				Config: testAccBindTransferAttachmentConfig(true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_bind_primary_domain_transfer.test", "id"),
					checkBindTransferAttachmentRemote(true),
				),
			},
		},
	})
}

func checkBindTransferAttachmentRemote(attached bool) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		domainState, ok := state.RootModule().Resources["opnsense_bind_primary_domain.transfer"]
		if !ok {
			return fmt.Errorf("primary domain is missing from Terraform state")
		}
		domainID := domainState.Primary.Attributes["id"]
		if domainID == "" {
			return fmt.Errorf("primary domain id is empty")
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		client := api.NewClient(api.Options{
			Uri:           os.Getenv("OPNSENSE_URI"),
			APIKey:        os.Getenv("OPNSENSE_API_KEY"),
			APISecret:     os.Getenv("OPNSENSE_API_SECRET"),
			AllowInsecure: os.Getenv("OPNSENSE_ALLOW_INSECURE") == "true",
		})
		domain, err := (&apibind.Controller{Api: client}).GetPrimaryDomain(ctx, domainID)
		if err != nil {
			return fmt.Errorf("read primary domain after transfer lifecycle operation: %w", err)
		}
		key := strings.TrimSpace(domain.TransferKey.String())
		notify := make([]string, 0, len(domain.AlsoNotify))
		for _, value := range domain.AlsoNotify {
			value = strings.TrimSpace(value)
			if value != "" {
				notify = append(notify, value)
			}
		}
		slices.Sort(notify)
		if !attached {
			if key != "" || len(notify) != 0 {
				return fmt.Errorf("transfer attachment remains after detach: key=%q notify=%q", key, notify)
			}
			return nil
		}

		attachmentState, ok := state.RootModule().Resources["opnsense_bind_primary_domain_transfer.test"]
		if !ok {
			return fmt.Errorf("transfer attachment is missing from Terraform state")
		}
		expectedKey := attachmentState.Primary.Attributes["transfer_key_id"]
		if key != expectedKey || len(notify) != 1 || notify[0] != "192.0.2.54" {
			return fmt.Errorf("unexpected remote transfer attachment: key=%q notify=%v", key, notify)
		}
		return nil
	}
}

func checkBindTransferAttachmentNamedConfig(attached bool) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if acctest.QGASocket() == "" {
			return nil
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		expect := "absent"
		if attached {
			expect = "present"
		}
		script := fmt.Sprintf(`set -eu
config=/usr/local/etc/namedb/named.conf
named-checkconf "$config"
block=$(awk '/zone "tfacc-transfer.invalid"/{on=1} on{print} on && /^};/{exit}' "$config")
case %q in
  present) printf '%%s\n' "$block" | grep -F '192.0.2.54 key "tfacc-secondary-transfer.invalid"' >/dev/null ;;
  absent) ! printf '%%s\n' "$block" | grep -F 'tfacc-secondary-transfer.invalid' >/dev/null ;;
esac`, expect)
		output, err := acctest.QGAGuestExec(ctx, "/bin/sh", "-c", script)
		if err != nil {
			return fmt.Errorf("unexpected rendered BIND transfer attachment state: %w; output: %s", err, output)
		}
		return nil
	}
}

func testAccBindTransferAttachmentConfig(withAttachment bool) string {
	attachment := ""
	if withAttachment {
		attachment = `
resource "opnsense_bind_tsig_key" "transfer" {
  name      = "tfacc-secondary-transfer.invalid"
  algorithm = "hmac-sha256"
  secret    = "dGVycmFmb3JtLXNlY29uZGFyeS10cmFuc2Zlci10ZXN0"
}

resource "opnsense_bind_primary_domain_transfer" "test" {
  domain_id       = opnsense_bind_primary_domain.transfer.id
  transfer_key_id = opnsense_bind_tsig_key.transfer.id
  also_notify     = ["192.0.2.54"]
}
`
	}
	return fmt.Sprintf(`
resource "opnsense_bind_settings" "test" {
  enabled     = true
  listen_ipv4 = ["127.0.0.1"]
  port        = 53531
}

resource "opnsense_bind_view" "transfer" {
  sequence          = 25
  name              = "tfacc_transfer_lifecycle"
  match_any         = true
  recursion         = false
  allow_query_any   = true
  dnssec_validation = "auto"
  depends_on        = [opnsense_bind_settings.test]
}

resource "opnsense_bind_primary_domain" "transfer" {
  view_id         = opnsense_bind_view.transfer.id
  domain_name     = "tfacc-transfer.invalid"
  transfer_key_id = ""
  also_notify     = []
  dnssec          = false
  ttl             = 60
  refresh         = 300
  retry           = 300
  expire          = 86400
  negative_ttl    = 60
  mail_admin      = "hostmaster@tfacc-transfer.invalid"
  dns_server      = "ns.tfacc-transfer.invalid"

  lifecycle {
    ignore_changes = [transfer_key_id, also_notify]
  }
}

%s`, attachment)
}
