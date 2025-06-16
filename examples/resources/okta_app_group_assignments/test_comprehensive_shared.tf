# Shared app configuration for comprehensive testing
resource "okta_app_bookmark" "shared" {
  label = "testAcc_shared_replace_with_uuid"
  url   = "https://test.com"
}

# Shared groups for all test scenarios
resource "okta_group" "test1" { name = "testAcc_replace_with_uuid_1" }
resource "okta_group" "test2" { name = "testAcc_replace_with_uuid_2" }
resource "okta_group" "test3" { name = "testAcc_replace_with_uuid_3" }
resource "okta_group" "test4" { name = "testAcc_replace_with_uuid_4" }