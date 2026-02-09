package idaas

import (
	"context"
	"fmt"
	"hash/crc32"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/okta/terraform-provider-okta/sdk"
	"github.com/okta/terraform-provider-okta/sdk/query"
)

func dataSourcePolicies() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcePoliciesRead,
		Schema: map[string]*schema.Schema{
			"type": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Policy type: OKTA_SIGN_ON, PASSWORD, MFA_ENROLL, IDP_DISCOVERY, OAUTH_AUTHORIZATION_POLICY, ACCESS_POLICY, PROFILE_ENROLLMENT, etc.",
			},
			"q": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Search by policy name (substring match).",
			},
			"policies": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Policy ID.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Policy name.",
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Policy type.",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Policy status.",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Policy description.",
						},
						"priority": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Policy priority.",
						},
					},
				},
			},
		},
		Description: "Get a list of policies from Okta.",
	}
}

func dataSourcePoliciesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	policyType := d.Get("type").(string)
	client := getOktaClientFromMetadata(meta)

	qp := &query.Params{Type: policyType}

	q := ""
	if v, ok := d.GetOk("q"); ok {
		q = v.(string)
	}

	policies, resp, err := client.Policy.ListPolicies(ctx, qp)
	if err != nil {
		return diag.Errorf("failed to list policies: %v", err)
	}

	var allPolicies []sdk.Policies
	allPolicies = append(allPolicies, policies...)
	for resp.HasNextPage() {
		var nextPolicies []sdk.Policies
		resp, err = resp.Next(ctx, &nextPolicies)
		if err != nil {
			return diag.Errorf("failed to list policies: %v", err)
		}
		allPolicies = append(allPolicies, nextPolicies...)
	}

	result := make([]map[string]interface{}, 0)
	for _, p := range allPolicies {
		policy := p.(*sdk.Policy)
		if q != "" && !strings.Contains(policy.Name, q) {
			continue
		}
		priority := int64(0)
		if policy.PriorityPtr != nil {
			priority = *policy.PriorityPtr
		}
		result = append(result, map[string]interface{}{
			"id":          policy.Id,
			"name":        policy.Name,
			"type":        policy.Type,
			"status":      policy.Status,
			"description": policy.Description,
			"priority":    priority,
		})
	}

	idSrc := fmt.Sprintf("type=%s&q=%s", policyType, q)
	d.SetId(fmt.Sprintf("%d", crc32.ChecksumIEEE([]byte(idSrc))))
	if err := d.Set("policies", result); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
