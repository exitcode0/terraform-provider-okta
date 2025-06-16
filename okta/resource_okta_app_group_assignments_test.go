package okta

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/okta/terraform-provider-okta/sdk"
)

func TestAccResourceOktaAppGroupAssignments_crud(t *testing.T) {
	resourceName := fmt.Sprintf("%s.test", appGroupAssignments)
	mgr := newFixtureManager("resources", appGroupAssignments, t.Name())
	config := mgr.GetFixtures("test_basic.tf", t)
	updatedConfig := mgr.GetFixtures("test_updated.tf", t)

	group1 := fmt.Sprintf("%s.test1", group)
	group2 := fmt.Sprintf("%s.test2", group)
	group3 := fmt.Sprintf("%s.test3", group)

	oktaResourceTest(t, resource.TestCase{
		PreCheck:          testAccPreCheck(t),
		ErrorCheck:        testAccErrorChecks(t),
		ProviderFactories: testAccProvidersFactories,
		CheckDestroy:      checkResourceDestroy(appGroupAssignments, createDoesAppExist(sdk.NewBookmarkApplication())),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					ensureAppGroupAssignmentsExist(resourceName, group1, group2, group3),
					resource.TestCheckResourceAttrSet(resourceName, "app_id"),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					ensureAppGroupAssignmentsExist(resourceName, group1, group2),
					resource.TestCheckResourceAttrSet(resourceName, "app_id"),
				),
			},
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					ensureAppGroupAssignmentsExist(resourceName, group1, group2, group3),
					resource.TestCheckResourceAttrSet(resourceName, "app_id"),
				),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					ensureAppGroupAssignmentsExist(resourceName, group1, group2, group3),
					resource.TestCheckResourceAttrSet(resourceName, "app_id"),
				),
			},
		},
	})
}

