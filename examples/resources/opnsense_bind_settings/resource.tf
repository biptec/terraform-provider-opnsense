resource "opnsense_bind_settings" "main" {
  enabled               = true
  listen_ipv4           = ["192.0.2.53", "10.0.0.1"]
  listen_ipv6           = ["2001:db8::53"]
  transfer_source       = "10.0.0.1"
  transfer_source_ipv6  = "2001:db8::53"
  notify_source         = "10.0.0.1"
  notify_source_ipv6    = "2001:db8::53"
  port                  = 53
  hide_hostname         = true
  hide_version          = true
  enable_rate_limiting  = true
  rate_limit_count      = 20
  rate_limit_exceptions = ["10.0.0.0/8"]
}
