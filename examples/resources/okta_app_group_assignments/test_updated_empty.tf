resource "okta_app_bookmark" "test" {
  label = "testAcc_replace_with_uuid"
  url   = "https://test.com"
}
resource "okta_group" "test1" { name = "testAcc_replace_with_uuid_1" }
resource "okta_group" "test2" { name = "testAcc_replace_with_uuid_2" }
resource "okta_group" "test3" { name = "testAcc_replace_with_uuid_3" }

resource "okta_app_group_assignments" "test" {
  app_id = okta_app_bookmark.test.id
}
