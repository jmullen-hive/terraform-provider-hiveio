provider "hiveio" {
  host     = "hive1"
  username = "admin"
  password = "admin"
  insecure = true
}

resource "hiveio_storage_pool" "nfs" {
  name   = "vms"
  type   = "nfs"
  server = "synology"
  path   = "/volume1/vms"
  roles  = ["guest", "iso", "template"]
}

resource "hiveio_storage_pool" "cifs" {
  name     = "cifs"
  type     = "cifs"
  server   = "synology"
  path     = "vms"
  username = "user1@test.domain.net"
  password = "password"
  roles    = ["template", "guest"]
}

variable "s3AccessKey" {
  type = string
}

resource "hiveio_storage_pool" "s3" {
  name                 = "s3"
  type                 = "s3"
  path                 = "hio-replication-test"
  s3_access_key_id     = "KEY_ID"
  s3_secret_access_key = var.s3AccessKey
  s3_region            = "us-east-1"
  roles                = ["iso", "backup"]
}

variable "azureAccessKey" {
  type = string
}

resource "hiveio_storage_pool" "azure" {
  name     = "azure"
  type     = "azure"
  path     = "rclone-test"
  username = "username"
  key      = var.azureAccessKey
  roles    = ["backup"]
}

resource "hiveio_storage_pool" "ftp" {
  name   = "ftp"
  type   = "ftp"
  server = "synology"
  path   = "/vms"
  roles  = ["template", "iso", "backup"]
}

resource "hiveio_storage_pool" "sftp" {
  name     = "sftp"
  type     = "sftp"
  server   = "synology"
  path     = "/vms"
  username = "user"
  password = "password"
  roles    = ["backup"]
}

resource "hiveio_storage_pool" "http" {
  name  = "http"
  type  = "http"
  url   = "https://cloud-images.ubuntu.com/jammy/current"
  roles = ["iso", "template"]
}

resource "hiveio_storage_pool" "ocsfs2" {
  name              = "ocsfs2-test"
  type              = "ocfs2"
  roles             = ["iso", "guest", "template"]
  device            = "/dev/sdb" #path to the shared disk
  fs_name           = "ocfs2-test"
  create_filesystem = true
  clear_disk        = true
}