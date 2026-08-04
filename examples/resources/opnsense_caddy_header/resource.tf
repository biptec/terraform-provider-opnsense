resource "opnsense_caddy_header" "preserve_host" {
  direction   = "header_up"
  name        = "Host"
  value       = "{host}"
  description = "Preserve the frontend Host header for the upstream"
}
