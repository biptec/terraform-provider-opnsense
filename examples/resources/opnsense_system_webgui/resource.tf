// The package must be available from a configured OPNsense package repository.
resource "opnsense_plugin" "api_extensions" {
  name = "os-api-extensions"
}

resource "opnsense_system_webgui" "main" {
  protocol         = "https"
  port             = 443
  interfaces       = ["lan"]
  certificate_ref  = "existing-certificate-reference"
  session_timeout  = 900
  hsts             = true
  allow_readdress  = false

  depends_on = [opnsense_plugin.api_extensions]
}
