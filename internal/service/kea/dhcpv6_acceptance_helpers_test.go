package kea_test

import "os"

func testKeaDhcpv6Interface() string {
	if iface := os.Getenv("OPNSENSE_KEA_DHCPV6_IFACE"); iface != "" {
		return iface
	}
	return "wan"
}
