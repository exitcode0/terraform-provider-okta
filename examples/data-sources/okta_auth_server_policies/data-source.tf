data "okta_auth_server" "default" {
  name = "default"
}

data "okta_auth_server_policies" "test" {
  auth_server_id = data.okta_auth_server.default.id
}
