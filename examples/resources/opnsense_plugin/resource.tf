resource "opnsense_plugin" "bind" {
  name                 = "os-bind"
  locked               = false
  uninstall_on_destroy = false
}
