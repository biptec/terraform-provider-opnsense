variable "acme_tsig_secret" {
  type      = string
  sensitive = true
}

resource "opnsense_bind_tsig_key" "acme" {
  name      = "acme-dns01"
  algorithm = "hmac-sha256"
  secret    = var.acme_tsig_secret
}
