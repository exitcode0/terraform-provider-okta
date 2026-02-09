resource "okta_app_oauth" "test" {
  label          = "testAcc_replace_with_uuid"
  type           = "web"
  grant_types    = ["authorization_code"]
  redirect_uris  = ["http://localhost:8080"]
  response_types = ["code"]
}

data "okta_app_tokens" "test" {
  app_id = okta_app_oauth.test.id
}
