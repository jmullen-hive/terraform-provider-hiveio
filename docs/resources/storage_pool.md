---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "hiveio_storage_pool Resource - terraform-provider-hiveio"
subcategory: ""
description: |-
  
---

# hiveio_storage_pool (Resource)



## Example Usage

```terraform
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

#create an ocfs2 storage pool on an iscsi disk and create the fs if it does not exist
resource "hiveio_storage_pool" "ocsfs2-test" {
  name              = "ocsfs2-test"
  type              = "ocfs2"
  roles             = ["iso", "guest", "template", "backup"]
  device            = hiveio_host_iscsi.iscsi_test.block_devices[0].path
  fs_name           = "ocfs2-test"
  create_filesystem = true # create the filesystem if it does not exist
  clear_disk        = false
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String)
- `roles` (List of String)
- `type` (String)

### Optional

- `clear_disk` (Boolean) Defaults to `false`.
- `create_filesystem` (Boolean) Defaults to `false`.
- `device` (String)
- `fs_name` (String)
- `hosts` (List of String) List of host IDs that should add the storage pool
- `key` (String, Sensitive)
- `mount_options` (List of String)
- `password` (String, Sensitive)
- `path` (String)
- `provider_override` (Block List, Max: 1) Override the provider configuration for this resource.  This can be used to connect to a different cluster or change credentials (see [below for nested schema](#nestedblock--provider_override))
- `s3_access_key_id` (String)
- `s3_region` (String)
- `s3_secret_access_key` (String, Sensitive)
- `server` (String)
- `tags` (List of String)
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- `url` (String)
- `username` (String)

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--provider_override"></a>
### Nested Schema for `provider_override`

Required:

- `password` (String, Sensitive) The password to use for connection to the server.

Optional:

- `host` (String) hostname or ip address of the server.
- `insecure` (Boolean) Ignore SSL certificate errors. Defaults to `false`.
- `port` (Number) The port to use to connect to the server. Defaults to 8443
- `realm` (String, Sensitive) The realm to use to connect to the server. Defaults to local
- `username` (String) The username to connect to the server. Defaults to admin


<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `delete` (String)
