package idaas

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	oktav5sdk "github.com/okta/okta-sdk-golang/v5/okta"
	"github.com/okta/terraform-provider-okta/okta/config"
)

var (
	_ datasource.DataSource              = &appTokensDataSource{}
	_ datasource.DataSourceWithConfigure = &appTokensDataSource{}
)

func newAppTokensDataSource() datasource.DataSource {
	return &appTokensDataSource{}
}

type appTokensDataSource struct {
	*config.Config
}

type appTokensDataSourceModel struct {
	AppID  types.String `tfsdk:"app_id"`
	Tokens types.List   `tfsdk:"tokens"`
}

var appTokenObjectType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"id":           types.StringType,
		"status":       types.StringType,
		"created":      types.StringType,
		"expires_at":   types.StringType,
		"last_updated": types.StringType,
		"client_id":    types.StringType,
		"user_id":      types.StringType,
		"issuer":       types.StringType,
		"scopes":       types.ListType{ElemType: types.StringType},
	},
}

func (d *appTokensDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app_tokens"
}

func (d *appTokensDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Get a list of OAuth2 tokens for an application from Okta.",
		Attributes: map[string]schema.Attribute{
			"app_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the application.",
			},
			"tokens": schema.ListAttribute{
				Computed:    true,
				Description: "The list of OAuth2 tokens for the application.",
				ElementType: appTokenObjectType,
			},
		},
	}
}

func (d *appTokensDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.Config = dataSourceConfiguration(req, resp)
}

func (d *appTokensDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data appTokensDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	appID := data.AppID.ValueString()

	tokens, apiResp, err := d.OktaIDaaSClient.OktaSDKClientV5().ApplicationTokensAPI.ListOAuth2TokensForApplication(ctx, appID).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error listing application tokens",
			fmt.Sprintf("Could not list tokens for application %s: %s", appID, err.Error()),
		)
		return
	}

	for apiResp.HasNextPage() {
		var nextTokens []oktav5sdk.OAuth2RefreshToken
		apiResp, err = apiResp.Next(&nextTokens)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error listing application tokens",
				fmt.Sprintf("Could not retrieve next page of tokens for application %s: %s", appID, err.Error()),
			)
			return
		}
		tokens = append(tokens, nextTokens...)
	}

	tokenObjects := make([]attr.Value, len(tokens))
	for i, token := range tokens {
		scopeVals := make([]attr.Value, len(token.GetScopes()))
		for j, scope := range token.GetScopes() {
			scopeVals[j] = types.StringValue(scope)
		}
		scopesList, diags := types.ListValue(types.StringType, scopeVals)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		created := ""
		if c, ok := token.GetCreatedOk(); ok && c != nil {
			created = c.Format(time.RFC3339)
		}
		expiresAt := ""
		if e, ok := token.GetExpiresAtOk(); ok && e != nil {
			expiresAt = e.Format(time.RFC3339)
		}
		lastUpdated := ""
		if l, ok := token.GetLastUpdatedOk(); ok && l != nil {
			lastUpdated = l.Format(time.RFC3339)
		}

		obj, diags := types.ObjectValue(appTokenObjectType.AttrTypes, map[string]attr.Value{
			"id":           types.StringValue(token.GetId()),
			"status":       types.StringValue(token.GetStatus()),
			"created":      types.StringValue(created),
			"expires_at":   types.StringValue(expiresAt),
			"last_updated": types.StringValue(lastUpdated),
			"client_id":    types.StringValue(token.GetClientId()),
			"user_id":      types.StringValue(token.GetUserId()),
			"issuer":       types.StringValue(token.GetIssuer()),
			"scopes":       scopesList,
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		tokenObjects[i] = obj
	}

	tokensList, diags := types.ListValue(appTokenObjectType, tokenObjects)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Tokens = tokensList
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
