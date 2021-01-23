
resource "hiveio_template" "win10_uefi" {
  name     = "win10-uefi"
  cpu      = 4
  mem      = 8192
  firmware = "uefi"
  os       = "win10"
  disk {
    disk_driver = "virtio"
    storage_id  = hiveio_storage_pool.vms.id
    filename    = "win10-1909-uefi.qcow2"
    type        = "disk"
  }

  interface {
    emulation = "virtio"
    network   = "prod"
    vlan      = 0
  }
}


resource "hiveio_template" "ubuntu_server" {
  name           = "ubuntu-server"
  cpu            = 2
  mem            = 1024
  firmware       = "bios"
  os             = "linux"
  display_driver = "cirrus"
  disk {
    disk_driver = "virtio"
    storage_id  = hiveio_storage_pool.vms.id
    filename    = hiveio_disk.disk3.filename
    type        = "disk"
  }

  interface {
    emulation = "virtio"
    network   = "prod"
    vlan      = 0
  }
}