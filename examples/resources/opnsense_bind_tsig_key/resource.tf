variable "acme_tsig_secret" {
  type      = string
  sensitive = true
}

resource "opnsense_bind_tsig_key" "acme" {
  name      = "_acme-challenge.web.example.host.acme.example.net"
  algorithm = "hmac-sha256"
  secret         = var.acme_tsig_secret
  secret_version = 1
}
