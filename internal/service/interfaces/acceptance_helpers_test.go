package interfaces_test

import (
	"fmt"
	"os"
	"testing"
)

func requireInterfaceLab(t *testing.T) {
	t.Helper()
	for _, name := range []string{
		"OPNSENSE_TEST_MANAGEMENT_DEVICE",
		"OPNSENSE_TEST_SPARE_DEVICE_1",
		"OPNSENSE_TEST_SPARE_DEVICE_2",
		"OPNSENSE_TEST_MANAGEMENT_INTERFACE",
	} {
		if os.Getenv(name) == "" {
			t.Skipf("%s must be set for interface acceptance tests", name)
		}
	}
}

func managementDevice() string    { return os.Getenv("OPNSENSE_TEST_MANAGEMENT_DEVICE") }
func spareDevice1() string        { return os.Getenv("OPNSENSE_TEST_SPARE_DEVICE_1") }
func spareDevice2() string        { return os.Getenv("OPNSENSE_TEST_SPARE_DEVICE_2") }
func managementInterface() string { return os.Getenv("OPNSENSE_TEST_MANAGEMENT_INTERFACE") }

func providerDataSourceConfig() string {
	return fmt.Sprintf(`
data "opnsense_interfaces_overview" "management" {
  device = %q
}
`, managementDevice())
}
