resource "okta_realm" "test" {
  name = "testAcc_replace_with_uuid"
}

data "okta_realms" "test" {
  depends_on = [okta_realm.test]
}
