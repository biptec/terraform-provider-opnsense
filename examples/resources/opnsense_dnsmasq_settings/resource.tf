resource "opnsense_dnsmasq_settings" "dns_listener" {
  enabled                  = true
  interfaces               = ["lan"]
  strict_interface_binding = true

  # Keep dnsmasq available for DHCP/RA while BIND owns port 53.
  dns_port = 0
}
