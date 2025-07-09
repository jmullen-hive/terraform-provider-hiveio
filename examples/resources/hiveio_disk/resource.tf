


resource "hiveio_disk" "new" {
  filename     = "new.qcow2"
  format       = "qcow2"
  storage_pool = storage_pool_id
  size         = 30
}

#Copy a disk
resource "hiveio_disk" "disk2" {
  filename     = "terraform2.qcow2"
  storage_pool = storage_pool_id
  src_storage  = storage_pool_id
  src_filename = "ubuntu-18.04-cloudimg-hiveio.img"
}

#Download the ubuntu 24.04 cloud image into a storage pool
resource "hiveio_disk" "ubuntu" {
  filename     = "ubuntu.qcow2"
  storage_pool = storage_pool_id
  src_url      = "https://cloud-images.ubuntu.com/noble/current/noble-server-cloudimg-amd64.img"
}

#Upload a file to a storage pool
resource "hiveio_disk" "upload-test" {
  filename     = "upload-test"
  storage_pool = storage_pool_id
  local_file   = "virtio-win.iso"
  format       = "raw"
}