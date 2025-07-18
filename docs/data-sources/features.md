---
page_title: "Data Source: okta_features"
description: |-
  Get a list of features from Okta.
---

# Data Source: okta_brands

Get a list of features from Okta.

## Example Usage

```terraform
data "okta_features" "example" {
  label = "Android Device Trust"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `label` (String) Searches for features whose label or name property matches this value exactly. Case sensitive
- `substring` (String) Searches for features whose label or name property substring match this value. Case sensitive

### Read-Only

- `features` (Set of Object) List of `okta_feature` belonging to the organization (see [below for nested schema](#nestedatt--features))
- `id` (String) Generated ID

<a id="nestedatt--features"></a>
### Nested Schema for `features`

Read-Only:

- `id` (String)
- `name` (String)
- `status` (String)
- `type` (String)
- `description` (String)
- `stage` (Map of String)


