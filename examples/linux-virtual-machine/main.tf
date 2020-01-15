provider "hiveio" {
  host     = "hive1"
  username = "admin"
  password = "admin"
  insecure = true
}

resource "hiveio_profile" "default_profile" {
  name     = "default"
  timezone = "disabled"
}

resource "hiveio_storage_pool" "vms" {
  name   = "vms"
  type   = "nfs"
  server = "synology"
  path   = "/volume1/vms"
  roles  = ["guest", "template"]
}

resource "hiveio_storage_pool" "backup" {
  name   = "backup"
  type   = "nfs"
  server = "synology"
  path   = "/volume1/backup"
  roles  = ["backup"]
}

resource "hiveio_virtual_machine" "kubuntu" {
  name           = "kubuntu"
  cpu            = 2
  memory         = 2048
  firmware       = "uefi"
  os             = "linux"
  display_driver = "qxl"
  inject_agent   = true
  disk {
    disk_driver = "virtio"
    storage_id  = hiveio_storage_pool.vms.id
    filename    = "kubuntu-18.10-uefi.qcow2"
    type        = "disk"
  }
  interface {
    emulation = "virtio"
    network   = "prod"
    vlan      = 0
  }
   backup {
    enabled   = true  
    frequency = "daily"
    target    = hiveio_storage_pool.backup.id
  }
}