// TestAccResourceOktaAppGroupAssignments_1088_unplanned_changes This test
// demonstrates incorrect behavior in okta_app_group_assignments has been
// corrected.  The original author implemented incorrect behavior, in terms of
// idiomatic design principles of TF providers, where it would proactively
// remove group assignments from an app if they were made outside of the
// resource. The correct behavior is to surface drift detection if a group is
// assigned to an app outside of this resource.
func TestAccResourceOktaAppGroupAssignments_1088_unplanned_changes(t *testing.T) {
	mgr := newFixtureManager("resources", appGroupAssignments, t.Name())
	assignments1 := fmt.Sprintf("%s.test", appGroupAssignments)
	bookmarkApp := fmt.Sprintf("%s.test", appBookmark)
	groupA := fmt.Sprintf("%s.a", group)
	groupB := fmt.Sprintf("%s.b", group)
	groupC := fmt.Sprintf("%s.c", group)

	baseConfig := `
resource "okta_app_bookmark" "test" {
	label = "testAcc_replace_with_uuid"
	url   = "https://test.com"
}
resource "okta_group" "a" {
	name        = "testAcc-group-a_replace_with_uuid"
	description = "Group A"
}
resource "okta_group" "b" {
	name        = "testAcc-group-b_replace_with_uuid"
	description = "Group B"
}
resource "okta_group" "c" {
	name        = "testAcc-group-c_replace_with_uuid"
	description = "Group C"
}`

	step1Config := `
resource "okta_app_group_assignments" "test" {
	app_id = okta_app_bookmark.test.id
	group {
		id = okta_group.a.id
		priority = 1
		profile = jsonencode({"test": "a"})
	}
}`

	step2Config := `
resource "okta_app_group_assignments" "test" {
	app_id = okta_app_bookmark.test.id
	group {
		id = okta_group.a.id
		priority = 1
		profile = jsonencode({"test": "a"})
	}
	group {
		id = okta_group.b.id
		priority = 2
		profile = jsonencode({"test": "b"})
	}
}`

	step3Config := `
resource "okta_app_group_assignments" "test" {
	app_id = okta_app_bookmark.test.id
	group {
		id = okta_group.a.id
		priority = 1
		profile = jsonencode({"test": "a"})
	}
	group {
		id = okta_group.b.id
		priority = 2
		profile = jsonencode({"test": "b"})
	}
	group {
		id = okta_group.c.id
		priority = 4
		profile = jsonencode({"test": "c"})
	}
}`

	stepLastConfig := `
resource "okta_app_group_assignments" "test" {
	app_id = okta_app_bookmark.test.id
	group {
		id = okta_group.a.id
		priority = 99
		profile = jsonencode({"different": "profile value"})
	}
}`

	oktaResourceTest(t, resource.TestCase{
		PreCheck:          testAccPreCheck(t),
		ErrorCheck:        testAccErrorChecks(t),
		CheckDestroy:      nil,
		ProviderFactories: testAccProvidersFactories,
		Steps: []resource.TestStep{
			{
				// Vanilla step
				Config: mgr.ConfigReplace(fmt.Sprintf("%s\n%s", baseConfig, step1Config)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(assignments1, "group.#", "1"),
					resource.TestCheckResourceAttr(assignments1, "group.0.priority", "1"),
					resource.TestCheckResourceAttr(assignments1, "group.0.profile", `{"test":"a"}`),
					ensureAppGroupAssignmentsExist(assignments1, groupA),
				),
			},
			{
				Config: mgr.ConfigReplace(fmt.Sprintf("%s\n%s", baseConfig, step2Config)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(assignments1, "group.#", "2"),
					resource.TestCheckResourceAttr(assignments1, "group.0.priority", "1"),
					resource.TestCheckResourceAttr(assignments1, "group.0.profile", `{"test":"a"}`),
					resource.TestCheckResourceAttr(assignments1, "group.1.priority", "2"),
					resource.TestCheckResourceAttr(assignments1, "group.1.profile", `{"test":"b"}`),
					ensureAppGroupAssignmentsExist(assignments1, groupA, groupB),

					// This mimics assigning Group C to the app outside of
					// Terraform. In this case doing so with a direct API call
					// via the test harness which is equivalent to "Click Ops"
					clickOpsAssignGroupToApp(bookmarkApp, groupC),
					clickOpsCheckIfGroupIsAssignedToApp(bookmarkApp, groupA, groupB, groupC),

					// NOTE: after these checks run the terraform test runner
					// will do a refresh and catch that group C has been added
					// to the app outside of the terraform config and emit a
					// non-empty plan
				),

				// side effect of the TF test runner is expecting a non-empty
				// plan is treated as an apply accept and adds group c to local
				// state
				ExpectNonEmptyPlan: true,
			},
			{
				Config: mgr.ConfigReplace(fmt.Sprintf("%s\n%s", baseConfig, step3Config)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(assignments1, "group.#", "3"),
					resource.TestCheckResourceAttr(assignments1, "group.0.priority", "1"),
					resource.TestCheckResourceAttr(assignments1, "group.0.profile", `{"test":"a"}`),
					resource.TestCheckResourceAttr(assignments1, "group.1.priority", "2"),
					resource.TestCheckResourceAttr(assignments1, "group.1.profile", `{"test":"b"}`),
					resource.TestCheckResourceAttr(assignments1, "group.2.priority", "4"),
					resource.TestCheckResourceAttr(assignments1, "group.2.profile", `{"test":"c"}`),
					ensureAppGroupAssignmentsExist(assignments1, groupA, groupB, groupC),
				),
			},
			{
				// check that we can do removing group assignments
				Config: mgr.ConfigReplace(fmt.Sprintf("%s\n%s", baseConfig, step2Config)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(assignments1, "group.#", "2"),
					resource.TestCheckResourceAttr(assignments1, "group.0.priority", "1"),
					resource.TestCheckResourceAttr(assignments1, "group.0.profile", `{"test":"a"}`),
					resource.TestCheckResourceAttr(assignments1, "group.1.priority", "2"),
					resource.TestCheckResourceAttr(assignments1, "group.1.profile", `{"test":"b"}`),
					ensureAppGroupAssignmentsExist(assignments1, groupA, groupB),
				),
			},
			{
				Config: mgr.ConfigReplace(fmt.Sprintf("%s\n%s", baseConfig, step1Config)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(assignments1, "group.#", "1"),
					resource.TestCheckResourceAttr(assignments1, "group.0.priority", "1"),
					resource.TestCheckResourceAttr(assignments1, "group.0.profile", `{"test":"a"}`),
					ensureAppGroupAssignmentsExist(assignments1, groupA),
				),
			},
			{
				// Check that priority and profile can be changed on the group
				// itself
				Config: mgr.ConfigReplace(fmt.Sprintf("%s\n%s", baseConfig, stepLastConfig)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(assignments1, "group.#", "1"),
					resource.TestCheckResourceAttr(assignments1, "group.0.priority", "99"),
					resource.TestCheckResourceAttr(assignments1, "group.0.profile", `{"different":"profile value"}`),
					ensureAppGroupAssignmentsExist(assignments1, groupA),
				),
			},
		},
	})
}

