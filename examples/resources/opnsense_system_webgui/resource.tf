import {
  to = opnsense_system_webgui.main
  id = "system_webgui"
}

resource "opnsense_system_webgui" "main" {
  protocol         = "https"
  port             = 443
  interfaces       = ["lan"]
  certificate_ref  = "existing-certificate-reference"
  session_timeout  = 900
  hsts             = true
  allow_readdress  = false
}
