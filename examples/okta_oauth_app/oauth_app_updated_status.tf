resource "okta_oauth_app" "testAcc_%[1]d" {
  status         = "INACTIVE"
  label          = "testAcc_%[1]d"
  type           = "web"
  grant_types    = ["implicit", "authorization_code"]
  redirect_uris  = ["http://d.com/"]
  response_types = ["code", "token", "id_token"]
}