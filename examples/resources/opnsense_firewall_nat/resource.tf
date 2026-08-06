resource "opnsense_firewall_nat" "example_one" {
  disable_nat = true
  sequence    = 10
  interface   = "wan"
  protocol  = "any"

  source = {
    net = "198.51.100.112/29"
  }

  log         = true
  description = "Do not NAT routed public subnet"
  depends_on  = [opnsense_firewall_nat_settings.outbound]
}

// Send internal networks through a dedicated public WAN IP alias.
resource "opnsense_firewall_nat_settings" "outbound" {
  mode = "hybrid"
}

resource "opnsense_firewall_alias" "internal_networks" {
  name = "INTERNAL_NETWORKS"
  type = "network"
  content = [
    "10.20.0.0/16",
    "10.30.0.0/16",
  ]
  description = "Networks using the dedicated egress IP"
}

resource "opnsense_interfaces_vip" "egress" {
  mode        = "ipalias"
  interface   = "wan"
  network     = "203.0.113.10/32"
  description = "Dedicated outbound NAT address"
}

resource "opnsense_firewall_nat" "dedicated_egress" {
  sequence  = 100
  interface = "wan"
  protocol  = "any"

  source = {
    net = opnsense_firewall_alias.internal_networks.name
  }

  target = {
    ip = trimsuffix(opnsense_interfaces_vip.egress.network, "/32")
  }

  description = "Internal networks through dedicated egress IP"
  depends_on  = [opnsense_firewall_nat_settings.outbound]
}

resource "opnsense_firewall_nat" "example_two" {
  enabled = false

  interface = "wan"
  protocol  = "TCP"

  source = {
    net = "wan" # This is equiv. to WAN Net
  }

  destination = {
    net  = "10.8.0.1"
    port = "443"
  }

  target = {
    ip   = "wanip"
    port = "http"
  }

  log         = true
  description = "Example"
}

resource "opnsense_firewall_nat" "example_three" {
  interface = "wan"
  protocol  = "TCP"

  source = {
    net    = "192.168.0.0/16" # This is equiv. to WAN Net
    invert = true
  }

  destination = {
    net  = "examplealias"
    port = "80-443"
  }

  target = {
    ip   = "wanip"
    port = "443"
  }

  description = "Example"
}
