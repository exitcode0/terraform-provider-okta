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
				Description: "Filter criteria. Filter value will be URL-encoded by the provider",
			},
			"network_zones": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of network zones matching the filter criteria",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the network zone",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the network zone",
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Type of the Network Zone - can be `IP`, `DYNAMIC` or `DYNAMIC_V2` only",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Network Status - can either be ACTIVE or INACTIVE only",
						},
						"usage": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Zone's purpose: POLICY or BLOCKLIST",
						},
						"dynamic_proxy_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Type of proxy being controlled by this dynamic network zone - can be one of `Any`, `TorAnonymizer` or `NotTorAnonymizer`. Use with type `DYNAMIC`",
						},
						"asns": {
							Type:        schema.TypeList,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "List of asns included. Format of each array value: a string representation of an ASN numeric value. Use with type `DYNAMIC` or `DYNAMIC_V2`",
						},
						"dynamic_locations": {
							Type:        schema.TypeList,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Array of locations ISO-3166-1(2) included. Format code: countryCode OR countryCode-regionCode. Use with type `DYNAMIC` or `DYNAMIC_V2`",
						},
						"dynamic_locations_exclude": {
							Type:        schema.TypeList,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Array of locations ISO-3166-1(2) excluded. Format code: countryCode OR countryCode-regionCode. Use with type `DYNAMIC_V2`",
						},
						"gateways": {
							Type:        schema.TypeList,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Array of values in CIDR/range form depending on the way it's been declared (i.e. CIDR will contain /suffix). Please check API docs for examples. Use with type `IP`",
						},
						"proxies": {
							Type:        schema.TypeList,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Array of values in CIDR/range form depending on the way it's been declared (i.e. CIDR will contain /suffix). Please check API docs for examples. Can not be set if `usage` is set to `BLOCKLIST`. Use with type `IP`",
						},
						"ip_service_categories_include": {
							Type:        schema.TypeList,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "List of ip service included. Use with type `DYNAMIC_V2`",
						},
						"ip_service_categories_exclude": {
							Type:        schema.TypeList,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "List of ip service excluded. Use with type `DYNAMIC_V2`",
						},
					},
				},
			},
		},
		Description: "Get a list of network zones from Okta.",
	}
}

func dataSourceNetworkZonesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := getOktaV6ClientFromMetadata(meta)
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
		var moreZones []v6okta.ListNetworkZones200ResponseInner
		resp, err = resp.Next(&moreZones)
		if err != nil {
			return diag.Errorf("failed to list network zones: %v", err)
		}
		zones = append(zones, moreZones...)
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
		result["asns"] = v.GetAsns()
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
