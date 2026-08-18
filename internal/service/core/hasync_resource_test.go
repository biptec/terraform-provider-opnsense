package core_test

import (
	"fmt"
	"testing"

	"github.com/biptec/terraform-provider-opnsense/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccCoreHasyncDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: `data "opnsense_core_hasync" "test" {}`,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("data.opnsense_core_hasync.test", "id", "core_hasync"),
				resource.TestCheckResourceAttrSet("data.opnsense_core_hasync.test", "pfsync_version"),
				resource.TestCheckResourceAttrSet("data.opnsense_core_hasync.test", "disable_preempt"),
				resource.TestCheckResourceAttrSet("data.opnsense_core_hasync.test", "pfsync_defer"),
			),
		}},
	})
}

func TestAccCoreHasyncResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCoreHasyncConfig(true, "tfacc-xmlrpc-v1", 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_core_hasync.test", "id", "core_hasync"),
					resource.TestCheckResourceAttr("opnsense_core_hasync.test", "disable_preempt", "true"),
					resource.TestCheckResourceAttr("opnsense_core_hasync.test", "pfsync_version", "1400"),
					resource.TestCheckResourceAttr("opnsense_core_hasync.test", "password_version", "1"),
					resource.TestCheckResourceAttr("opnsense_core_hasync.test", "password_configured", "true"),
					resource.TestCheckNoResourceAttr("opnsense_core_hasync.test", "password"),
				),
			},
			{
				Config: testAccCoreHasyncConfig(false, "tfacc-xmlrpc-v2", 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_core_hasync.test", "disable_preempt", "false"),
					resource.TestCheckResourceAttr("opnsense_core_hasync.test", "pfsync_version", "1400"),
					resource.TestCheckResourceAttr("opnsense_core_hasync.test", "password_version", "2"),
					resource.TestCheckResourceAttr("opnsense_core_hasync.test", "password_configured", "true"),
					resource.TestCheckNoResourceAttr("opnsense_core_hasync.test", "password"),
				),
			},
		},
	})
}

func testAccCoreHasyncConfig(disablePreempt bool, password string, passwordVersion int) string {
	value := "false"
	if disablePreempt {
		value = "true"
	}
	return `
resource "opnsense_core_hasync" "test" {
  disable_preempt  = ` + value + `
  pfsync_version   = "1400"
  pfsync_defer     = false
  password         = "` + password + `"
  password_version = ` + fmt.Sprintf("%d", passwordVersion) + `
}
`
}
