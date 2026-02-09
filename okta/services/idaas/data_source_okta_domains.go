package idaas

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDomains() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDomainsRead,
		Schema: map[string]*schema.Schema{
			"domains": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"domain": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"certificate_source_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"validation_status": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
		Description: "Get a list of custom domains from Okta.",
	}
}

func dataSourceDomainsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	domainList, _, err := getOktaV5ClientFromMetadata(m).CustomDomainAPI.ListCustomDomains(ctx).Execute()
	if err != nil {
		return diag.Errorf("failed to list domains: %v", err)
	}

	d.SetId("domains")
	arr := make([]map[string]interface{}, len(domainList.Domains))
	for i, domain := range domainList.Domains {
		arr[i] = map[string]interface{}{
			"id":                      domain.GetId(),
			"domain":                  domain.GetDomain(),
			"certificate_source_type": domain.GetCertificateSourceType(),
			"validation_status":       domain.GetValidationStatus(),
		}
	}
	_ = d.Set("domains", arr)
	return nil
}