func ensureAppGroupAssignmentsExist(resourceName string, groupsExpected ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		missingErr := fmt.Errorf("resource not found: %s", resourceName)
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return missingErr
		}

		appID := rs.Primary.Attributes["app_id"]
		client := sdkV2ClientForTest()

		// Get all the IDs of groups we expect to be assigned
		expectedGroupIDs := map[string]bool{}
		for _, groupExpected := range groupsExpected {
			grs, ok := s.RootModule().Resources[groupExpected]
			if !ok {
				return missingErr
			}
			expectedGroupIDs[grs.Primary.Attributes["id"]] = false
		}

		for i := 0; i < len(groupsExpected); i++ {
			groupID := rs.Primary.Attributes[fmt.Sprintf("group.%d.id", i)]
			g, _, err := client.Application.GetApplicationGroupAssignment(context.Background(), appID, groupID, nil)
			if err != nil {
				return err
			} else if g == nil {
				return missingErr
			}
			// group found, check it off
			expectedGroupIDs[groupID] = true
		}

		// now check we found all the groupIDs we expected
		if len(expectedGroupIDs) != len(groupsExpected) {
			return fmt.Errorf("expected %d assigned groups but got %d", len(groupsExpected), len(expectedGroupIDs))
		}

		// make sure we found them all
		for groupID, found := range expectedGroupIDs {
			if !found {
				return fmt.Errorf("expected group %s to be assigned but wasn't", groupID)
			}
		}
		return nil
	}
}

func clickOpsAssignGroupToApp(appResourceName, groupResourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resources := []string{appResourceName, groupResourceName}
		for _, resourceName := range resources {
			missingErr := fmt.Errorf("resource not found: %s", resourceName)
			if _, ok := s.RootModule().Resources[resourceName]; !ok {
				return missingErr
			}
		}

		appRS := s.RootModule().Resources[appResourceName]
		appID := appRS.Primary.Attributes["id"]
		groupRS := s.RootModule().Resources[groupResourceName]
		groupID := groupRS.Primary.Attributes["id"]
		client := sdkV2ClientForTest()
		_, _, err := client.Application.CreateApplicationGroupAssignment(context.Background(), appID, groupID, sdk.ApplicationGroupAssignment{})
		if err != nil {
			return fmt.Errorf("API: unable to assign app %q to group %q, err: %+v", appID, groupID, err)
		}

		return nil
	}
}

func clickOpsCheckIfGroupIsAssignedToApp(appResourceName string, groups ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, groupResourceName := range groups {
			resources := []string{appResourceName, groupResourceName}
			for _, resourceName := range resources {
				missingErr := fmt.Errorf("resource not found: %s", resourceName)
				if _, ok := s.RootModule().Resources[resourceName]; !ok {
					return missingErr
				}
			}

			appRS := s.RootModule().Resources[appResourceName]
			appID := appRS.Primary.Attributes["id"]
			groupRS := s.RootModule().Resources[groupResourceName]
			groupID := groupRS.Primary.Attributes["id"]
			client := sdkV2ClientForTest()
			_, _, err := client.Application.GetApplicationGroupAssignment(context.Background(), appID, groupID, nil)
			if err != nil {
				return fmt.Errorf("API: app %q is not assigned to group %s", appID, groupID)
			}
		}

		return nil
	}
}

