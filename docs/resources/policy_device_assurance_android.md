---
page_title: "Resource: okta_policy_device_assurance_android"
description: |-
  Manages a device assurance policy for android.
---

# Resource: okta_policy_device_assurance_android

Manages a device assurance policy for android.

## Example Usage

```terraform
resource "okta_policy_device_assurance_android" "example" {
  name                    = "example"
  os_version              = "12"
  disk_encryption_type    = toset(["FULL", "USER"])
  jailbreak               = false
  secure_hardware_present = true
  screenlock_type         = toset(["BIOMETRIC"])
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Policy device assurance name

### Optional

- `disk_encryption_type` (Set of String) List of disk encryption type, can be `FULL`, `USER`
- `jailbreak` (Boolean) Is the device jailbroken in the device assurance policy.
- `os_version` (String) Minimum os version of the device in the device assurance policy.
- `screenlock_type` (Set of String) List of screenlock type, can be `BIOMETRIC` or `BIOMETRIC, PASSCODE`
- `secure_hardware_present` (Boolean) Indicates if the device contains a secure hardware functionality

### Read-Only

- `created_by` (String) Created by
- `created_date` (String) Created date
- `id` (String) Policy assurance id
- `last_update` (String) Last update
- `last_updated_by` (String) Last updated by
- `platform` (String) Policy device assurance platform

## Import

Import is supported using the following syntax:

```shell
terraform import okta_policy_device_assurance_android.example <device_assurance_id>
```
