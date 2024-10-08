---
page_title: "Resource: okta_group_memberships"
description: |-
  Resource to manage a set of memberships for a specific group.
  This resource will allow you to bulk manage group membership in Okta for a given
  group. This offers an interface to pass multiple users into a single resource
  call, for better API resource usage. If you need a relationship of a single
  user to many groups, please use the 'oktausergroupmemberships' resource.
  Important: The default behavior of the resource is to only maintain the
  state of user ids that are assigned it. This behavior will signal drift only if
  those users stop being part of the group. If the desired behavior is track all
  users that are added/removed from the group make use of the 'trackall_users'
  argument with this resource.
---

# Resource: okta_group_memberships

Resource to manage a set of memberships for a specific group.
This resource will allow you to bulk manage group membership in Okta for a given
group. This offers an interface to pass multiple users into a single resource
call, for better API resource usage. If you need a relationship of a single 
user to many groups, please use the 'okta_user_group_memberships' resource.
**Important**: The default behavior of the resource is to only maintain the
state of user ids that are assigned it. This behavior will signal drift only if
those users stop being part of the group. If the desired behavior is track all
users that are added/removed from the group make use of the 'track_all_users'
argument with this resource.

## Example Usage

```terraform
resource "okta_group" "test" {
  name        = "testAcc_replace_with_uuid"
  description = "testing, testing"
}

resource "okta_group_memberships" "test" {
  group_id = okta_group.test.id
  users = [
    okta_user.test1.id,
    okta_user.test2.id,
  ]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `group_id` (String) ID of a Okta group.
- `users` (Set of String) The list of Okta user IDs which the group should have membership managed for.

### Optional

- `track_all_users` (Boolean) The resource concerns itself with all users added/deleted to the group; even those managed outside of the resource.

### Read-Only

- `id` (String) The ID of this resource.

## Import

Import is supported using the following syntax:

```shell
# an Okta Group's memberships can be imported via the Okta group ID.
terraform import okta_group_memberships.test <group_id>
# optional parameter track all users will also import all user id currently assigned to the group
terraform import okta_group_memberships.test <group_id>/<true>
```
