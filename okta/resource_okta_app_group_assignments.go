package okta

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/okta/terraform-provider-okta/sdk"
)

func resourceAppGroupAssignments() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAppGroupAssignmentsCreate,
		ReadContext:   resourceAppGroupAssignmentsRead,
		DeleteContext: resourceAppGroupAssignmentsDelete,
		UpdateContext: resourceAppGroupAssignmentsUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				_ = d.Set("app_id", d.Id())
				return []*schema.ResourceData{d}, nil
			},
		},
		Description: `Manages ALL group assignments for an application.

**Important**: 
- This resource takes complete ownership of group assignments for the specified application
- Any groups assigned outside of Terraform will be imported into the state
- Group priorities control the relative ordering of assignments
- Okta may re-sequence priority values (e.g., 1,3,5 becomes 1,2,3) while maintaining your specified relative ordering
- Priority 0 is a valid priority value; omitting the priority field means no specific priority
- Do not use in conjunction with for_each`,
		Schema: map[string]*schema.Schema{
			"app_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the application to assign a group to.",
			},
			"group": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "A group to assign to this application",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "A group to associate with the application",
						},
						"priority": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Priority of group assignment. Okta's API may sometimes re-sequence priority values to ensure they are sequential while maintaining relative ordering. Priority 0 is valid; omit this field if no specific priority is needed.",
						},
						"profile": {
							Type:             schema.TypeString,
							ValidateDiagFunc: stringIsJSON,
							StateFunc:        normalizeDataJSON,
							Required:         true,
							DiffSuppressFunc: noChangeInObjectFromUnmarshaledJSON,
							DefaultFunc: func() (interface{}, error) {
								return "{}", nil
							},
							Description: "JSON document containing [application profile](https://developer.okta.com/docs/reference/api/apps/#profile-object)",
						},
					},
				},
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(1 * time.Hour),
			Read:   schema.DefaultTimeout(1 * time.Hour),
			Update: schema.DefaultTimeout(1 * time.Hour),
		},
	}
}

func resourceAppGroupAssignmentsCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := getOktaClientFromMetadata(meta)
	assignments := tfGroupsToGroupAssignments(d)

	// run through all groups in the set and create an assignment
	for i := range assignments {
		_, _, err := client.Application.CreateApplicationGroupAssignment(
			ctx,
			d.Get("app_id").(string),
			assignments[i].Id,
			*assignments[i],
		)
		if err != nil {
			return diag.Errorf("failed to create application group assignment: %v", err)
		}
	}

	// okta_app_group_assignments completely control all assignments for an application
	d.SetId(d.Get("app_id").(string))
	return resourceAppGroupAssignmentsRead(ctx, d, meta)
}

func resourceAppGroupAssignmentsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := getOktaClientFromMetadata(meta)
	// remember, current group assignments is an API call and are all groups
	// assigned to the app, even those initiated outside the provider, for
	// instance those assignments from "click ops"
	currentGroupAssignments, resp, err := listApplicationGroupAssignments(
		ctx,
		client,
		d.Get("app_id").(string),
	)
	if err := suppressErrorOn404(resp, err); err != nil {
		return diag.Errorf("failed to fetch group assignments: %v", err)
	}
	if currentGroupAssignments == nil {
		d.SetId("")
		return nil
	}
	g, ok := d.GetOk("group")
	if ok {
		groupAssignments := syncGroups(d, g.([]interface{}), currentGroupAssignments)
		err := setNonPrimitives(d, map[string]interface{}{"group": groupAssignments})
		if err != nil {
			return diag.Errorf("failed to set group properties: %v", err)
		}
	} else {
		groupAssignments := make([]map[string]interface{}, len(currentGroupAssignments))
		for i := range currentGroupAssignments {
			groupAssignments[i] = groupAssignmentToTFGroup(currentGroupAssignments[i])
		}
		err := setNonPrimitives(d, map[string]interface{}{"group": groupAssignments})
		if err != nil {
			return diag.Errorf("failed to set group properties: %v", err)
		}
	}
	return nil
}

func resourceAppGroupAssignmentsUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := getOktaClientFromMetadata(meta)
	appID := d.Get("app_id").(string)
	toAssign, toRemove, err := splitAssignmentsTargets(d)
	if err != nil {
		return diag.Errorf("failed to discern group assignment splits: %v", err)
	}
	err = deleteGroupAssignments(
		client.Application.DeleteApplicationGroupAssignment,
		ctx,
		appID,
		toRemove,
	)
	if err != nil {
		return diag.Errorf("failed to delete group assignment: %v", err)
	}
	err = addGroupAssignments(
		client.Application.CreateApplicationGroupAssignment,
		ctx,
		appID,
		toAssign,
	)
	if err != nil {
		return diag.Errorf("failed to add/update group assignment: %v", err)
	}
	return resourceAppGroupAssignmentsRead(ctx, d, meta)
}

func resourceAppGroupAssignmentsDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := getOktaClientFromMetadata(meta)
	for _, rawGroup := range d.Get("group").([]interface{}) {
		group := rawGroup.(map[string]interface{})
		resp, err := client.Application.DeleteApplicationGroupAssignment(
			ctx,
			d.Get("app_id").(string),
			group["id"].(string),
		)
		if err := suppressErrorOn404(resp, err); err != nil {
			return diag.Errorf("failed to delete application group assignment: %v", err)
		}
	}
	return nil
}

// syncGroups compares tfGroups - the groups set in the config, with all group
// assignments, and all assignments known to the API. This function now handles
// Okta's priority re-sequencing behavior by preserving relative ordering while
// updating absolute priority values to match the API response.
func syncGroups(d *schema.ResourceData, tfGroups []interface{}, groupAssignments []*sdk.ApplicationGroupAssignment) []interface{} {
	var result []interface{}

	// Build maps for easier lookup
	apiGroupsMap := make(map[string]*sdk.ApplicationGroupAssignment)
	for _, assignment := range groupAssignments {
		apiGroupsMap[assignment.Id] = assignment
	}

	// Build expected relative ordering from terraform config
	// Use d.GetOk to check if priorities were explicitly set (same as tfGroupsToGroupAssignments)
	configPriorities := make(map[string]int)
	configOrder := make([]string, 0)
	hasPriorities := false

	for i, tfGroup := range tfGroups {
		group := tfGroup.(map[string]interface{})
		groupID := group["id"].(string)
		configOrder = append(configOrder, groupID)

		// Use d.GetOk to check if priority was explicitly set (same method as create/update)
		priority, ok := d.GetOk(fmt.Sprintf("group.%d.priority", i))
		if ok {
			configPriorities[groupID] = priority.(int)
			hasPriorities = true
		}
	}

	// First pass: handle groups that were in the terraform config
	for i, tfGroup := range tfGroups {
		group := tfGroup.(map[string]interface{})
		groupID := group["id"].(string)

		// Check if this group still exists in API
		apiAssignment, exists := apiGroupsMap[groupID]
		if !exists {
			// Group no longer exists in API, skip it
			continue
		}

		resultGroup := map[string]interface{}{
			"id":      apiAssignment.Id,
			"profile": buildProfile(d, i, apiAssignment),
		}

		// Handle priority: sync API values for groups that had priority in config
		if apiAssignment.PriorityPtr != nil {
			// Check if this group had priority in the original config
			if _, hadPriorityInConfig := configPriorities[groupID]; hadPriorityInConfig {
				// Always sync the API value - this handles both re-sequencing and rejection cases
				// If Okta rejected the change, the user will see drift and can decide how to handle it
				resultGroup["priority"] = int(*apiAssignment.PriorityPtr)
			}
			// If user didn't set priority for this group, don't sync it to avoid drift
		}

		result = append(result, resultGroup)
	}

	// Second pass: add any groups that exist in API but not in config
	// (these would be added via click-ops or other means)
	for _, apiAssignment := range groupAssignments {
		found := false
		for _, tfGroup := range tfGroups {
			group := tfGroup.(map[string]interface{})
			if group["id"].(string) == apiAssignment.Id {
				found = true
				break
			}
		}

		if found {
			continue
		}

		// This is a new group not in terraform config
		newGroup := map[string]interface{}{
			"id": apiAssignment.Id,
		}

		// Handle profile
		if apiAssignment.Profile != nil {
			if p, ok := apiAssignment.Profile.(string); ok {
				newGroup["profile"] = p
			} else {
				// Convert to JSON string if it's not already
				jsonProfile, err := json.Marshal(apiAssignment.Profile)
				if err == nil {
					newGroup["profile"] = string(jsonProfile)
				} else {
					newGroup["profile"] = "{}"
				}
			}
		} else {
			newGroup["profile"] = "{}"
		}

		// Handle priority
		if apiAssignment.PriorityPtr != nil {
			newGroup["priority"] = int(*apiAssignment.PriorityPtr)
		}

		result = append(result, newGroup)
	}

	// Validate that relative ordering is preserved for configured groups
	// Only do this if the user actually specified priorities in their config
	if hasPriorities && len(configPriorities) > 1 {
		if !validateRelativeOrdering(result, configOrder, configPriorities) {
			// Log a warning but don't fail - this indicates Okta changed the relative ordering
			// which might be intentional (e.g., if there were conflicts)
		}
	}

	return result
}

