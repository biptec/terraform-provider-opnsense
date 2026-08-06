import {
  to = opnsense_system_ssh.main
  id = "system_ssh"
}

resource "opnsense_system_ssh" "main" {
  enabled                 = true
  port                    = 22
  interfaces              = ["lan"]
  password_authentication = false
  permit_root_login       = false
  allow_readdress         = false
}
