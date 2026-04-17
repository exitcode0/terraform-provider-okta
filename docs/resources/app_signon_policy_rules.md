---
page_title: "Resource: okta_app_signon_policy_rules"
subcategory: "Applications"
description: |-

  Manages multiple app sign-on policy rules for a single policy. This resource allows you to define all rules for a policy in a single configuration block, ensuring consistent priority ordering and avoiding drift issues.

---

# Resource: okta_app_signon_policy_rules


Manages multiple app sign-on policy rules for a single policy. This resource allows you to define all rules for a policy in a single configuration block, ensuring consistent priority ordering and avoiding drift issues.


~> **IMPORTANT:** This resource uses name-first matching to identify and update rules. When migrating from individual `okta_app_signon_policy_rule` resources, ensure rule names remain consistent to enable safe adoption without data loss.

~> **NOTE ON RENAMING RULES:** If you rename a rule without explicitly preserving its `id`, the provider will treat it as a deletion of the old rule and creation of a new rule. To rename a rule while preserving its configuration and ID, you must explicitly set the `id` attribute in your configuration before changing the `name`.

## Links

- [Okta API docs](https://developer.okta.com/docs/api/openapi/okta-management/management/tag/ApplicationPolicies/)
- [Provider source](https://github.com/okta/terraform-provider-okta/blob/master/okta/services/idaas/resource_okta_app_signon_policy_rules.go)

## Related Resources

- [`okta_app_signon_policy`](../resources/app_signon_policy) — Parent sign-on policy

## Example Usage

```terraform
resource "okta_app_saml" "test" {
  label                     = "testAcc_replace_with_uuid"
  sso_url                   = "http://google.com"
  recipient                 = "http://here.com"
  destination               = "http://its-about-the-journey.com"
  audience                  = "http://audience.com"
  subject_name_id_template  = "$${user.userName}"
  subject_name_id_format    = "urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress"
  response_signed           = true
  signature_algorithm       = "RSA_SHA256"
  digest_algorithm          = "SHA256"
  honor_force_authn         = false
  authn_context_class_ref   = "urn:oasis:names:tc:SAML:2.0:ac:classes:PasswordProtectedTransport"
  single_logout_issuer      = "https://dunshire.okta.com"
  single_logout_url         = "https://dunshire.okta.com/logout"
  single_logout_certificate = "MIIFnDCCA4QCCQDBSLbiON2T1zANBgkqhkiG9w0BAQsFADCBjzELMAkGA1UEBhMCVVMxDjAMBgNV\r\nBAgMBU1haW5lMRAwDgYDVQQHDAdDYXJpYm91MRcwFQYDVQQKDA5Tbm93bWFrZXJzIEluYzEUMBIG\r\nA1UECwwLRW5naW5lZXJpbmcxDTALBgNVBAMMBFNub3cxIDAeBgkqhkiG9w0BCQEWEWVtYWlsQGV4\r\nYW1wbGUuY29tMB4XDTIwMTIwMzIyNDY0M1oXDTMwMTIwMTIyNDY0M1owgY8xCzAJBgNVBAYTAlVT\r\nMQ4wDAYDVQQIDAVNYWluZTEQMA4GA1UEBwwHQ2FyaWJvdTEXMBUGA1UECgwOU25vd21ha2VycyBJ\r\nbmMxFDASBgNVBAsMC0VuZ2luZWVyaW5nMQ0wCwYDVQQDDARTbm93MSAwHgYJKoZIhvcNAQkBFhFl\r\nbWFpbEBleGFtcGxlLmNvbTCCAiIwDQYJKoZIhvcNAQEBBQADggIPADCCAgoCggIBANMmWDjXPdoa\r\nPyzIENqeY9njLan2FqCbQPSestWUUcb6NhDsJVGSQ7XR+ozQA5TaJzbP7cAJUj8vCcbqMZsgOQAu\r\nO/pzYyQEKptLmrGvPn7xkJ1A1xLkp2NY18cpDTeUPueJUoidZ9EJwEuyUZIktzxNNU1pA1lGijiu\r\n2XNxs9d9JR/hm3tCu9Im8qLVB4JtX80YUa6QtlRjWR/H8a373AYCOASdoB3c57fIPD8ATDNy2w/c\r\nfCVGiyKDMFB+GA/WTsZpOP3iohRp8ltAncSuzypcztb2iE+jijtTsiC9kUA2abAJqqpoCJubNShi\r\nVff4822czpziS44MV2guC9wANi8u3Uyl5MKsU95j01jzadKRP5S+2f0K+n8n4UoV9fnqZFyuGAKd\r\nCJi9K6NlSAP+TgPe/JP9FOSuxQOHWJfmdLHdJD+evoKi9E55sr5lRFK0xU1Fj5Ld7zjC0pXPhtJf\r\nsgjEZzD433AsHnRzvRT1KSNCPkLYomznZo5n9rWYgCQ8HcytlQDTesmKE+s05E/VSWNtH84XdDrt\r\nieXwfwhHfaABSu+WjZYxi9CXdFCSvXhsgufUcK4FbYAHl/ga/cJxZc52yFC7Pcq0u9O2BSCjYPdQ\r\nDAHs9dhT1RhwVLM8RmoAzgxyyzau0gxnAlgSBD9FMW6dXqIHIp8yAAg9cRXhYRTNAgMBAAEwDQYJ\r\nKoZIhvcNAQELBQADggIBADofEC1SvG8qa7pmKCjB/E9Sxhk3mvUO9Gq43xzwVb721Ng3VYf4vGU3\r\nwLUwJeLt0wggnj26NJweN5T3q9T8UMxZhHSWvttEU3+S1nArRB0beti716HSlOCDx4wTmBu/D1MG\r\nt/kZYFJw+zuzvAcbYct2pK69AQhD8xAIbQvqADJI7cCK3yRry+aWtppc58P81KYabUlCfFXfhJ9E\r\nP72ffN4jVHpX3lxxYh7FKAdiKbY2FYzjsc7RdgKI1R3iAAZUCGBTvezNzaetGzTUjjl/g1tcVYij\r\nltH9ZOQBPlUMI88lxUxqgRTerpPmAJH00CACx4JFiZrweLM1trZyy06wNDQgLrqHr3EOagBF/O2h\r\nhfTehNdVr6iq3YhKWBo4/+RL0RCzHMh4u86VbDDnDn4Y6HzLuyIAtBFoikoKM6UHTOa0Pqv2bBr5\r\nwbkRkVUxl9yJJw/HmTCdfnsM9dTOJUKzEglnGF2184Gg+qJDZB6fSf0EAO1F6sTqiSswl+uHQZiy\r\nDaZzyU7Gg5seKOZ20zTRaX3Ihj9Zij/ORnrARE7eM/usKMECp+7syUwAUKxDCZkGiUdskmOhhBGL\r\nJtbyK3F2UvoJoLsm3pIcvMak9KwMjSTGJB47ABUP1+w+zGcNk0D5Co3IJ6QekiLfWJyQ+kKsWLKt\r\nzOYQQatrnBagM7MI2/T4\r\n"

  attribute_statements {
    type         = "GROUP"
    name         = "groups"
    filter_type  = "REGEX"
    filter_value = ".*"
  }
}

data "okta_app_signon_policy" "test" {
  app_id = okta_app_saml.test.id
}

resource "okta_app_signon_policy_rule" "test" {
  policy_id  = data.okta_app_signon_policy.test.id
  name       = "testAcc_replace_with_uuid"
  risk_score = "LOW"
  platform_include {
    os_expression = ""
    os_type       = "OTHER"
    type          = "DESKTOP"
  }
}

resource "okta_app_signon_policy_rules" "policy_rules" {
  policy_id = data.okta_app_signon_policy.test.id

  rule {
    # Replace with actual rule ID
    name               = "Rule1-updatedTF-23/02/26"
    priority           = 4
    status             = "ACTIVE"
    factor_mode        = "2FA"
    inactivity_period  = "PT1H"
    network_connection = "ANYWHERE"
  }

  rule {
    name               = "Rule2-updatedTF-23/02/26"
    priority           = 2
    status             = "ACTIVE"
    factor_mode        = "2FA"
    inactivity_period  = "PT1H"
    network_connection = "ANYWHERE"
  }

  rule {
    name               = "Rule3-updatedTF-23/02/26"
    priority           = 1
    status             = "ACTIVE"
    factor_mode        = "2FA"
    inactivity_period  = "PT1H"
    network_connection = "ANYWHERE"
  }

  rule {

    name               = "Rule4-updatedTF-23/02/26"
    priority           = 3
    status             = "ACTIVE"
    factor_mode        = "2FA"
    inactivity_period  = "PT1H"
    network_connection = "ANYWHERE"
  }

  rule {
    name               = "Rule5-updatedTF-23/02/26"
    priority           = 5
    status             = "ACTIVE"
    factor_mode        = "2FA"
    inactivity_period  = "PT1H"
    network_connection = "ANYWHERE"
    constraints = [jsonencode(
      {
        "authenticationMethods" : [
          {
            "key" : "okta_password",
            "method" : "password"
          }
        ],
        "next" : [{
          "authenticationMethods" : [{
            "key" : "okta_email",
            "method" : "email"
          }]
        }]
      }
    )]
  }
}
```

## Priority and system rules

- Okta evaluates rules in priority order (lower number = higher priority).
- Every policy includes a system *Catch-all Rule* that cannot be modified.
- Avoid managing the Catch-all Rule via Terraform.

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `policy_id` (String) ID of the policy to manage rules for.

### Optional

- `rule` (Block List) List of policy rules. Rules are processed in priority order (lowest number = highest priority). (see [below for nested schema](#nestedblock--rule))

### Read-Only

- `id` (String) The ID of this resource (same as policy_id).

<a id="nestedblock--rule"></a>
### Nested Schema for `rule`

Required:

- `name` (String) Policy Rule Name. Must be unique within the policy.

Optional:

- `access` (String) Access decision: ALLOW or DENY.
- `chains` (List of String) List of authentication method chain objects as JSON-encoded strings. Use with `type = "AUTH_METHOD_CHAIN"` only.
- `constraints` (List of String) List of authenticator constraints as JSON-encoded strings.
- `custom_expression` (String) Custom Okta Expression Language condition for advanced matching.
- `device_assurances_included` (Set of String) Set of device assurance policy IDs to include.
- `device_is_managed` (Boolean) Require device to be managed by a device management system.
- `device_is_registered` (Boolean) Require device to be registered with Okta Verify.
- `factor_mode` (String) Number of factors required: 1FA or 2FA.
- `groups_excluded` (Set of String) Set of group IDs to exclude from this rule.
- `groups_included` (Set of String) Set of group IDs to include in this rule.
- `id` (String) ID of the rule. Can be specified to adopt an existing rule during migration.
- `inactivity_period` (String) Inactivity period before re-authentication in ISO 8601 duration format.
- `network_connection` (String) Network selection mode: ANYWHERE, ZONE, ON_NETWORK, or OFF_NETWORK.
- `network_excludes` (List of String) List of network zone IDs to exclude.
- `network_includes` (List of String) List of network zone IDs to include.
- `platform_include` (Block List) Platform conditions to include. (see [below for nested schema](#nestedblock--rule--platform_include))
- `priority` (Number) Priority of the rule. Lower numbers are evaluated first.
- `re_authentication_frequency` (String) Re-authentication frequency in ISO 8601 duration format (e.g., PT2H for 2 hours). When using authentication chains with reauthenticateIn, this value is computed by the API based on the chain configuration.
- `risk_score` (String) Risk score level to match: ANY, LOW, MEDIUM, or HIGH.
- `status` (String) Status of the rule: ACTIVE or INACTIVE.
- `type` (String) Verification method type.
- `user_types_excluded` (Set of String) Set of user type IDs to exclude.
- `user_types_included` (Set of String) Set of user type IDs to include.
- `users_excluded` (Set of String) Set of user IDs to exclude from this rule.
- `users_included` (Set of String) Set of user IDs to include in this rule.

Read-Only:

- `system` (Boolean) Whether this is a system rule (e.g., Catch-all Rule). System rules cannot be modified.

<a id="nestedblock--rule--platform_include"></a>
### Nested Schema for `rule.platform_include`

Optional:

- `os_expression` (String) Custom OS expression for advanced matching.
- `os_type` (String) OS type: ANY, IOS, ANDROID, WINDOWS, OSX, MACOS, CHROMEOS, or OTHER.
- `type` (String) Platform type: ANY, MOBILE, or DESKTOP.
