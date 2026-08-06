// The package must be available from a configured OPNsense package repository.
resource "opnsense_plugin" "api_extensions" {
  name = "os-api-extensions"
}

resource "opnsense_system_ssh" "main" {
  enabled                 = true
  port                    = 22
  interfaces              = ["lan"]
  password_authentication = false
  permit_root_login       = false
  allow_readdress         = false

  depends_on = [opnsense_plugin.api_extensions]
}
