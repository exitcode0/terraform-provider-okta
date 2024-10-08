---
page_title: "Resource: okta_auth_server_policy_rule"
description: |-
  Creates an Authorization Server Policy Rule.
  This resource allows you to create and configure an Authorization Server Policy Rule.
  -> This resource is concurrency safe. However, when creating/updating/deleting
  multiple rules belonging to a policy, the Terraform meta argument
  'depends_on' https://www.terraform.io/language/meta-arguments/depends_on
  should be added to each rule chaining them all in sequence. Base the sequence on
  the 'priority' property in ascending value.
---

# Resource: okta_auth_server_policy_rule

Creates an Authorization Server Policy Rule.
This resource allows you to create and configure an Authorization Server Policy Rule.
-> This resource is concurrency safe. However, when creating/updating/deleting
multiple rules belonging to a policy, the Terraform meta argument
['depends_on'](https://www.terraform.io/language/meta-arguments/depends_on)
should be added to each rule chaining them all in sequence. Base the sequence on
the 'priority' property in ascending value.

## Example Usage

```terraform
resource "okta_auth_server_policy_rule" "example" {
  auth_server_id       = "<auth server id>"
  policy_id            = "<auth server policy id>"
  status               = "ACTIVE"
  name                 = "example"
  priority             = 1
  group_whitelist      = ["<group ids>"]
  grant_type_whitelist = ["implicit"]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `auth_server_id` (String) Auth server ID
- `grant_type_whitelist` (Set of String) Accepted grant type values, `authorization_code`, `implicit`, `password`, `client_credentials`, `urn:ietf:params:oauth:grant-type:saml2-bearer` (*Early Access Property*), `urn:ietf:params:oauth:grant-type:token-exchange` (*Early Access Property*),`urn:ietf:params:oauth:grant-type:device_code` (*Early Access Property*), `interaction_code` (*OIE only*). For `implicit` value either `user_whitelist` or `group_whitelist` should be set.
- `name` (String) Auth server policy rule name
- `policy_id` (String) Auth server policy ID
- `priority` (Number) Priority of the auth server policy rule

### Optional

- `access_token_lifetime_minutes` (Number) Lifetime of access token. Can be set to a value between 5 and 1440 minutes. Default is `60`.
- `group_blacklist` (Set of String) Specifies a set of Groups whose Users are to be excluded.
- `group_whitelist` (Set of String) Specifies a set of Groups whose Users are to be included. Can be set to Group ID or to the following: `EVERYONE`.
- `inline_hook_id` (String) The ID of the inline token to trigger.
- `refresh_token_lifetime_minutes` (Number) Lifetime of refresh token.
- `refresh_token_window_minutes` (Number) Window in which a refresh token can be used. It can be a value between 5 and 2628000 (5 years) minutes. Default is `10080` (7 days).`refresh_token_window_minutes` must be between `access_token_lifetime_minutes` and `refresh_token_lifetime_minutes`.
- `scope_whitelist` (Set of String) Scopes allowed for this policy rule. They can be whitelisted by name or all can be whitelisted with ` * `
- `status` (String) Default to `ACTIVE`
- `type` (String) Auth server policy rule type, unlikely this will be anything other then the default
- `user_blacklist` (Set of String) Specifies a set of Users to be excluded.
- `user_whitelist` (Set of String) Specifies a set of Users to be included.

### Read-Only

- `id` (String) The ID of this resource.
- `system` (Boolean) The rule is the system (default) rule for its associated policy

## Import

Import is supported using the following syntax:

```shell
terraform import okta_auth_server_policy_rule.example <auth_server_id>/<policy_id>/<policy_rule_id>
```
