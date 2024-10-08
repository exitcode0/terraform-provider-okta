---
page_title: "Resource: okta_admin_role_custom"
description: |-
  Resource to manage administrative Role assignments for a User
  These operations allow the creation and manipulation of custom roles as custom collections of permissions.
---

# Resource: okta_admin_role_custom

Resource to manage administrative Role assignments for a User

These operations allow the creation and manipulation of custom roles as custom collections of permissions.

## Example Usage

```terraform
resource "okta_admin_role_custom" "example" {
  label       = "AppAssignmentManager"
  description = "This role allows app assignment management"
  permissions = ["okta.apps.assignment.manage"]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `description` (String) A human-readable description of the new Role
- `label` (String) The name given to the new Role

### Optional

- `permissions` (Set of String) The permissions that the new Role grants. At least one
				permission must be specified when creating custom role. Valid values: "okta.authzServers.manage",
			  "okta.authzServers.read",
			  "okta.apps.assignment.manage",
			  "okta.apps.manage",
			  "okta.apps.read",
			  "okta.customizations.manage",
			  "okta.customizations.read",
			  "okta.groups.appAssignment.manage",
			  "okta.groups.create",
			  "okta.groups.manage",
			  "okta.groups.members.manage",
			  "okta.groups.read",
			  "okta.profilesources.import.run",
			  "okta.users.appAssignment.manage",
			  "okta.users.create",
			  "okta.users.credentials.expirePassword",
			  "okta.users.credentials.manage",
			  "okta.users.credentials.resetFactors",
			  "okta.users.credentials.resetPassword",
			  "okta.users.groupMembership.manage",
			  "okta.users.lifecycle.activate",
			  "okta.users.lifecycle.clearSessions",
			  "okta.users.lifecycle.deactivate",
			  "okta.users.lifecycle.delete",
			  "okta.users.lifecycle.manage",
			  "okta.users.lifecycle.suspend",
			  "okta.users.lifecycle.unlock",
			  "okta.users.lifecycle.unsuspend",
			  "okta.users.manage",
			  "okta.users.read",
			  "okta.users.userprofile.manage",
			  "okta.workflows.invoke".,

### Read-Only

- `id` (String) The ID of this resource.

## Import

Import is supported using the following syntax:

```shell
terraform import okta_admin_role_custom.example <custom_role_id>
```
