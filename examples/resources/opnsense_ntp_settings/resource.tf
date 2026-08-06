// The package must be available from a configured OPNsense package repository.
resource "opnsense_plugin" "api_extensions" {
  name = "os-api-extensions"
}

resource "opnsense_ntp_settings" "main" {
  enabled    = true
  interfaces = ["opt1"]
  orphan     = 12
  max_clock  = 10

  servers = [
    {
      host     = "0.pool.ntp.org"
      pool     = true
      iburst   = true
      prefer   = true
      noselect = false
    }
  ]

  kiss_of_death          = true
  rate_limiting          = true
  deny_modifications     = true
  disable_queries        = true
  deny_peer_associations = true
  deny_trap_service      = true

  depends_on = [opnsense_plugin.api_extensions]
}
