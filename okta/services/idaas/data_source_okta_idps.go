package idaas

import (
	"context"
	"fmt"
	"hash/crc32"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/okta/terraform-provider-okta/okta/utils"
	"github.com/okta/terraform-provider-okta/sdk/query"
)

func dataSourceIdps() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceIdpsRead,
		Schema: map[string]*schema.Schema{
			"q": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Searches the name property of identity providers for matching value.",
			},
			"type": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filters identity providers by type. Supported values are OIDC, SAML2, APPLE, FACEBOOK, GOOGLE, LINKEDIN, and MICROSOFT.",
			},
			"idps": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Identity provider ID.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Identity provider name.",
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Identity provider type.",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Identity provider status.",
						},
						"protocol_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The type of protocol used by the identity provider.",
						},
						"issuer_mode": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Indicates whether Okta uses the original Okta org domain URL, a custom domain URL, or dynamic.",
						},
					},
				},
			},
		},
		Description: "Get a list of identity providers from Okta.",
	}
}

func dataSourceIdpsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	qp := &query.Params{Limit: utils.DefaultPaginationLimit}
	if q, ok := d.GetOk("q"); ok {
		qp.Q = q.(string)
	}
	if idpType, ok := d.GetOk("type"); ok {
		qp.Type = idpType.(string)
	}

	idps, _, err := getOktaClientFromMetadata(meta).IdentityProvider.ListIdentityProviders(ctx, qp)
	if err != nil {
		return diag.Errorf("failed to list identity providers: %v", err)
	}

	d.SetId(fmt.Sprintf("%d", crc32.ChecksumIEEE([]byte(qp.String()))))

	arr := make([]map[string]interface{}, len(idps))
	for i, idp := range idps {
		arr[i] = map[string]interface{}{
			"id":     idp.Id,
			"name":   idp.Name,
			"type":   idp.Type,
			"status": idp.Status,
		}
		if idp.Protocol != nil {
			arr[i]["protocol_type"] = idp.Protocol.Type
		}
		if idp.IssuerMode != "" {
			arr[i]["issuer_mode"] = idp.IssuerMode
		}
	}

	_ = d.Set("idps", arr)
	return nil
}
