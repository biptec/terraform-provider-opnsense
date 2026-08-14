resource "opnsense_bind_primary_domain_update_key" "acme" {
  domain_id     = opnsense_bind_primary_domain.public.id
  update_key_id = opnsense_bind_tsig_key.acme.id
}
