resource "opnsense_unbound_service" "resolver" {
  # BIND owns TCP/UDP 53 in this deployment.
  enabled = false
}
