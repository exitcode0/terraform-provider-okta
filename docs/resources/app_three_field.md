---
page_title: "Resource: okta_app_three_field"
description: |-
  Creates a Three Field Application.
          This resource allows you to create and configure a Three Field Application.
          -> During an apply if there is change in 'status' the app will first be
          activated or deactivated in accordance with the 'status' change. Then, all
          other arguments that changed will be applied.
---

# Resource: okta_app_three_field

Creates a Three Field Application.
		This resource allows you to create and configure a Three Field Application.
		-> During an apply if there is change in 'status' the app will first be
		activated or deactivated in accordance with the 'status' change. Then, all
		other arguments that changed will be applied.

## Example Usage

```terraform
resource "okta_app_three_field" "example" {
  label                = "Example App"
  sign_on_url          = "https://example.com/login.html"
  sign_on_redirect_url = "https://example.com"
  reveal_password      = true
  credentials_scheme   = "EDIT_USERNAME_AND_PASSWORD"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `button_selector` (String) Login button field CSS selector
- `extra_field_selector` (String) Extra field CSS selector
- `extra_field_value` (String) Value for extra form field
- `label` (String) The Application's display name.
- `password_selector` (String) Login password field CSS selector
- `url` (String) Login URL
- `username_selector` (String) Login username field CSS selector

### Optional

- `accessibility_error_redirect_url` (String) Custom error page URL
- `accessibility_login_redirect_url` (String) Custom login page URL
- `accessibility_self_service` (Boolean) Enable self service. Default is `false`
- `admin_note` (String) Application notes for admins.
- `app_links_json` (String) Displays specific appLinks for the app. The value for each application link should be boolean.
- `auto_submit_toolbar` (Boolean) Display auto submit toolbar
- `credentials_scheme` (String) Application credentials scheme. One of: `EDIT_USERNAME_AND_PASSWORD`, `ADMIN_SETS_CREDENTIALS`, `EDIT_PASSWORD_ONLY`, `EXTERNAL_PASSWORD_SYNC`, or `SHARED_USERNAME_AND_PASSWORD`
- `enduser_note` (String) Application notes for end users.
- `hide_ios` (Boolean) Do not display application icon on mobile app
- `hide_web` (Boolean) Do not display application icon to users
- `logo` (String) Local file path to the logo. The file must be in PNG, JPG, or GIF format, and less than 1 MB in size.
- `reveal_password` (Boolean) Allow user to reveal password. It can not be set to `true` if `credentials_scheme` is `ADMIN_SETS_CREDENTIALS`, `SHARED_USERNAME_AND_PASSWORD` or `EXTERNAL_PASSWORD_SYNC`.
- `shared_password` (String) Shared password, required for certain schemes.
- `shared_username` (String) Shared username, required for certain schemes.
- `status` (String) Status of application. By default, it is `ACTIVE`
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- `url_regex` (String) A regex that further restricts URL to the specified regex
- `user_name_template` (String) Username template. Default: `${source.login}`
- `user_name_template_push_status` (String) Push username on update. Valid values: `PUSH` and `DONT_PUSH`
- `user_name_template_suffix` (String) Username template suffix
- `user_name_template_type` (String) Username template type. Default: `BUILT_IN`

### Read-Only

- `id` (String) The ID of this resource.
- `logo_url` (String) URL of the application's logo
- `name` (String) Name of the app.
- `sign_on_mode` (String) Sign on mode of application.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)
- `read` (String)
- `update` (String)

## Import

Import is supported using the following syntax:

```shell
terraform import okta_app_three_field.example <app_id>
```
