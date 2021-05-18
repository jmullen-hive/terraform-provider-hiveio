# HiveIO Terraform

Terraform provider for hive fabric

## Host Settings
provider settings can be read from environment variables or stored in tf files
```bash
# Can use environment variables or store in tf file
export HIO_HOST=hive1
export HIO_USER=admin
export HIO_PASS=password
export HIO_REALM=local
```

## Run

```
terraform init
terraform plan
terraform apply
```

### Example file

main.tf:
```

provider "hiveio" {
  host = "hive1"
  username = "admin"
  password = "admin"
  insecure = true
}

resource "hiveio_realm" "test_profile" {
  name = "TEST"
  fqdn = "test.domain.com"
}

resource "hiveio_disk" "terraform_disk" {
  filename = "terraform.qcow2"
  format = "qcow2"
  storage_pool = "${hiveio_storage_pool.vms.id}"
  size = 10
}

resource "hiveio_profile" "test_profile" {
  name = "test"
  timezone = "disabled"
  ad_config {
    domain = "${hiveio_realm.test.name}"
    username = "adJoinUser"
    password = "Password1"
    user_group = "Users"
  }
  user_volumes {
    repository = "${hiveio_storage_pool.uvs.id}"
    backup_schedule = 3600
    target = "disk"
    size = 10
  }
}

resource "hiveio_storage_pool" "vms" {
  name   = "vms"
  type   = "nfs"
  server = "synology"
  path   = "/volume1/vms"
  roles = ["guest", "template"]
}

resource "hiveio_storage_pool" "iso" {
  name   = "iso"
  type   = "nfs"
  server = "synology"
  path   = "/volume1/iso"
  roles = ["iso"]
}

resource "hiveio_storage_pool" "uvs" {
  name   = "uvs"
  type   = "nfs"
  server = "synology"
  path   = "/volume1/uvs"
  roles = ["userVolume"]
}

resource "hiveio_template" "win10_bios" {
  name = "win10"
  cpu = 2
  mem = 2048
  firmware = "bios"
  os = "win10"
  disk {
    disk_driver = "virtio"
    storage_id = "${hiveio_storage_pool.vms.id}"
    filename = "win10.qcow2"
    type = "disk"
  }
  interface {
    emulation = "virtio"
    network = "prod"
    vlan = 0
  }
}

resource "hiveio_guest_pool" "win10_pool" {
  name = "win10"
  cpu = 2
  memory = 2048
  density = [1,2]
  seed = "POOL"
  template = "${hiveio_template.win10_bios.id}"
  profile = "${hiveio_profile.home_profile.id}"
  storage_type = "disk"
  storage_id = "disk"
}

resource "hiveio_virtual_machine" "kubuntu" {
  name = "kubuntu"
  cpu = 2
  memory = 1024
  firmware = "uefi"
  os = "linux"
  display_driver = "qxl"
  disk {
    disk_driver = "virtio"
    storage_id = "${hiveio_storage_pool.vms.id}"
    filename = "kubuntu-18.10-uefi.qcow2"
    type = "disk"
  }
  interface {
    emulation = "virtio"
    network = "prod"
    vlan = 0
  }
}

```