// validateRelativeOrdering checks if the API response maintains the relative
// ordering specified in the terraform configuration
func validateRelativeOrdering(result []interface{}, configOrder []string, configPriorities map[string]int) bool {
	// Build current priorities from result
	currentPriorities := make(map[string]int)
	for _, r := range result {
		group := r.(map[string]interface{})
		// Handle both int and int64 types from API
		if priority, ok := group["priority"].(int); ok {
			currentPriorities[group["id"].(string)] = priority
		} else if priority, ok := group["priority"].(int64); ok {
			currentPriorities[group["id"].(string)] = int(priority)
		}
	}

	// Check if relative ordering is maintained
	configGroups := make([]string, 0)
	for _, groupID := range configOrder {
		if _, hasConfigPriority := configPriorities[groupID]; hasConfigPriority {
			if _, hasCurrentPriority := currentPriorities[groupID]; hasCurrentPriority {
				configGroups = append(configGroups, groupID)
			}
		}
	}

	// Sort by configured priority
	sort.Slice(configGroups, func(i, j int) bool {
		return configPriorities[configGroups[i]] < configPriorities[configGroups[j]]
	})

	// Sort by current priority
	currentGroups := make([]string, len(configGroups))
	copy(currentGroups, configGroups)
	sort.Slice(currentGroups, func(i, j int) bool {
		return currentPriorities[currentGroups[i]] < currentPriorities[currentGroups[j]]
	})

	// Check if the ordering is the same
	for i := range configGroups {
		if configGroups[i] != currentGroups[i] {
			return false
		}
	}

	return true
}

func buildProfile(d *schema.ResourceData, i int, assignment *sdk.ApplicationGroupAssignment) string {
	if i < 0 || assignment == nil {
		return ""
	}

	oldProfile, ok := d.Get(fmt.Sprintf("group.%d.profile", i)).(string)
	if !ok {
		return ""
	}
	opm := make(map[string]interface{})

	err := json.Unmarshal([]byte(oldProfile), &opm)
	if err != nil {
		return ""
	}
	ap, ok := assignment.Profile.(map[string]interface{})
	if !ok {
		return ""
	}

	// copy new values from assignment profile to the old profile only if old
	// profile has the attribute and the new value is not nil
	for k, v := range ap {
		if v == nil {
			continue
		}
		if _, ok := opm[k]; ok {
			opm[k] = v
		}
	}

	jsonProfile, err := json.Marshal(&opm)
	if err != nil {
		return ""
	}

	return string(jsonProfile)
}

