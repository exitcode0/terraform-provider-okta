package idaas

import (
	"context"
	"fmt"
	"hash/crc32"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/okta/terraform-provider-okta/okta/utils"
	"github.com/okta/terraform-provider-okta/sdk"
	"github.com/okta/terraform-provider-okta/sdk/query"
)

func dataSourceGroupRules() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceGroupRulesRead,
		Schema: map[string]*schema.Schema{
			"search": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Searches rules by name for matching value.",
			},
			"group_rules": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"group_assignments": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"expression_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"expression_value": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"users_excluded": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},
		Description: "Get a list of group rules from Okta.",
	}
}

func dataSourceGroupRulesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	qp := &query.Params{Limit: utils.DefaultPaginationLimit}
	search, ok := d.GetOk("search")
	if ok {
		qp.Search = search.(string)
	}
	groupRules, err := collectGroupRules(ctx, getOktaClientFromMetadata(m), qp)
	if err != nil {
		return diag.Errorf("failed to list group rules: %v", err)
	}
	d.SetId(fmt.Sprintf("%d", crc32.ChecksumIEEE([]byte(qp.String()))))
	arr := make([]map[string]interface{}, len(groupRules))
	for i, rule := range groupRules {
		m := map[string]interface{}{
			"id":     rule.Id,
			"name":   rule.Name,
			"status": rule.Status,
			"type":   rule.Type,
		}
		if rule.Conditions != nil && rule.Conditions.Expression != nil {
			m["expression_type"] = rule.Conditions.Expression.Type
			m["expression_value"] = rule.Conditions.Expression.Value
		}
		if rule.Conditions != nil && rule.Conditions.People != nil && rule.Conditions.People.Users != nil {
			m["users_excluded"] = rule.Conditions.People.Users.Exclude
		}
		if rule.Actions != nil && rule.Actions.AssignUserToGroups != nil {
			m["group_assignments"] = rule.Actions.AssignUserToGroups.GroupIds
		}
		arr[i] = m
	}
	_ = d.Set("group_rules", arr)
	return nil
}

func collectGroupRules(ctx context.Context, client *sdk.Client, qp *query.Params) ([]*sdk.GroupRule, error) {
	groupRules, resp, err := client.Group.ListGroupRules(ctx, qp)
	if err != nil {
		return nil, err
	}
	for resp.HasNextPage() {
		var nextGroupRules []*sdk.GroupRule
		resp, err = resp.Next(ctx, &nextGroupRules)
		if err != nil {
			return nil, err
		}
		groupRules = append(groupRules, nextGroupRules...)
	}
	return groupRules, nil
}
