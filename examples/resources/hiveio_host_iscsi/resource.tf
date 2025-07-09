
resource "hiveio_host_iscsi" "iscsi-test" {
  hostid   = hiveio_host.id
  portal   = "iscsi-portal-address"
  target   = "iqn.2000-01.com.synology:synology.Target-1.444c59e2ab"
  username = "iscsi-user"
  password = "iscsi-password"
}