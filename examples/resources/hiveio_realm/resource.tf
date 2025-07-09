variable "realm_user" { type = string }
variable "realm_password" { type = string }

resource "hiveio_realm" "test" {
  name     = "TEST"
  fqdn     = "test.my.lan"
  username = var.realm_user
  password = var.realm_password
}