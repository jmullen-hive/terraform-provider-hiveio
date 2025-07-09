resource "hiveio_gateway_host" "gateway" {
  hostid     = hiveio_host.gateway.id
  start_port = 10000
  end_port   = 10100
  address    = "gw.domain.lan"
}