# Legacy UUID lookup.
data "opnsense_bind_primary_domain" "by_id" {
  id = "00000000-0000-4000-8000-000000000010"
}

# Preferred semantic lookup: BIND zone identity is view + domain name.
data "opnsense_bind_primary_domain" "internal" {
  domain_name = "example.net"
  view_name   = "internal"
}

# Optional low-level compatibility when a view UUID is already available.
data "opnsense_bind_primary_domain" "by_view_id" {
  domain_name = "example.net"
  view_id     = "00000000-0000-4000-8000-000000000001"
}
