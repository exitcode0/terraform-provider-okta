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

func dataSourceAuthServers() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAuthServersRead,
		Schema: map[string]*schema.Schema{
			"q": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Searches the name and audiences of authorization servers for matching value.",
			},
			"auth_servers": {
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
						"audiences": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"kid": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"credentials_last_rotated": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"credentials_next_rotation": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"credentials_rotation_mode": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"issuer": {
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
		Description: "Get a list of authorization servers from Okta.",
	}
}

func dataSourceAuthServersRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	qp := &query.Params{Limit: utils.DefaultPaginationLimit}
	if q, ok := d.GetOk("q"); ok {
		qp.Q = q.(string)
	}

	servers, err := collectAuthServers(ctx, getOktaClientFromMetadata(m), qp)
	if err != nil {
		return diag.Errorf("failed to list authorization servers: %v", err)
	}

	d.SetId(fmt.Sprintf("%d", crc32.ChecksumIEEE([]byte(qp.String()))))

	arr := make([]map[string]interface{}, len(servers))
	for i := range servers {
		arr[i] = map[string]interface{}{
			"id":          servers[i].Id,
			"name":        servers[i].Name,
			"description": servers[i].Description,
			"audiences":   utils.ConvertStringSliceToInterfaceSlice(servers[i].Audiences),
			"status":      servers[i].Status,
			"issuer":      servers[i].Issuer,
			"issuer_mode": servers[i].IssuerMode,
		}
		if servers[i].Credentials != nil && servers[i].Credentials.Signing != nil {
			arr[i]["kid"] = servers[i].Credentials.Signing.Kid
			arr[i]["credentials_rotation_mode"] = servers[i].Credentials.Signing.RotationMode
			if servers[i].Credentials.Signing.LastRotated != nil {
				arr[i]["credentials_last_rotated"] = servers[i].Credentials.Signing.LastRotated.String()
			}
			if servers[i].Credentials.Signing.NextRotation != nil {
				arr[i]["credentials_next_rotation"] = servers[i].Credentials.Signing.NextRotation.String()
			}
		}
	}

	_ = d.Set("auth_servers", arr)
	return nil
}

func collectAuthServers(ctx context.Context, client *sdk.Client, qp *query.Params) ([]*sdk.AuthorizationServer, error) {
	servers, resp, err := client.AuthorizationServer.ListAuthorizationServers(ctx, qp)
	if err != nil {
		return nil, err
	}
	for resp.HasNextPage() {
		var more []*sdk.AuthorizationServer
		resp, err = resp.Next(ctx, &more)
		if err != nil {
			return nil, err
		}
		servers = append(servers, more...)
	}
	return servers, nil
}
