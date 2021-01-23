resource "hiveio_storage_pool" "vms" {
  name   = "vms"
  type   = "nfs"
  server = "synology"
  path   = "/volume1/vms"
  roles  = ["guest", "template"]
}

resource "hiveio_storage_pool" "iso" {
  name   = "iso"
  type   = "nfs"
  server = "synology"
  path   = "/volume1/iso"
  roles  = ["iso"]
}

resource "hiveio_storage_pool" "uvs" {
  name   = "uvs"
  type   = "nfs"
  server = "synology"
  path   = "/volume1/uvs"
  roles  = ["userVolume"]
}

resource "hiveio_storage_pool" "cifs" {
  name     = "cifs"
  type     = "cifs"
  server   = "cifs_server_hostname"
  path     = "test"
  username = "user1"
  password = "Password"
  roles    = ["template", "iso"]
}


#S3 storage pool for backups
variable "s3_key_id" { type = string }
variable "s3_access_key" { type = string }
resource "hiveio_storage_pool" "s3_backup" {
  name                 = "s3_backup"
  type                 = "s3"
  path                 = "bucket_name"
  s3_access_key_id     = var.s3_key_id
  s3_secret_access_key = var.s3_access_key
  s3_region            = "us-east-1"
  roles                = ["backup"]
}

#Azure storage pool for backups
variable "azure_username" { type = string }
variable "azure_key" { type = string }
resource "hiveio_storage_pool" "azure" {
  name     = "azure"
  type     = "azure"
  path     = "backup-test"
  username = var.azure_username
  key      = var.azure_key
  roles    = ["backup"]
}