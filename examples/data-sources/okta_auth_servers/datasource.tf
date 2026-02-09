resource "okta_auth_server" "test" {
  name        = "testAcc_replace_with_uuid"
  description = "test"
  audiences   = ["api://example"]
}

data "okta_auth_servers" "test" {
  q          = okta_auth_server.test.name
  depends_on = [okta_auth_server.test]
}
