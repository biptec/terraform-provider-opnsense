# This resource must be the only Terraform owner of:
# - opnsense_bind_settings.enabled;
# - opnsense_unbound_service.enabled;
# - opnsense_dnsmasq_settings.dns_port.
# Omit those attributes from other managed singleton resources.

resource "opnsense_bind_settings" "resolver" {
  # Leave enabled unmanaged. The cutover resource is the single owner of the
  # active DNS service flag.
  disable_ipv6 = true
  listen_ipv4  = ["10.53.0.2", "198.51.100.53"]
  listen_ipv6  = ["::1"]
  port         = 53
}

resource "opnsense_dns_service_cutover" "resolver" {
  target        = "bind"
  allow_cutover = true

  depends_on = [
    opnsense_bind_settings.resolver,
    # Add BIND views, zones, records, and TSIG resources here.
  ]
}