// splitAssignmentsTargets uses schema change to determine what if any
// assignments to keep and which to remove. This is in the context of the local
// terraform state. Get changes returns old state vs new state. Anything in the
// old state but not in the new state will be removed.  Otherwise, everything is
// to be assigned. That way, if there are changes to an existing assignment
// (e.g. on priority or profile) they'll still be posted to the API for update.
func splitAssignmentsTargets(d *schema.ResourceData) (toAssign, toRemove []*sdk.ApplicationGroupAssignment, err error) {
	// 1. Anything in old, but not in new, needs to be deleted
	// 2. Treat everything else as to be added that will also take care of field
	//    updates on priority and profile
	o, n := d.GetChange("group")
	oldState, ok := o.([]interface{})
	if !ok {
		err = fmt.Errorf("expected old groups to be slice, got %T", o)
		return
	}
	newState, ok := n.([]interface{})
	if !ok {
		err = fmt.Errorf("expected new groups to be slice, got %T", n)
		return
	}

	oldIDs := map[string]interface{}{}
	newIDs := map[string]interface{}{}
	for _, old := range oldState {
		if o, ok := old.(map[string]interface{}); ok {
			id := o["id"].(string)
			oldIDs[id] = o
		}
	}
	for _, new := range newState {
		if n, ok := new.(map[string]interface{}); ok {
			id := n["id"].(string)
			newIDs[id] = n
		}
	}

	// delete
	for id := range oldIDs {
		if newIDs[id] == nil {
			// only id is needed
			toRemove = append(toRemove, &sdk.ApplicationGroupAssignment{
				Id: id,
			})
		}
	}

	// anything in the new state treat as an assign even though it might already
	// exist and might be unchanged
	for id, group := range newIDs {
		a := group.(map[string]interface{})
		assignment := &sdk.ApplicationGroupAssignment{
			Id: id,
		}
		if profile, ok := a["profile"]; ok {
			var p interface{}
			if err = json.Unmarshal([]byte(profile.(string)), &p); err == nil {
				assignment.Profile = p
			}
			err = nil // need to reset err as it is a named return value
		}
		if priority, ok := a["priority"]; ok {
			assignment.PriorityPtr = int64Ptr(priority.(int))
		}
		toAssign = append(toAssign, assignment)
	}

	return
}

func groupAssignmentToTFGroup(assignment *sdk.ApplicationGroupAssignment) map[string]interface{} {
	jsonProfile, _ := json.Marshal(assignment.Profile)
	profile := "{}"
	if string(jsonProfile) != "" {
		profile = string(jsonProfile)
	}
	result := map[string]interface{}{
		"id":      assignment.Id,
		"profile": profile,
	}
	if assignment.PriorityPtr != nil {
		result["priority"] = int(*assignment.PriorityPtr)
	}
	return result
}

func tfGroupsToGroupAssignments(d *schema.ResourceData) []*sdk.ApplicationGroupAssignment {
	assignments := make([]*sdk.ApplicationGroupAssignment, len(d.Get("group").([]interface{})))
	for i := range d.Get("group").([]interface{}) {
		rawProfile := d.Get(fmt.Sprintf("group.%d.profile", i))
		var profile interface{}
		_ = json.Unmarshal([]byte(rawProfile.(string)), &profile)
		a := &sdk.ApplicationGroupAssignment{
			Id:      d.Get(fmt.Sprintf("group.%d.id", i)).(string),
			Profile: profile,
		}
		priority, ok := d.GetOk(fmt.Sprintf("group.%d.priority", i))
		if ok {
			a.PriorityPtr = int64Ptr(priority.(int))
		}
		assignments[i] = a
	}
	return assignments
}

// addGroupAssignments adds all group assignments
func addGroupAssignments(
	add func(context.Context, string, string, sdk.ApplicationGroupAssignment) (*sdk.ApplicationGroupAssignment, *sdk.Response, error),
	ctx context.Context,
	appID string,
	assignments []*sdk.ApplicationGroupAssignment,
) error {
	for _, assignment := range assignments {
		_, _, err := add(ctx, appID, assignment.Id, *assignment)
		if err != nil {
			return err
		}
	}
	return nil
}

// deleteGroupAssignments deletes all group assignments
func deleteGroupAssignments(
	delete func(context.Context, string, string) (*sdk.Response, error),
	ctx context.Context,
	appID string,
	assignments []*sdk.ApplicationGroupAssignment,
) error {
	for i := range assignments {
		_, err := delete(ctx, appID, assignments[i].Id)
		if err != nil {
			return fmt.Errorf("could not delete assignment for group %s, to application %s: %w", assignments[i].Id, appID, err)
		}
	}
	return nil
}
