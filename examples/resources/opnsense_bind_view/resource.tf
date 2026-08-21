variable "dns_xfr_internal_secret" {
  type      = string
  sensitive = true
  ephemeral = true
}

variable "dns_xfr_public_secret" {
  type      = string
  sensitive = true
  ephemeral = true
}

resource "opnsense_bind_tsig_key" "internal_transfer" {
  name           = "dns-xfr-internal.example.net"
  algorithm      = "hmac-sha256"
  secret         = var.dns_xfr_internal_secret
  secret_version = 1
}

resource "opnsense_bind_tsig_key" "public_transfer" {
  name           = "dns-xfr-public.example.net"
  algorithm      = "hmac-sha256"
  secret         = var.dns_xfr_public_secret
  secret_version = 1
}

resource "opnsense_bind_view" "internal" {
  sequence                          = 10
  name                              = "internal"
  match_client_acl_ids              = [opnsense_bind_acl.internal.id]
  match_client_tsig_key_ids         = [opnsense_bind_tsig_key.internal_transfer.id]
  exclude_match_client_tsig_key_ids = [opnsense_bind_tsig_key.public_transfer.id]
  recursion                         = true
  allow_recursion_acl_ids           = [opnsense_bind_acl.internal.id]
  allow_query_acl_ids               = [opnsense_bind_acl.internal.id]
  dnssec_validation                 = "auto"
}

resource "opnsense_bind_view" "public" {
  sequence        = 100
  name            = "public"
  match_any       = true
  allow_query_any = true
  recursion       = false
}
