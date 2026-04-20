package idaas_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/okta/terraform-provider-okta/okta/acctest"
)

func TestAccResourceOktaDevicePostureCheck_crud(t *testing.T) {
	mgr := newFixtureManager("resources", "okta_device_posture_check", t.Name())
	acctest.OktaResourceTest(t, resource.TestCase{
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactoriesForTestAcc(t),
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: mgr.ConfigReplace(`resource okta_device_posture_check test {
					name          = "testAcc-replace_with_uuid"
					platform      = "MACOS"
					query         = "SELECT 1 FROM disk_encryption WHERE encrypted = 1;"
					description   = "Check disk encryption is enabled"
					variable_name = "testAccDiskEncryptionreplace_with_uuid"
					mapping_type  = "CHECKBOX"
					remediation_settings = {
						message = {
							default_i18n_key = null
							custom_text      = "Please enable disk encryption."
						}
						link = {
							default_url = null
							custom_url  = "https://support.apple.com/en-us/HT204837"
						}
					}
				}`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("okta_device_posture_check.test", "name", fmt.Sprintf("testAcc-%d", mgr.Seed)),
					resource.TestCheckResourceAttr("okta_device_posture_check.test", "platform", "MACOS"),
					resource.TestCheckResourceAttr("okta_device_posture_check.test", "query", "SELECT 1 FROM disk_encryption WHERE encrypted = 1;"),
					resource.TestCheckResourceAttr("okta_device_posture_check.test", "description", "Check disk encryption is enabled"),
					resource.TestCheckResourceAttr("okta_device_posture_check.test", "mapping_type", "CHECKBOX"),
					resource.TestCheckResourceAttr("okta_device_posture_check.test", "remediation_settings.message.custom_text", "Please enable disk encryption."),
					resource.TestCheckResourceAttrSet("okta_device_posture_check.test", "id"),
					resource.TestCheckResourceAttrSet("okta_device_posture_check.test", "variable_name"),
					resource.TestCheckResourceAttrSet("okta_device_posture_check.test", "created_by"),
					resource.TestCheckResourceAttrSet("okta_device_posture_check.test", "created_date"),
				),
			},
			{
				Config: mgr.ConfigReplace(`resource okta_device_posture_check test {
					name          = "testAcc-replace_with_uuid"
					platform      = "MACOS"
					query         = "SELECT 1 FROM disk_encryption WHERE encrypted = 1 AND type = 'AES-XTS';"
					description   = "Check disk encryption with AES-XTS"
					variable_name = "testAccDiskEncryptionreplace_with_uuid"
					mapping_type  = "CHECKBOX"
					remediation_settings = {
						message = {
							default_i18n_key = null
							custom_text      = "Please enable FileVault disk encryption."
						}
						link = {
							default_url = null
							custom_url  = "https://support.apple.com/en-us/HT204837"
						}
					}
				}`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("okta_device_posture_check.test", "name", fmt.Sprintf("testAcc-%d", mgr.Seed)),
					resource.TestCheckResourceAttr("okta_device_posture_check.test", "query", "SELECT 1 FROM disk_encryption WHERE encrypted = 1 AND type = 'AES-XTS';"),
					resource.TestCheckResourceAttr("okta_device_posture_check.test", "description", "Check disk encryption with AES-XTS"),
					resource.TestCheckResourceAttr("okta_device_posture_check.test", "remediation_settings.message.custom_text", "Please enable FileVault disk encryption."),
					resource.TestCheckResourceAttr("okta_device_posture_check.test", "remediation_settings.link.custom_url", "https://support.apple.com/en-us/HT204837"),
				),
			},
			{
				ResourceName:      "okta_device_posture_check.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
