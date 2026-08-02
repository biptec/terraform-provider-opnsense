# Public domain with automatic ACME certificates.
resource "opnsense_caddy_domain" "public" {
  domain           = "app.example.com"
  certificate_mode = "acme"
  description      = "Public application"
}

# Internal domain with a dynamically issued certificate from an existing CA.
resource "opnsense_caddy_domain" "internal" {
  domain                             = "app.internal.example"
  certificate_mode                   = "internal"
  internal_ca_name                   = "internal.example"
  internal_certificate_lifetime_days = 3650
  description                        = "Internal application"
}

# Plain HTTP listener, normally used only on trusted networks.
resource "opnsense_caddy_domain" "http" {
  domain           = "health.internal.example"
  protocol         = "http"
  certificate_mode = "none"
}
