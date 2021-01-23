variable "realm_user" { type = string }
variable "realm_password" { type = string }

resource "hiveio_realm" "home" {
  name     = "TEST"
  fqdn     = "test.test-domain.net"
  username = var.realm_user
  password = var.realm_password
}