resource "opnsense_bind_acl" "internal" {
  name     = "internal-networks"
  networks = ["10.0.0.0/8", "2001:db8:100::/48"]
}