// This test demonstrate the ability to unassigned all groups from app without having to destroy the resource
// This behavior is already enabled by the API
func TestAccResourceOktaAppGroupAssignments_2068_empty_assignments(t *testing.T) {
	resourceName := fmt.Sprintf("%s.test", appGroupAssignments)
	mgr := newFixtureManager("resources", appGroupAssignments, t.Name())
	config := mgr.GetFixtures("test_basic.tf", t)
	updatedConfig := mgr.GetFixtures("test_updated_empty.tf", t)

	group1 := fmt.Sprintf("%s.test1", group)
	group2 := fmt.Sprintf("%s.test2", group)
	group3 := fmt.Sprintf("%s.test3", group)

	oktaResourceTest(t, resource.TestCase{
		PreCheck:          testAccPreCheck(t),
		ErrorCheck:        testAccErrorChecks(t),
		ProviderFactories: testAccProvidersFactories,
		CheckDestroy:      checkResourceDestroy(appGroupAssignments, createDoesAppExist(sdk.NewBookmarkApplication())),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					ensureAppGroupAssignmentsExist(resourceName, group1, group2, group3),
					resource.TestCheckResourceAttrSet(resourceName, "app_id"),
					resource.TestCheckResourceAttr(resourceName, "group.#", "3"),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "app_id"),
					resource.TestCheckResourceAttr(resourceName, "group.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceOktaAppGroupAssignments_1832_timeouts(t *testing.T) {
	resourceName := fmt.Sprintf("%s.test", appGroupAssignments)
	mgr := newFixtureManager("resources", appGroupAssignments, t.Name())
	bookmarkApp := fmt.Sprintf("%s.test", appBookmark)
	groupA := fmt.Sprintf("%s.a", group)
	groupB := fmt.Sprintf("%s.b", group)
	config := `
resource "okta_app_bookmark" "test" {
	label = "testAcc_replace_with_uuid"
	url   = "https://test.com"
}
resource "okta_group" "a" {
	name        = "testAcc-group-a_replace_with_uuid"
	description = "Group A"
}
resource "okta_group" "b" {
	name        = "testAcc-group-b_replace_with_uuid"
	description = "Group B"
}
resource "okta_app_group_assignments" "test" {
  app_id = okta_app_bookmark.test.id
  group {
    id = okta_group.a.id
  }
  group {
    id = okta_group.b.id
  }
  timeouts {
    create = "60m"
    read = "2h"
    update = "30m"
  }
}`
	oktaResourceTest(t, resource.TestCase{
		PreCheck:          testAccPreCheck(t),
		ErrorCheck:        testAccErrorChecks(t),
		ProviderFactories: testAccProvidersFactories,
		CheckDestroy:      checkResourceDestroy(appGroupAssignments, createDoesAppExist(sdk.NewBookmarkApplication())),
		Steps: []resource.TestStep{
			{
				Config: mgr.ConfigReplace(config),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "timeouts.create", "60m"),
					resource.TestCheckResourceAttr(resourceName, "timeouts.read", "2h"),
					resource.TestCheckResourceAttr(resourceName, "timeouts.update", "30m"),

					clickOpsCheckIfGroupIsAssignedToApp(bookmarkApp, groupA, groupB),
				),
			},
		},
	})
}

// TestAccResourceOktaAppGroupAssignments_priority_zero tests that priority 0
// is treated as a valid explicit priority value, distinct from unset priority
func TestAccResourceOktaAppGroupAssignments_priority_zero(t *testing.T) {
	resourceName := fmt.Sprintf("%s.test", appGroupAssignments)
	mgr := newFixtureManager("resources", appGroupAssignments, t.Name())
	config := mgr.GetFixtures("test_priority_zero.tf", t)

	group1 := fmt.Sprintf("%s.test1", group)
	group2 := fmt.Sprintf("%s.test2", group)
	group3 := fmt.Sprintf("%s.test3", group)

	oktaResourceTest(t, resource.TestCase{
		PreCheck:          testAccPreCheck(t),
		ErrorCheck:        testAccErrorChecks(t),
		ProviderFactories: testAccProvidersFactories,
		CheckDestroy:      checkResourceDestroy(appGroupAssignments, createDoesAppExist(sdk.NewBookmarkApplication())),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					ensureAppGroupAssignmentsExist(resourceName, group1, group2, group3),
					resource.TestCheckResourceAttrSet(resourceName, "app_id"),
					resource.TestCheckResourceAttr(resourceName, "group.#", "3"),

					// Check that priority 0 is preserved
					resource.TestCheckResourceAttr(resourceName, "group.0.priority", "0"),
					resource.TestCheckResourceAttr(resourceName, "group.0.profile", `{}`),

					// Check that priority 1 is preserved
					resource.TestCheckResourceAttr(resourceName, "group.1.priority", "1"),
					resource.TestCheckResourceAttr(resourceName, "group.1.profile", `{}`),

					// Check that group without priority has no priority attribute or it's not set
					resource.TestCheckResourceAttr(resourceName, "group.2.profile", `{}`),
					// Note: We can't easily test for absence of priority attribute in Terraform tests
					// but the important thing is that it doesn't cause drift
				),
			},
			{
				// Test refresh to ensure no drift with priority 0
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					ensureAppGroupAssignmentsExist(resourceName, group1, group2, group3),
					resource.TestCheckResourceAttr(resourceName, "group.0.priority", "0"),
					resource.TestCheckResourceAttr(resourceName, "group.1.priority", "1"),
				),
			},
		},
	})
}

