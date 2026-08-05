data "opnsense_bind_dnssec_status" "public" {
  domain_id = opnsense_bind_primary_domain.public.id
  zone      = opnsense_bind_primary_domain.public.domain_name
}
