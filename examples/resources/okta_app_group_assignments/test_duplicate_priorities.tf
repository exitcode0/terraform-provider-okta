resource "okta_app_bookmark" "test" {
  label = "testAcc_replace_with_uuid"
  url   = "https://test.com"
}
resource "okta_group" "test1" { name = "testAcc_replace_with_uuid_1" }
resource "okta_group" "test2" { name = "testAcc_replace_with_uuid_2" }
resource "okta_group" "test3" { name = "testAcc_replace_with_uuid_3" }

# Test priority gaps - let Okta re-sequence them
resource "okta_app_group_assignments" "test" {
  app_id = okta_app_bookmark.test.id

  group {
    id       = okta_group.test1.id
    priority = 1
    profile  = jsonencode({})
  }
  group {
    id       = okta_group.test2.id
    priority = 5 # Gap in sequence - Okta might re-sequence
    profile  = jsonencode({})
  }
  group {
    id       = okta_group.test3.id
    priority = 10 # Another gap - Okta might re-sequence
    profile  = jsonencode({})
  }
}
