package idaas

import (
	"context"
	"fmt"
	"hash/crc32"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/okta/terraform-provider-okta/okta/utils"
	"github.com/okta/terraform-provider-okta/sdk"
)

func dataSourceAuthServerPolicies() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAuthServerPoliciesRead,
		Schema: map[string]*schema.Schema{
			"auth_server_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Auth server ID.",
			},
			"policies": {
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
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"priority": {
							Type:     schema.TypeInt,
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
						"assigned_clients": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},
		Description: "Get a list of policies for an authorization server from Okta.",
	}
}

func dataSourceAuthServerPoliciesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	authServerID := d.Get("auth_server_id").(string)
	policies, _, err := getOktaClientFromMetadata(m).AuthorizationServer.ListAuthorizationServerPolicies(ctx, authServerID)
	if err != nil {
		return diag.Errorf("failed to list authorization server policies: %v", err)
	}
	var s string
	arr := make([]map[string]interface{}, len(policies))
	for i := range policies {
		s += policies[i].Name
		arr[i] = flattenAuthServerPolicy(policies[i])
	}
	_ = d.Set("policies", arr)
	d.SetId(fmt.Sprintf("%s.%d", authServerID, crc32.ChecksumIEEE([]byte(s))))
	return nil
}

func flattenAuthServerPolicy(p *sdk.AuthorizationServerPolicy) map[string]interface{} {
	m := map[string]interface{}{
		"id":          p.Id,
		"name":        p.Name,
		"description": p.Description,
		"status":      p.Status,
		"type":        p.Type,
	}
	if p.PriorityPtr != nil {
		m["priority"] = *p.PriorityPtr
	}
	if p.Conditions != nil && p.Conditions.Clients != nil {
		m["assigned_clients"] = utils.ConvertStringSliceToSet(p.Conditions.Clients.Include)
	}
	return m
}
