resource "opnsense_caddy_handler" "application" {
  domain_id         = opnsense_caddy_domain.internal.id
  upstream_domains  = ["10.20.0.10", "10.20.0.11"]
  upstream_port     = 8443
  upstream_protocol = "https"

  tls_trust_ca_ref_id = var.internal_ca_ref_id
  tls_server_name     = "backend.internal.example"

  load_balancing_policy = "round_robin"
  health_uri            = "/health"
  health_status         = "200"
  description           = "Application backends"
}
