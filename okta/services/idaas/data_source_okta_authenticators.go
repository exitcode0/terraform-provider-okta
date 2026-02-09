package idaas

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/crc32"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/okta/terraform-provider-okta/okta/resources"
)

func dataSourceAuthenticators() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAuthenticatorsRead,
		Schema: map[string]*schema.Schema{
			"authenticators": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the authenticator.",
						},
						"key": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "A human-readable string that identifies the authenticator.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the authenticator.",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Status of the authenticator.",
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Type of the authenticator.",
						},
						"settings": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Authenticator settings in JSON format.",
						},
						"provider_json": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Authenticator provider in JSON format.",
						},
						"provider_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Provider type.",
						},
					},
				},
			},
		},
		Description: "Get a list of authenticators from Okta.",
	}
}

func dataSourceAuthenticatorsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if providerIsClassicOrg(ctx, meta) {
		return datasourceOIEOnlyFeatureError(resources.OktaIDaaSAuthenticators)
	}

	authenticators, _, err := getOktaClientFromMetadata(meta).Authenticator.ListAuthenticators(ctx)
	if err != nil {
		return diag.Errorf("failed to list authenticators: %v", err)
	}

	d.SetId(fmt.Sprintf("%d", crc32.ChecksumIEEE([]byte("authenticators"))))

	arr := make([]map[string]interface{}, len(authenticators))
	for i, authenticator := range authenticators {
		attrs := map[string]interface{}{
			"id":     authenticator.Id,
			"key":    authenticator.Key,
			"name":   authenticator.Name,
			"status": authenticator.Status,
			"type":   authenticator.Type,
		}
		if authenticator.Settings != nil {
			b, _ := json.Marshal(authenticator.Settings)
			attrs["settings"] = string(b)
		}
		if authenticator.Provider != nil {
			b, _ := json.Marshal(authenticator.Provider)
			attrs["provider_json"] = string(b)
			attrs["provider_type"] = authenticator.Provider.Type
		}
		arr[i] = attrs
	}

	_ = d.Set("authenticators", arr)
	return nil
}
