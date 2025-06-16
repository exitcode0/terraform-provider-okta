resource "okta_app_oauth" "test" {
  label          = "testAcc_replace_with_uuid"
  type           = "native"
  grant_types    = ["authorization_code"]
  redirect_uris  = ["http://d.com/"]
  response_types = ["code"]
}
resource "okta_group" "test1" { name = "testAcc_replace_with_uuid_1" }
resource "okta_group" "test2" { name = "testAcc_replace_with_uuid_2" }
resource "okta_group" "test3" { name = "testAcc_replace_with_uuid_3" }

# Test profile changes with priority handling
resource "okta_app_group_assignments" "test" {
  app_id = okta_app_oauth.test.id

  group {
    id       = okta_group.test1.id
    priority = 1
    profile  = jsonencode({})
  }
  group {
    id       = okta_group.test2.id
    priority = 2
    profile  = jsonencode({})
  }
  group {
    id = okta_group.test3.id
    # No priority - should be unset
    profile = jsonencode({})
  }
}