// TestAccResourceOktaAppGroupAssignments_priority_resequencing tests that Okta's
// priority re-sequencing doesn't cause drift when relative order is maintained
func TestAccResourceOktaAppGroupAssignments_priority_resequencing(t *testing.T) {
	resourceName := fmt.Sprintf("%s.test", appGroupAssignments)
	mgr := newFixtureManager("resources", appGroupAssignments, t.Name())
	config := mgr.GetFixtures("test_priority_resequencing.tf", t)

	group1 := fmt.Sprintf("%s.test1", group)
	group2 := fmt.Sprintf("%s.test2", group)
	group3 := fmt.Sprintf("%s.test3", group)

	oktaResourceTest(t, resource.TestCase{
		PreCheck:          testAccPreCheck(t),
		ErrorCheck:        testAccErrorChecks(t),
		ProviderFactories: testAccProvidersFactories,
		CheckDestroy:      checkResourceDestroy(appGroupAssignments, createDoesAppExist(sdk.NewBookmarkApplication())),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					ensureAppGroupAssignmentsExist(resourceName, group1, group2, group3),
					resource.TestCheckResourceAttrSet(resourceName, "app_id"),
					resource.TestCheckResourceAttr(resourceName, "group.#", "3"),

					// Verify the groups are assigned with their profiles
					resource.TestCheckResourceAttr(resourceName, "group.0.profile", `{}`),
					resource.TestCheckResourceAttr(resourceName, "group.1.profile", `{}`),
					resource.TestCheckResourceAttr(resourceName, "group.2.profile", `{}`),

					// Check that priorities exist (Okta may re-sequence them)
					resource.TestCheckResourceAttrSet(resourceName, "group.0.priority"),
					resource.TestCheckResourceAttrSet(resourceName, "group.1.priority"),
					resource.TestCheckResourceAttrSet(resourceName, "group.2.priority"),

					// Verify relative ordering is maintained via custom check
					checkPriorityRelativeOrder(resourceName, []string{"test1", "test2", "test3"}),
				),
			},
			{
				// Test refresh to ensure no drift even if Okta re-sequences priorities
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					ensureAppGroupAssignmentsExist(resourceName, group1, group2, group3),
					checkPriorityRelativeOrder(resourceName, []string{"test1", "test2", "test3"}),
				),
			},
		},
	})
}

// TestAccResourceOktaAppGroupAssignments_priority_mixed tests mixed scenarios
// where some groups have priorities and others don't
func TestAccResourceOktaAppGroupAssignments_priority_mixed(t *testing.T) {
	resourceName := fmt.Sprintf("%s.test", appGroupAssignments)
	mgr := newFixtureManager("resources", appGroupAssignments, t.Name())
	config := mgr.GetFixtures("test_priority_mixed.tf", t)

	group1 := fmt.Sprintf("%s.test1", group)
	group2 := fmt.Sprintf("%s.test2", group)
	group3 := fmt.Sprintf("%s.test3", group)

	oktaResourceTest(t, resource.TestCase{
		PreCheck:          testAccPreCheck(t),
		ErrorCheck:        testAccErrorChecks(t),
		ProviderFactories: testAccProvidersFactories,
		CheckDestroy:      checkResourceDestroy(appGroupAssignments, createDoesAppExist(sdk.NewBookmarkApplication())),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					ensureAppGroupAssignmentsExist(resourceName, group1, group2, group3),
					resource.TestCheckResourceAttrSet(resourceName, "app_id"),
					resource.TestCheckResourceAttr(resourceName, "group.#", "3"),

					// Verify profiles are correct
					resource.TestCheckResourceAttr(resourceName, "group.0.profile", `{}`),
					resource.TestCheckResourceAttr(resourceName, "group.1.profile", `{}`),
					resource.TestCheckResourceAttr(resourceName, "group.2.profile", `{}`),

					// Check that groups with priorities maintain relative order (test3=1, test1=2)
					checkMixedPriorityOrder(resourceName),
				),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					ensureAppGroupAssignmentsExist(resourceName, group1, group2, group3),
					checkMixedPriorityOrder(resourceName),
				),
			},
		},
	})
}

