resource "opnsense_bind_forward_domain" "corp" {
  view_id         = opnsense_bind_view.internal.id
  domain_name     = "corp.example.net"
  forward_servers = ["10.10.0.53", "10.10.0.54"]
}
