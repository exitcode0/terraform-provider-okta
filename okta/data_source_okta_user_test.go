package okta

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccOktaDataSourceUser_read(t *testing.T) {
	ri := acctest.RandInt()
	mgr := newFixtureManager(user)
	baseConfig := mgr.GetFixtures("datasource.tf", ri, t)

	// NOTE: The ACC tests on the datasource.tf can flap as sometimes these
	// tests can run faster than the Okta org becoming eventually consistent.
	//
	// TF_ACC=1 go test -tags unit -mod=readonly -test.v -run ^TestAccOktaDataSourceUser_read$
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProvidersFactories,
		Steps: []resource.TestStep{
			{
				Config: baseConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("okta_user.test", "id"),

					resource.TestCheckResourceAttrSet("data.okta_user.first_and_last", "id"),
					resource.TestCheckResourceAttr("data.okta_user.first_and_last", "first_name", "TestAcc"),
					resource.TestCheckResourceAttr("data.okta_user.first_and_last", "last_name", "Smith"),

					resource.TestCheckResourceAttrSet("data.okta_user.read_by_id", "user_id"),
					resource.TestCheckResourceAttrSet("data.okta_user.read_by_id", "admin_roles.#"),
					resource.TestCheckResourceAttr("data.okta_user.read_by_id", "first_name", "TestAcc"),
					resource.TestCheckResourceAttr("data.okta_user.read_by_id", "last_name", "Smith"),
					resource.TestCheckTypeSetElemAttr("data.okta_user.read_by_id", "admin_roles.*", "READ_ONLY_ADMIN"),

					resource.TestCheckResourceAttrSet("data.okta_user.read_by_id_with_skip", "id"),
					resource.TestCheckResourceAttrSet("data.okta_user.read_by_id_with_skip", "skip_groups"),
					resource.TestCheckResourceAttrSet("data.okta_user.read_by_id_with_skip", "skip_roles"),
					resource.TestCheckResourceAttr("data.okta_user.read_by_id_with_skip", "skip_groups", "true"),
					resource.TestCheckResourceAttr("data.okta_user.read_by_id_with_skip", "skip_roles", "true"),
					resource.TestCheckNoResourceAttr("data.okta_user.read_by_id_with_skip", "groups"),
					resource.TestCheckNoResourceAttr("data.okta_user.read_by_id_with_skip", "admin_roles.#"),

					resource.TestCheckResourceAttrSet("data.okta_user.compound_search", "id"),
					resource.TestCheckResourceAttrSet("data.okta_user.compound_search", "compound_search_operator"),
					resource.TestCheckResourceAttr("data.okta_user.compound_search", "compound_search_operator", "or"),
					resource.TestCheckResourceAttr("data.okta_user.compound_search", "first_name", "Some"),
					resource.TestCheckResourceAttr("data.okta_user.compound_search", "last_name", "One"),

					resource.TestCheckResourceAttrSet("data.okta_user.expression_search", "id"),
					resource.TestCheckResourceAttr("data.okta_user.expression_search", "first_name", "Some"),
					resource.TestCheckResourceAttr("data.okta_user.expression_search", "last_name", "One"),
					resource.TestCheckResourceAttr("data.okta_user.expression_search", "custom_profile_attributes", `{"array123":["cool","feature"]}`),
				),
			},
		},
	})
}
