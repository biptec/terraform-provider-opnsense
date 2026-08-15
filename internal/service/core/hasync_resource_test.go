package core_test

import (
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
