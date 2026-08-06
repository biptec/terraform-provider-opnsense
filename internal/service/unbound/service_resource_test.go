package unbound_test

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/biptec/terraform-provider-opnsense/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccUnboundServiceResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccUnboundServiceConfig(true),
				ResourceName:       "opnsense_unbound_service.test",
				ImportState:        true,
				ImportStateId:      "unbound_service",
				ImportStatePersist: true,
			},
			{
				Config: testAccUnboundServiceConfig(false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_unbound_service.test", "enabled", "false"),
					resource.TestCheckResourceAttr("data.opnsense_unbound_service.test", "enabled", "false"),
					checkUnboundRuntime(false),
				),
			},
			{
				Config: testAccUnboundServiceConfig(true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_unbound_service.test", "enabled", "true"),
					resource.TestCheckResourceAttr("data.opnsense_unbound_service.test", "enabled", "true"),
					checkUnboundRuntime(true),
				),
			},
		},
	})
}

func TestAccUnboundServiceResource_CreateBlocked(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config:      testAccUnboundServiceConfig(false),
			ExpectError: regexp.MustCompile("Cannot Create Singleton Resource"),
		}},
	})
}

func testAccUnboundServiceConfig(enabled bool) string {
	return fmt.Sprintf(`
resource "opnsense_unbound_service" "test" {
  enabled = %t
}

data "opnsense_unbound_service" "test" {
  depends_on = [opnsense_unbound_service.test]
}
`, enabled)
}
func checkUnboundRuntime(expected bool) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if acctest.QGASocket() == "" {
			return nil
		}
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()
		portPattern := regexp.MustCompile(`(?:^|\s)(?:\*|[0-9.]+):53(?:\s|$)`)
		var lastOutput string
		for ctx.Err() == nil {
			output, err := acctest.QGAGuestExec(
				ctx,
				"/bin/sh",
				"-c",
				`printf 'pids='; pgrep -x unbound | tr '\n' ',' || true; printf '\nsockets\n'; sockstat -4 -l | awk '$1 == "unbound" {print}'`,
			)
			if err == nil {
				lastOutput = output
				lines := strings.Split(output, "\n")
				processRunning := len(lines) > 0 && strings.TrimPrefix(lines[0], "pids=") != ""
				dnsListener := portPattern.MatchString(output)
				if processRunning == expected && dnsListener == expected {
					return nil
				}
			}
			time.Sleep(time.Second)
		}
		return fmt.Errorf(
			"Unbound runtime expected enabled=%t; output: %s",
			expected,
			strings.TrimSpace(lastOutput),
		)
	}
}
