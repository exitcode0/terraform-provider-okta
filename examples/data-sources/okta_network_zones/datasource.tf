resource "okta_network_zone" "test" {
  name     = "testAcc_replace_with_uuid"
  type     = "IP"
  gateways = ["1.2.3.4/24"]
  proxies  = ["2.2.3.4/24"]
}

data "okta_network_zones" "test" {
  depends_on = [okta_network_zone.test]
}