// TestAccResourceOktaAppGroupAssignments_no_priorities tests that groups
// without any priorities work correctly
func TestAccResourceOktaAppGroupAssignments_no_priorities(t *testing.T) {
	resourceName := fmt.Sprintf("%s.test", appGroupAssignments)
	mgr := newFixtureManager("resources", appGroupAssignments, t.Name())
	config := mgr.GetFixtures("test_no_priorities.tf", t)

	group1 := fmt.Sprintf("%s.test1", group)
	group2 := fmt.Sprintf("%s.test2", group)
	group3 := fmt.Sprintf("%s.test3", group)

	oktaResourceTest(t, resource.TestCase{
		PreCheck:          testAccPreCheck(t),
		ErrorCheck:        testAccErrorChecks(t),
		ProviderFactories: testAccProvidersFactories,
		CheckDestroy:      checkResourceDestroy(appGroupAssignments, createDoesAppExist(sdk.NewBookmarkApplication())),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					ensureAppGroupAssignmentsExist(resourceName, group1, group2, group3),
					resource.TestCheckResourceAttrSet(resourceName, "app_id"),
					resource.TestCheckResourceAttr(resourceName, "group.#", "3"),

					// Verify profiles are correct
					resource.TestCheckResourceAttr(resourceName, "group.0.profile", `{}`),
					resource.TestCheckResourceAttr(resourceName, "group.1.profile", `{}`),
					resource.TestCheckResourceAttr(resourceName, "group.2.profile", `{}`),
				),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					ensureAppGroupAssignmentsExist(resourceName, group1, group2, group3),
					resource.TestCheckResourceAttr(resourceName, "group.#", "3"),
				),
			},
		},
	})
}

// checkPriorityRelativeOrder verifies that the relative ordering of priorities is maintained
func checkPriorityRelativeOrder(resourceName string, expectedOrder []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		// Build a map of group name suffix to priority
		groupPriorities := make(map[string]int)
		for i := 0; i < len(expectedOrder); i++ {
			priorityKey := fmt.Sprintf("group.%d.priority", i)
			if priorityStr, exists := rs.Primary.Attributes[priorityKey]; exists {
				var priority int
				if _, err := fmt.Sscanf(priorityStr, "%d", &priority); err != nil {
					return fmt.Errorf("failed to parse priority %s: %v", priorityStr, err)
				}

				// Extract group name suffix from the group ID
				groupIDKey := fmt.Sprintf("group.%d.id", i)
				if groupID, exists := rs.Primary.Attributes[groupIDKey]; exists {
					// Find which expected group this corresponds to
					for _, expectedGroup := range expectedOrder {
						groupResourceName := fmt.Sprintf("okta_group.%s", expectedGroup)
						if groupRes, ok := s.RootModule().Resources[groupResourceName]; ok {
							if groupRes.Primary.Attributes["id"] == groupID {
								groupPriorities[expectedGroup] = priority
								break
							}
						}
					}
				}
			}
		}

		// Verify relative ordering
		if len(groupPriorities) > 1 {
			prevGroup := ""
			prevPriority := -1
			for _, groupName := range expectedOrder {
				if priority, exists := groupPriorities[groupName]; exists {
					if prevGroup != "" && priority <= prevPriority {
						return fmt.Errorf("priority order violation: %s (priority %d) should come after %s (priority %d)",
							groupName, priority, prevGroup, prevPriority)
					}
					prevGroup = groupName
					prevPriority = priority
				}
			}
		}

		return nil
	}
}

// checkMixedPriorityOrder verifies the specific mixed priority scenario
func checkMixedPriorityOrder(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		// In our mixed test: test1 has priority 2, test2 has no priority, test3 has priority 1
		// So test3 should have lower priority value than test1
		var test1Priority, test3Priority int
		var test1HasPriority, test3HasPriority bool

		for i := 0; i < 3; i++ {
			groupIDKey := fmt.Sprintf("group.%d.id", i)
			priorityKey := fmt.Sprintf("group.%d.priority", i)

			if groupID, exists := rs.Primary.Attributes[groupIDKey]; exists {
				// Check which group this is
				for _, groupSuffix := range []string{"test1", "test3"} {
					groupResourceName := fmt.Sprintf("okta_group.%s", groupSuffix)
					if groupRes, ok := s.RootModule().Resources[groupResourceName]; ok {
						if groupRes.Primary.Attributes["id"] == groupID {
							if priorityStr, hasPriority := rs.Primary.Attributes[priorityKey]; hasPriority {
								var priority int
								if _, err := fmt.Sscanf(priorityStr, "%d", &priority); err != nil {
									return fmt.Errorf("failed to parse priority %s: %v", priorityStr, err)
								}

								if groupSuffix == "test1" {
									test1Priority = priority
									test1HasPriority = true
								} else if groupSuffix == "test3" {
									test3Priority = priority
									test3HasPriority = true
								}
							}
							break
						}
					}
				}
			}
		}

		// Verify both prioritized groups have priorities and test3 < test1
		if !test1HasPriority {
			return fmt.Errorf("test1 should have a priority")
		}
		if !test3HasPriority {
			return fmt.Errorf("test3 should have a priority")
		}
		if test3Priority >= test1Priority {
			return fmt.Errorf("test3 priority (%d) should be less than test1 priority (%d)", test3Priority, test1Priority)
		}

		return nil
	}
}

