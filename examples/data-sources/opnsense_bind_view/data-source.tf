# Legacy UUID lookup.
data "opnsense_bind_view" "by_id" {
  id = "00000000-0000-4000-8000-000000000001"
}

# Preferred semantic lookup.
data "opnsense_bind_view" "internal" {
  name = "internal"
}
