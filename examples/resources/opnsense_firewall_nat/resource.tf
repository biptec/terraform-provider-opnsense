resource "opnsense_firewall_nat" "example_one" {
  disable_nat = true

  interface = "wan"
  protocol  = "any"

  source = {
    net = "198.51.100.112/29"
  }

  log         = true
  description = "Do not NAT routed public subnet"
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
    ip = "wanip"
    port = "http"
  }

  log         = true
  description = "Example"
}

resource "opnsense_firewall_nat" "example_three" {
  interface = "wan"
  protocol  = "TCP"

  source = {
    net = "192.168.0.0/16" # This is equiv. to WAN Net
    invert = true
  }

  destination = {
    net  = "examplealias"
    port = "80-443"
  }

  target = {
    ip = "wanip"
    port = "443"
  }

  description = "Example"
}
