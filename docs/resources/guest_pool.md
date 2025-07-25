---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "hiveio_guest_pool Resource - terraform-provider-hiveio"
subcategory: ""
description: |-
  
---

# hiveio_guest_pool (Resource)



## Example Usage

```terraform
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
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `density` (List of Number)
- `name` (String)
- `profile` (String)
- `seed` (String)
- `template` (String)

### Optional

- `allowed_hosts` (List of String)
- `backup` (Block List, Max: 1) (see [below for nested schema](#nestedblock--backup))
- `broker_connection` (Block List) (see [below for nested schema](#nestedblock--broker_connection))
- `broker_default_connection` (String) Defaults to ``.
- `cloudinit_enabled` (Boolean) Defaults to `false`.
- `cloudinit_userdata` (String) Defaults to ``.
- `cpu` (Number)
- `gpu` (Boolean) Defaults to `false`.
- `memory` (Number)
- `persistent` (Boolean) Defaults to `false`.
- `provider_override` (Block List, Max: 1) Override the provider configuration for this resource.  This can be used to connect to a different cluster or change credentials (see [below for nested schema](#nestedblock--provider_override))
- `storage_id` (String) Defaults to `disk`.
- `storage_type` (String) Defaults to `disk`.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- `wait_for_build` (Boolean) Defaults to `false`.

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--backup"></a>
### Nested Schema for `backup`

Required:

- `enabled` (Boolean)
- `frequency` (String)
- `target` (String)


<a id="nestedblock--broker_connection"></a>
### Nested Schema for `broker_connection`

Required:

- `name` (String)
- `port` (Number)
- `protocol` (String)

Optional:

- `description` (String) Defaults to ``.
- `disable_html5` (Boolean) Defaults to `false`.
- `gateway` (Block List, Max: 1) (see [below for nested schema](#nestedblock--broker_connection--gateway))

<a id="nestedblock--broker_connection--gateway"></a>
### Nested Schema for `broker_connection.gateway`

Optional:

- `disabled` (Boolean) Defaults to `false`.
- `persistent` (Boolean) Defaults to `false`.
- `protocols` (List of String)



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
