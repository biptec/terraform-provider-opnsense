resource "opnsense_bind_record" "ns1_address" {
  domain_id = opnsense_bind_primary_domain.public.id
  name      = "ns1"
  type      = "A"
  value     = "192.0.2.53"
}

resource "opnsense_bind_record" "ns1" {
  domain_id = opnsense_bind_primary_domain.public.id
  name      = "@"
  type      = "NS"
  value     = "ns1.example.net."
}
