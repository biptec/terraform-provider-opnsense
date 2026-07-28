terraform {
  required_providers {
    opnsense = {
      source = "biptec/opnsense"
    }
  }
}

# OPNSENSE_URI, OPNSENSE_API_KEY, OPNSENSE_API_SECRET, and optionally
# OPNSENSE_ALLOW_INSECURE are read from the environment.
provider "opnsense" {}

data "opnsense_interfaces_overview_all" "all" {}

output "interface_devices" {
  value = [
    for item in data.opnsense_interfaces_overview_all.all.interfaces : item.device
  ]
}
