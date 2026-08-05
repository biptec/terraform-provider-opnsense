resource "opnsense_bind_primary_domain" "public" {
  view_id        = opnsense_bind_view.public.id
  domain_name    = "example.net"
  update_key_ids = [opnsense_bind_tsig_key.acme.id]
  update_policy  = "zonesub_txt"
  dnssec         = true

  ttl          = 300
  refresh      = 3600
  retry        = 600
  expire       = 1209600
  negative_ttl = 300
  mail_admin   = "hostmaster@example.net"
  dns_server   = "ns1.example.net"
}
