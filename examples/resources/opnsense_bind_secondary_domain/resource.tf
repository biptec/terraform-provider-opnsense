variable "transfer_secret" {
  type      = string
  sensitive = true
  ephemeral = true
}

resource "opnsense_bind_secondary_domain" "example" {
  view_id                = opnsense_bind_view.public.id
  domain_name            = "example.net"
  primary_ips            = ["192.0.2.53"]
  allow_notify           = ["192.0.2.53"]
  transfer_key_name      = "secondary-transfer"
  transfer_key_algorithm = "hmac-sha256"
  transfer_key           = var.transfer_secret
  transfer_key_version   = 1
}
