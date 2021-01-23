
variable "address" {
  type = string
}

resource "hiveio_external_guest" "desktop" {
  name = "desktop"

  # Address can be an ip address or hostname
  address = var.address

  username = "user1"
  realm    = "realm_name"
  os       = "linux"
}