variable "acme_tsig_secret" {
  type      = string
  sensitive = true
}

variable "secondary_transfer_tsig_secret" {
  type      = string
  sensitive = true
}

resource "opnsense_bind_tsig_key" "acme" {
  name      = "_acme-challenge.example.net"
  algorithm = "hmac-sha256"
  secret    = var.acme_tsig_secret
}

resource "opnsense_bind_tsig_key" "secondary_transfer" {
  name      = "secondary-transfer.example.net"
  algorithm = "hmac-sha256"
  secret    = var.secondary_transfer_tsig_secret
}

resource "opnsense_bind_primary_domain" "public" {
  view_id         = opnsense_bind_view.public.id
  domain_name     = "example.net"
  transfer_key_id = opnsense_bind_tsig_key.secondary_transfer.id
  also_notify     = ["192.0.2.54"]
  update_key_ids  = [opnsense_bind_tsig_key.acme.id]
  update_policy   = "self_txt"
  dnssec          = true

  ttl          = 300
  refresh      = 3600
  retry        = 600
  expire       = 1209600
  negative_ttl = 300
  mail_admin   = "hostmaster@example.net"
  dns_server   = "ns1.example.net"
}
