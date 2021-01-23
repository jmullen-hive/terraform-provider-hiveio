
# Create a persistent Windows 10 desktop pool on nfs
resource "hiveio_guest_pool" "win10_pool" {
  name         = "win10"
  cpu          = 4
  memory       = 8192
  density      = [1, 2]
  seed         = "WIN10"
  template     = "template_id"
  profile      = "profile_id"
  persistent   = false
  storage_type = "nfs"
  storage_id   = hiveio_storage_pool.vms.id
}

#Create a non-persistent ubuntu pool on disk
resource "hiveio_guest_pool" "ubuntu_pool" {
  name         = "ubuntu"
  cpu          = 2
  memory       = 1024
  density      = [2, 4]
  seed         = "UBUNTU"
  template     = hiveio_template.ubuntu_server.id
  profile      = hiveio_profile.default_profile.id
  persistent   = false
  storage_type = "disk"
  storage_id   = "disk"
  backup {
    enabled   = true
    frequency = "daily"
    target    = hiveio_storage_pool.backup.id
  }
}