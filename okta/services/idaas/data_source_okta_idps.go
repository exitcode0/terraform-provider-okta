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
				Description: "Filters by identity provider type (OIDC, SAML2, APPLE, FACEBOOK, GOOGLE, LINKEDIN, MICROSOFT).",
			},
			"idps": {
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
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"protocol_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"issuer_mode": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
		Description: "Get a list of identity providers from Okta.",
	}
}

func dataSourceIdpsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	qp := &query.Params{Limit: utils.DefaultPaginationLimit}
	if q, ok := d.GetOk("q"); ok {
		qp.Q = q.(string)
	}
	if idpType, ok := d.GetOk("type"); ok {
		qp.Type = idpType.(string)
	}

	idps, err := collectIdps(ctx, getOktaClientFromMetadata(m), qp)
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

func collectIdps(ctx context.Context, client *sdk.Client, qp *query.Params) ([]*sdk.IdentityProvider, error) {
	idps, resp, err := client.IdentityProvider.ListIdentityProviders(ctx, qp)
	if err != nil {
		return nil, err
	}
	for resp.HasNextPage() {
		var more []*sdk.IdentityProvider
		resp, err = resp.Next(ctx, &more)
		if err != nil {
			return nil, err
		}
		idps = append(idps, more...)
	}
	return idps, nil
}
