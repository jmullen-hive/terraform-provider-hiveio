variable "ip_address" {
  type = string
}

resource "hiveio_host" "host" {
  ip_address = var.ip_address
  username   = "admin"
  password   = "password"
}