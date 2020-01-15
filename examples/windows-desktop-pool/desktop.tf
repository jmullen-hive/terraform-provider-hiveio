
provider "hiveio" {
  host     = "hive1"
  username = "admin"
  password = "admin"
  insecure = true
}

resource "hiveio_realm" "test" {
  name = "TEST"
  fqdn = "test.my-domain.net"
}

resource "hiveio_storage_pool" "vms" {
  name   = "vms"
  type   = "nfs"
  server = "synology"
  path   = "/volume1/vms"
  roles  = ["guest", "template"]
}

resource "hiveio_storage_pool" "uvs" {
  name   = "uvs"
  type   = "nfs"
  server = "synology"
  path   = "/volume1/uvs"
  roles  = ["userVolume"]
}

resource "hiveio_profile" "test_profile" {
  name     = "test"
  timezone = "disabled"
  ad_config {
    domain     = hiveio_realm.test.name
    username   = "serviceAccount"
    password   = "Password123"
    user_group = "Users"
  }
  user_volumes {
    repository      = hiveio_storage_pool.uvs.id
    backup_schedule = 3600
    target          = "disk"
    size            = 10
  }
}

resource "hiveio_template" "win10" {
  name     = "win10"
  cpu      = 2
  mem      = 2048
  firmware = "bios"
  os       = "win10"
  disk {
    disk_driver = "virtio"
    storage_id  = hiveio_storage_pool.vms.id
    filename    = "win10.qcow2"
    type        = "disk"
  }

  interface {
    emulation = "virtio"
    network   = "prod"
    vlan      = 0
  }
}

resource "hiveio_guest_pool" "win10_pool" {
  name         = "win10"
  cpu          = 2
  memory       = 2048
  density      = [1, 2]
  seed         = "WIN10"
  template     = hiveio_template.win10.id
  profile      = hiveio_profile.test_profile.id
  persistent   = true
  storage_type = "nfs"
  storage_id   = hiveio_storage_pool.vms.id
}
