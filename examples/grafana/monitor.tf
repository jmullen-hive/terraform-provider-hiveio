provider "hiveio" {
  host     = "hive1"
  username = "admin"
  password = "admin"
  insecure = true
}

data "hiveio_storage_pool" "vms" {
  name = "vms"
}

resource "hiveio_disk" "monitor" {
  filename     = "monitor.qcow2"
  storage_pool = data.hiveio_storage_pool.vms.id
  src_url = "https://cloud-images.ubuntu.com/eoan/current/eoan-server-cloudimg-amd64.img"
}

data "template_file" "user_data" {
  template = file("user-data")
}

resource "hiveio_virtual_machine" "monitor" {
  name = "monitor"
  cpu = 2
  memory = 2048
  firmware = "uefi"
  os = "linux"
  display_driver = "qxl"
  disk {
    disk_driver = "virtio"
    storage_id = data.hiveio_storage_pool.vms.id
    filename = hiveio_disk.monitor.filename
    type = "disk"
  }
  interface {
    emulation = "virtio"
    network = "prod"
    vlan = 0
  }
  cloudinit_enabled = true
  cloudinit_userdata = data.template_file.user_data.rendered
}
