import {
  to = opnsense_caddy_settings.main
  id = "caddy_settings"
}

resource "opnsense_caddy_settings" "main" {
  enabled       = true
  http_port     = 8080
  https_port    = 8443
  acme_email    = "operations@example.com"
  run_as_user   = "www"
  http_versions = ["h1", "h2"]
}
