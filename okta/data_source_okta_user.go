package okta

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/okta/okta-sdk-golang/v2/okta"
	"github.com/okta/okta-sdk-golang/v2/okta/query"
)

var userSearchSchemaDescription = "Filter to find " +
	"user/users. Each filter will be concatenated with " +
	"the compound search operator. Please be aware profile properties " +
	"must match what is in Okta, which is likely camel case. " +
	"Expression is a free form expression filter " +
	"https://developer.okta.com/docs/reference/core-okta-api/#filter . " +
	"The set name/value/comparison properties will be ignored if expression is present"

var userSearchSchema = map[string]*schema.Schema{
	"name": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Property name to search for. This requires the search feature be on. Please see Okta documentation on their filter API for users. https://developer.okta.com/docs/api/resources/users#list-users-with-search",
	},
	"value": {
		Type:     schema.TypeString,
		Optional: true,
	},
	"comparison": {
		Type:     schema.TypeString,
		Optional: true,
		Default:  "eq",
	},
	"expression": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "A raw search expression string. This requires the search feature be on. Please see Okta documentation on their filter API for users. https://developer.okta.com/docs/api/resources/users#list-users-with-search",
	},
}

func dataSourceUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceUserRead,
		Schema: buildUserDataSourceSchema(map[string]*schema.Schema{
			"user_id": {
				Type:        schema.TypeString,
				Description: "Retrieve a single user based on their id",
				Optional:    true,
			},
			"search": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: userSearchSchemaDescription,
				Elem: &schema.Resource{
					Schema: userSearchSchema,
				},
			},
			"skip_groups": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Do not populate user groups information (prevents additional API call)",
			},
			"skip_roles": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Do not populate user roles information (prevents additional API call)",
			},
			"compound_search_operator": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          "and",
				ValidateDiagFunc: elemInSlice([]string{"and", "or"}),
				Description:      "Search operator used when joining mulitple search clauses",
			},
		}),
	}
}

func dataSourceUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := getOktaClientFromMetadata(m)
	var user *okta.User
	var err error
	userID, ok := d.GetOk("user_id")
	_, searchCriteriaOk := d.GetOk("search")
	if ok {
		logger(m).Info("reading user by ID", "id", userID.(string))
		user, _, err = client.User.GetUser(ctx, userID.(string))
		if err != nil {
			return diag.Errorf("failed to get user: %v", err)
		}
	} else if searchCriteriaOk {
		var users []*okta.User
		sc := getSearchCriteria(d)
		logger(m).Info("reading user using search", "search", sc)
		users, _, err = client.User.ListUsers(ctx, &query.Params{Search: sc, Limit: 1})
		if err != nil {
			return diag.Errorf("failed to list users: %v", err)
		} else if len(users) < 1 {
			return diag.Errorf("no users found using search criteria: %+v", sc)
		}
		user = users[0]
	}
	d.SetId(user.Id)
	rawMap := flattenUser(user)
	err = setNonPrimitives(d, rawMap)
	if err != nil {
		return diag.Errorf("failed to set user's properties: %v", err)
	}

	if d.Get("skip_roles") == false {
		err = setAdminRoles(ctx, d, m)
		if err != nil {
			return diag.Errorf("failed to set user's admin roles: %v", err)
		}
	}

	if d.Get("skip_groups") == false {
		err = setAllGroups(ctx, d, client)
		if err != nil {
			return diag.Errorf("failed to set user's groups: %v", err)
		}
	}

	return nil
}

func getSearchCriteria(d *schema.ResourceData) string {
	rawFilters := d.Get("search").(*schema.Set)
	filterList := make([]string, rawFilters.Len())
	for i, f := range rawFilters.List() {
		fmap := f.(map[string]interface{})
		rawExpression := fmap["expression"]
		if rawExpression != "" && rawExpression != nil {
			filterList[i] = fmt.Sprintf(`%s`, fmap["expression"])
			continue
		}

		// Need to set up the filter clause to allow comparions that do not
		// accept a right hand argument and those that do.
		// profile.email pr
		filterList[i] = fmt.Sprintf(`%s %s`, fmap["name"], fmap["comparison"])
		if fmap["value"] != "" {
			// profile.email eq "example@example.com"
			filterList[i] = fmt.Sprintf(`%s "%s"`, filterList[i], fmap["value"])
		}
	}

	operator := " and "
	cso := d.Get("compound_search_operator")
	if cso != nil && cso.(string) != "" {
		operator = fmt.Sprintf(" %s ", cso.(string))
	}
	return strings.Join(filterList, operator)
}
