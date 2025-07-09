resource "hiveio_host_network" "vlan15" {
  name      = "vlan15"
  hostid    = "hostid-12345" # Replace with actual host ID
  interface = "ens3"
  vlan      = 15
}

resource "hiveio_host_network" "testnet" {
  name      = "testnet"
  hostid    = "hostid-12345" # Replace with actual host ID
  ip        = "192.168.1.2"  # Replace with actual IP
  mask      = "255.255.255.0"
  interface = "enp2s0"
  vlan      = 25
}