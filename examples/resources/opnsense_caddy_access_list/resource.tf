resource "opnsense_caddy_access_list" "management" {
  name       = "management"
  client_ips = ["10.0.0.0/24", "10.10.0.0/24"]

  request_matcher = "client_ip"
  description     = "Management and VPN networks"
}
