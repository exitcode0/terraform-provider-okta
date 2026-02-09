resource "okta_auth_server" "test" {
  name        = "testAcc_replace_with_uuid"
  description = "test"
  audiences   = ["api://example"]
}

resource "okta_auth_server_policy" "test" {
  auth_server_id   = okta_auth_server.test.id
  name             = "testAcc_replace_with_uuid"
  description      = "test"
  priority         = 1
  client_whitelist = ["ALL_CLIENTS"]
}

data "okta_auth_server_policies" "test" {
  auth_server_id = okta_auth_server.test.id
  depends_on     = [okta_auth_server_policy.test]
}