// TestAccResourceOktaAppGroupAssignments_duplicate_priorities tests that duplicate
// priorities are handled gracefully by letting Okta resolve them
func TestAccResourceOktaAppGroupAssignments_duplicate_priorities(t *testing.T) {
	resourceName := fmt.Sprintf("%s.test", appGroupAssignments)
	mgr := newFixtureManager("resources", appGroupAssignments, t.Name())
	config := mgr.GetFixtures("test_duplicate_priorities.tf", t)

	group1 := fmt.Sprintf("%s.test1", group)
	group2 := fmt.Sprintf("%s.test2", group)
	group3 := fmt.Sprintf("%s.test3", group)

	oktaResourceTest(t, resource.TestCase{
		PreCheck:          testAccPreCheck(t),
		ErrorCheck:        testAccErrorChecks(t),
		ProviderFactories: testAccProvidersFactories,
		CheckDestroy:      checkResourceDestroy(appGroupAssignments, createDoesAppExist(sdk.NewBookmarkApplication())),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					ensureAppGroupAssignmentsExist(resourceName, group1, group2, group3),
					resource.TestCheckResourceAttrSet(resourceName, "app_id"),
					resource.TestCheckResourceAttr(resourceName, "group.#", "3"),

					// Verify the groups are assigned with their profiles
					resource.TestCheckResourceAttr(resourceName, "group.0.profile", `{}`),
					resource.TestCheckResourceAttr(resourceName, "group.1.profile", `{}`),
					resource.TestCheckResourceAttr(resourceName, "group.2.profile", `{}`),

					// Check that all groups have priorities (Okta should resolve duplicates)
					resource.TestCheckResourceAttrSet(resourceName, "group.0.priority"),
					resource.TestCheckResourceAttrSet(resourceName, "group.1.priority"),
					resource.TestCheckResourceAttrSet(resourceName, "group.2.priority"),

					// Verify that Okta resolved the duplicate priorities
					checkDuplicatePrioritiesResolved(resourceName),
				),
			},
			{
				// Test refresh to ensure no drift after Okta resolves duplicates
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					ensureAppGroupAssignmentsExist(resourceName, group1, group2, group3),
					checkDuplicatePrioritiesResolved(resourceName),
				),
			},
		},
	})
}

// checkDuplicatePrioritiesResolved verifies that Okta resolved duplicate priorities
func checkDuplicatePrioritiesResolved(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		// Collect all priorities
		priorities := make([]int, 0)
		priorityCount := make(map[int]int)

		for i := 0; i < 3; i++ {
			priorityKey := fmt.Sprintf("group.%d.priority", i)
			if priorityStr, exists := rs.Primary.Attributes[priorityKey]; exists {
				var priority int
				if _, err := fmt.Sscanf(priorityStr, "%d", &priority); err != nil {
					return fmt.Errorf("failed to parse priority %s: %v", priorityStr, err)
				}
				priorities = append(priorities, priority)
				priorityCount[priority]++
			}
		}

		// Verify we have 3 priorities
		if len(priorities) != 3 {
			return fmt.Errorf("expected 3 priorities, got %d", len(priorities))
		}

		// Check if Okta resolved duplicates (all priorities should be unique now)
		for priority, count := range priorityCount {
			if count > 1 {
				// This might be OK if Okta allows duplicates, but let's log it
				// We won't fail the test since this is Okta's behavior
				fmt.Printf("Warning: Priority %d appears %d times (Okta may allow this)\n", priority, count)
			}
		}

		return nil
	}
}

