terraform {
  required_providers {
    opnsense = {
      source = "biptec/opnsense"
    }
  }
}

provider "opnsense" {}

variable "acme_tsig_secret" {
  type      = string
  sensitive = true
}

resource "opnsense_bind_acl" "internal" {
  name     = "internal-networks"
  networks = ["10.0.0.0/8", "2001:db8:100::/48"]
}

resource "opnsense_bind_view" "internal" {
  sequence                = 10
  name                    = "internal"
  match_client_acl_ids    = [opnsense_bind_acl.internal.id]
  recursion               = true
  allow_recursion_acl_ids = [opnsense_bind_acl.internal.id]
  allow_query_acl_ids     = [opnsense_bind_acl.internal.id]
  dnssec_validation       = "auto"
}

resource "opnsense_bind_view" "public" {
  sequence        = 100
  name            = "public"
  match_any       = true
  allow_query_any = true
  recursion       = false
}

resource "opnsense_bind_tsig_key" "acme" {
  name      = "acme-dns01"
  algorithm = "hmac-sha256"
  secret    = var.acme_tsig_secret
}

resource "opnsense_bind_primary_domain" "public" {
  view_id        = opnsense_bind_view.public.id
  domain_name    = "example.net"
  update_key_ids = [opnsense_bind_tsig_key.acme.id]
  update_policy  = "zonesub_txt"
  dnssec         = true
  mail_admin     = "hostmaster@example.net"
  dns_server     = "ns1.example.net"
}

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

data "opnsense_bind_dnssec_status" "public" {
  domain_id = opnsense_bind_primary_domain.public.id
  zone      = opnsense_bind_primary_domain.public.domain_name
}

output "registrar_ds_records" {
  value = data.opnsense_bind_dnssec_status.public.ds_records
}
