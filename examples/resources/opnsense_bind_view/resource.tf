resource "opnsense_bind_view" "internal" {
  sequence                = 10
  name                    = "internal"
  match_client_acl_ids    = [opnsense_bind_acl.internal.id]
  recursion               = true
  allow_recursion_acl_ids = [opnsense_bind_acl.internal.id]
  allow_query_acl_ids     = [opnsense_bind_acl.internal.id]
  forwarders              = ["1.1.1.1", "9.9.9.9"]
  dnssec_validation       = "auto"
}

resource "opnsense_bind_view" "public" {
  sequence        = 100
  name            = "public"
  match_any       = true
  allow_query_any = true
  recursion       = false
}
