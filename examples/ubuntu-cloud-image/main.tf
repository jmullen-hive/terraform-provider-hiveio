// This will download https://cloud-images.ubuntu.com/eoan/current/eoan-server-cloudimg-amd64.img into a nfs storage pool and create a guest pool

provider "hiveio" {
  host     = "hive1"
  username = "admin"
  password = "admin"
  insecure = true
}

resource "hiveio_disk" "disk3" {
  filename     = "terraform3.qcow2"
  format       = "qcow2"
  storage_pool = hiveio_storage_pool.vms.id
  src_url = "https://cloud-images.ubuntu.com/eoan/current/eoan-server-cloudimg-amd64.img"
  size = 30
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

data "template_file" "user_data" {
  template = file("user-data")
}

resource "hiveio_guest_pool" "ubuntu_pool" {
  name         = "ubuntu"
  cpu          = 1
  memory       = 1024
  density      = [2, 4]
  seed         = "UBUNTU"
  template     = hiveio_template.ubuntu_server.id
  profile      = hiveio_profile.default_profile.id
  persistent = false
  storage_type = "disk"
  storage_id   = "disk"
  cloudinit_enabled = true
  cloudinit_userdata = data.template_file.user_data.rendered
}
