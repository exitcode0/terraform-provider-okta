---
page_title: "Data Source: okta_user_type"
description: |-
  Get a user type from Okta.
---

# Data Source: okta_user_type

Get a user type from Okta.

## Example Usage

```terraform
data "okta_user_type" "example" {
  name = "example"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `id` (String) ID of the user type to retrieve, conflicts with `name`.
- `name` (String) Name of user type to retrieve, conflicts with `id`.

### Read-Only

- `description` (String) Description of user type.
- `display_name` (String) Display name of user type.


