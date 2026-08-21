variable "transfer_secret" {
  type      = string
  sensitive = true
  ephemeral = true
}

resource "opnsense_bind_tsig_key" "secondary_transfer" {
  name           = "dns-xfr-public.example.net"
  algorithm      = "hmac-sha256"
  secret         = var.transfer_secret
  secret_version = 1
}

resource "opnsense_bind_secondary_domain" "example" {
  view_id         = opnsense_bind_view.public.id
  domain_name     = "example.net"
  primary_ips     = ["192.0.2.53"]
  transfer_key_id = opnsense_bind_tsig_key.secondary_transfer.id
}
