---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "influxdb-v2_legacy_authorization Resource - terraform-provider-influxdb-v2"
subcategory: ""
description: |-
---

# influxdb-v2_legacy_authorization (Resource)

## Example Usage

```terraform
locals {
  org_id    = "example_org_id"
  bucket_id = "example_bucket_id"
}

resource "influxdb-v2_legacy_authorization" "example_authorization" {
  org_id      = local.org_id
  description = "Example description"
  name        = "username"
  password    = "secure password"
  status      = "active"
  permissions {
    action = "read"
    resource {
      id     = local.bucket_id
      org_id = local.org_id
      type   = "buckets"
    }
  }
  permissions {
    action = "write"
    resource {
      id     = local.bucket_id
      org_id = local.org_id
      type   = "buckets"
    }
  }
}
```

<!-- schema generated by tfplugindocs -->

## Schema

### Required

- `name` (String)
- `org_id` (String)
- `password` (String, Sensitive)
- `permissions` (Block Set, Min: 1) (see [below for nested schema](#nestedblock--permissions))

### Optional

- `description` (String)
- `status` (String)
- `user_id` (String)
- `user_org_id` (String)

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--permissions"></a>

### Nested Schema for `permissions`

Required:

- `action` (String)
- `resource` (Block Set, Min: 1) (see [below for nested schema](#nestedblock--permissions--resource))

<a id="nestedblock--permissions--resource"></a>

### Nested Schema for `permissions.resource`

Required:

- `org_id` (String)
- `type` (String)

Optional:

- `org` (String)

Read-Only:

- `id` (String) The ID of this resource.