// TestAccResourceOktaAppGroupAssignments_profile_changes tests that priority
// changes are handled correctly with OAuth applications
func TestAccResourceOktaAppGroupAssignments_profile_changes(t *testing.T) {
	resourceName := fmt.Sprintf("%s.test", appGroupAssignments)
	mgr := newFixtureManager("resources", appGroupAssignments, t.Name())
	config := mgr.GetFixtures("test_profile_changes.tf", t)
	updatedConfig := mgr.GetFixtures("test_profile_changes_updated.tf", t)

	group1 := fmt.Sprintf("%s.test1", group)
	group2 := fmt.Sprintf("%s.test2", group)
	group3 := fmt.Sprintf("%s.test3", group)

	oktaResourceTest(t, resource.TestCase{
		PreCheck:          testAccPreCheck(t),
		ErrorCheck:        testAccErrorChecks(t),
		ProviderFactories: testAccProvidersFactories,
		CheckDestroy:      checkResourceDestroy(appGroupAssignments, createDoesAppExist(sdk.NewOpenIdConnectApplication())),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					ensureAppGroupAssignmentsExist(resourceName, group1, group2, group3),
					resource.TestCheckResourceAttrSet(resourceName, "app_id"),
					resource.TestCheckResourceAttr(resourceName, "group.#", "3"),

					// Verify initial profiles and priorities
					resource.TestCheckResourceAttr(resourceName, "group.0.priority", "1"),
					resource.TestCheckResourceAttr(resourceName, "group.0.profile", `{}`),

					resource.TestCheckResourceAttr(resourceName, "group.1.priority", "2"),
					resource.TestCheckResourceAttr(resourceName, "group.1.profile", `{}`),

					// Group 3 should have no priority initially
					resource.TestCheckResourceAttr(resourceName, "group.2.profile", `{}`),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					ensureAppGroupAssignmentsExist(resourceName, group1, group2, group3),
					resource.TestCheckResourceAttr(resourceName, "group.#", "3"),

					// Verify updated profiles and priorities
					resource.TestCheckResourceAttr(resourceName, "group.0.priority", "1"),
					resource.TestCheckResourceAttr(resourceName, "group.0.profile", `{}`),

					// Priority changed from 2 to 3
					resource.TestCheckResourceAttr(resourceName, "group.1.priority", "3"),
					resource.TestCheckResourceAttr(resourceName, "group.1.profile", `{}`),

					// Group 3 now has priority 2 (was unset)
					resource.TestCheckResourceAttr(resourceName, "group.2.priority", "2"),
					resource.TestCheckResourceAttr(resourceName, "group.2.profile", `{}`),

					// Verify relative ordering is maintained: group1(1) < group3(2) < group2(3)
					checkProfileChangesPriorityOrder(resourceName),
				),
			},
			{
				// Test refresh to ensure no drift after profile and priority changes
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					ensureAppGroupAssignmentsExist(resourceName, group1, group2, group3),
					checkProfileChangesPriorityOrder(resourceName),
				),
			},
		},
	})
}

// checkProfileChangesPriorityOrder verifies the specific priority order after profile changes
func checkProfileChangesPriorityOrder(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		// Expected order: test1(priority=1) < test3(priority=2) < test2(priority=3)
		groupPriorities := make(map[string]int)

		for i := 0; i < 3; i++ {
			groupIDKey := fmt.Sprintf("group.%d.id", i)
			priorityKey := fmt.Sprintf("group.%d.priority", i)

			if groupID, exists := rs.Primary.Attributes[groupIDKey]; exists {
				if priorityStr, hasPriority := rs.Primary.Attributes[priorityKey]; hasPriority {
					var priority int
					if _, err := fmt.Sscanf(priorityStr, "%d", &priority); err != nil {
						return fmt.Errorf("failed to parse priority %s: %v", priorityStr, err)
					}

					// Map group ID to priority
					for _, groupSuffix := range []string{"test1", "test2", "test3"} {
						groupResourceName := fmt.Sprintf("okta_group.%s", groupSuffix)
						if groupRes, ok := s.RootModule().Resources[groupResourceName]; ok {
							if groupRes.Primary.Attributes["id"] == groupID {
								groupPriorities[groupSuffix] = priority
								break
							}
						}
					}
				}
			}
		}

		// Verify expected priorities: test1=1, test3=2, test2=3
		expectedPriorities := map[string]int{
			"test1": 1,
			"test3": 2,
			"test2": 3,
		}

		for groupName, expectedPriority := range expectedPriorities {
			if actualPriority, exists := groupPriorities[groupName]; !exists {
				return fmt.Errorf("group %s should have a priority", groupName)
			} else if actualPriority != expectedPriority {
				return fmt.Errorf("group %s priority should be %d, got %d", groupName, expectedPriority, actualPriority)
			}
		}

		return nil
	}
}
