provider "hiveio" {
  host     = "hive1"
  username = "admin"
  password = "admin"
  insecure = true
}

variable "license" {
  type    = string
}

resource "hiveio_license" "license" {
  license = var.license
}

//IP Addresses or hostnames of the hosts to add
variable "hosts" {
  type    = set(string)
  default = ["192.168.3.197", "192.168.3.121" ]
}

resource "hiveio_host" "host" {
  for_each = var.hosts
  ip_address = each.key
  username = "admin"
  password = "admin"
  depends_on = [
    hiveio_license.license
  ]
}