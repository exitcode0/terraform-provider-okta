package idaas

import (
	"context"
	"fmt"
	"hash/crc32"
	"strings"

	v6okta "github.com/okta/okta-sdk-golang/v6/okta"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetworkZones() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceNetworkZonesRead,
		Schema: map[string]*schema.Schema{
			"filter": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter criteria. Filter value will be URL-encoded by the provider.",
			},
			"network_zones": {
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
						"usage": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"dynamic_proxy_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"asns": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"dynamic_locations": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"dynamic_locations_exclude": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"gateways": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"proxies": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"ip_service_categories_include": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"ip_service_categories_exclude": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},
		Description: "Get a list of network zones from Okta.",
	}
}

func dataSourceNetworkZonesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := getOktaV6ClientFromMetadata(m)
	req := client.NetworkZoneAPI.ListNetworkZones(ctx)
	filter, ok := d.GetOk("filter")
	if ok {
		req = req.Filter(filter.(string))
	}
	zones, resp, err := req.Execute()
	if err != nil {
		return diag.Errorf("failed to list network zones: %v", err)
	}
	for resp.HasNextPage() {
		var more []v6okta.ListNetworkZones200ResponseInner
		resp, err = resp.Next(&more)
		if err != nil {
			return diag.Errorf("failed to list network zones: %v", err)
		}
		zones = append(zones, more...)
	}

	filterVal := ""
	if ok {
		filterVal = filter.(string)
	}
	d.SetId(fmt.Sprintf("%d", crc32.ChecksumIEEE([]byte(filterVal))))

	arr := make([]map[string]interface{}, len(zones))
	for i := range zones {
		arr[i] = flattenNetworkZone(&zones[i])
	}
	_ = d.Set("network_zones", arr)
	return nil
}

func flattenNetworkZone(data *v6okta.ListNetworkZones200ResponseInner) map[string]interface{} {
	result := map[string]interface{}{
		"id":                            "",
		"name":                          "",
		"type":                          "",
		"status":                        "",
		"usage":                         "",
		"dynamic_proxy_type":            "",
		"asns":                          []string{},
		"dynamic_locations":             []string{},
		"dynamic_locations_exclude":     []string{},
		"gateways":                      []string{},
		"proxies":                       []string{},
		"ip_service_categories_include": []string{},
		"ip_service_categories_exclude": []string{},
	}
	if data == nil {
		return result
	}
	nz := data.GetActualInstance()
	if nz == nil {
		return result
	}
	switch v := nz.(type) {
	case *v6okta.DynamicNetworkZone:
		result["id"] = v.GetId()
		result["name"] = v.GetName()
		result["type"] = v.GetType()
		result["status"] = v.GetStatus()
		result["usage"] = v.GetUsage()
		result["dynamic_proxy_type"] = v.GetProxyType()
		if asns := v.GetAsns(); asns != nil {
			result["asns"] = asns
		}
		result["dynamic_locations"] = flattenLocationsToStringSlice(v.GetLocations())
	case *v6okta.EnhancedDynamicNetworkZone:
		result["id"] = v.GetId()
		result["name"] = v.GetName()
		result["type"] = v.GetType()
		result["status"] = v.GetStatus()
		result["usage"] = v.GetUsage()
		if asns, ok := v.GetAsnsOk(); ok {
			result["asns"] = asns.GetInclude()
		}
		if location, ok := v.GetLocationsOk(); ok {
			result["dynamic_locations"] = flattenLocationsToStringSlice(location.GetInclude())
			result["dynamic_locations_exclude"] = flattenLocationsToStringSlice(location.GetExclude())
		}
		if ips, ok := v.GetIpServiceCategoriesOk(); ok {
			result["ip_service_categories_include"] = ips.GetInclude()
			result["ip_service_categories_exclude"] = ips.GetExclude()
		}
	case *v6okta.IPNetworkZone:
		result["id"] = v.GetId()
		result["name"] = v.GetName()
		result["type"] = v.GetType()
		result["status"] = v.GetStatus()
		result["usage"] = v.GetUsage()
		result["gateways"] = flattenAddressesToStringSlice(v.GetGateways())
		result["proxies"] = flattenAddressesToStringSlice(v.GetProxies())
	}
	return result
}

func flattenAddressesToStringSlice(addresses []v6okta.NetworkZoneAddress) []string {
	result := make([]string, len(addresses))
	for i := range addresses {
		result[i] = addresses[i].GetValue()
	}
	return result
}

func flattenLocationsToStringSlice(locations []v6okta.NetworkZoneLocation) []string {
	result := make([]string, len(locations))
	for i := range locations {
		if strings.Contains(locations[i].GetRegion(), "-") {
			result[i] = locations[i].GetRegion()
		} else {
			result[i] = locations[i].GetCountry()
		}
	}
	return result
}
