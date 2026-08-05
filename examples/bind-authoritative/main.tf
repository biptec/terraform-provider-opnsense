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

# Import this singleton before the first apply:
# tofu import opnsense_bind_settings.dns bind_settings
resource "opnsense_bind_settings" "dns" {
  enabled     = true
  listen_ipv4 = ["10.0.0.1", "192.0.2.53"]
  port        = 53
}

resource "opnsense_bind_acl" "internal_clients" {
  name     = "internal-clients"
  networks = ["10.0.0.0/8", "2001:db8:100::/48"]
}

resource "opnsense_bind_acl" "internal_dns_addresses" {
  name     = "internal-dns-addresses"
  networks = ["10.0.0.1/32"]
}

resource "opnsense_bind_acl" "public_dns_addresses" {
  name     = "public-dns-addresses"
  networks = ["192.0.2.53/32"]
}

resource "opnsense_bind_view" "internal" {
  sequence                  = 10
  name                      = "internal"
  match_client_acl_ids      = [opnsense_bind_acl.internal_clients.id]
  match_destination_acl_ids = [opnsense_bind_acl.internal_dns_addresses.id]
  recursion                 = true
  allow_recursion_acl_ids   = [opnsense_bind_acl.internal_clients.id]
  allow_query_acl_ids       = [opnsense_bind_acl.internal_clients.id]
  dnssec_validation         = "auto"

  depends_on = [opnsense_bind_settings.dns]
}

resource "opnsense_bind_view" "public" {
  sequence                  = 100
  name                      = "public"
  match_any                 = true
  match_destination_acl_ids = [opnsense_bind_acl.public_dns_addresses.id]
  allow_query_any           = true
  recursion                 = false

  depends_on = [opnsense_bind_settings.dns]
}

# With self_txt, the key name is also the only TXT owner it may update.
resource "opnsense_bind_tsig_key" "acme" {
  name      = "_acme-challenge.example.net"
  algorithm = "hmac-sha256"
  secret    = var.acme_tsig_secret
}

resource "opnsense_bind_primary_domain" "public" {
  view_id        = opnsense_bind_view.public.id
  domain_name    = "example.net"
  update_key_ids = [opnsense_bind_tsig_key.acme.id]
  update_policy  = "self_txt"
  dnssec         = true
  mail_admin     = "hostmaster@example.net"
  dns_server     = "ns1.example.net"
}

resource "opnsense_bind_record" "public_ns1_address" {
  domain_id = opnsense_bind_primary_domain.public.id
  name      = "ns1"
  type      = "A"
  value     = "192.0.2.53"
}

resource "opnsense_bind_record" "public_ns1" {
  domain_id = opnsense_bind_primary_domain.public.id
  name      = "@"
  type      = "NS"
  value     = "ns1.example.net."
}

# The same zone has a private version only in the internal view.
resource "opnsense_bind_primary_domain" "internal" {
  view_id     = opnsense_bind_view.internal.id
  domain_name = "example.net"
  dnssec      = false
  mail_admin  = "hostmaster@example.net"
  dns_server  = "ns1.example.net"
}

resource "opnsense_bind_record" "internal_ns1_address" {
  domain_id = opnsense_bind_primary_domain.internal.id
  name      = "ns1"
  type      = "A"
  value     = "10.0.0.1"
}

resource "opnsense_bind_record" "internal_ns1" {
  domain_id = opnsense_bind_primary_domain.internal.id
  name      = "@"
  type      = "NS"
  value     = "ns1.example.net."
}

resource "opnsense_bind_record" "internal_git" {
  domain_id = opnsense_bind_primary_domain.internal.id
  name      = "git"
  type      = "A"
  value     = "10.20.0.10"
}

data "opnsense_bind_dnssec_status" "public" {
  domain_id = opnsense_bind_primary_domain.public.id
  zone      = opnsense_bind_primary_domain.public.domain_name
}

output "registrar_ds_records" {
  value = data.opnsense_bind_dnssec_status.public.ds_records
}
