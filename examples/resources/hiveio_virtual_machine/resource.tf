resource "hiveio_virtual_machine" "win10" {
  name           = "win10"
  cpu            = 2
  memory         = 4096
  firmware       = "uefi"
  os             = "win10"
  display_driver = "qxl"
  inject_agent   = true
  disk {
    disk_driver = "virtio"
    storage_id  = hiveio_storage_pool.vms.id
    filename    = "windows-uefi.qcow2"
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