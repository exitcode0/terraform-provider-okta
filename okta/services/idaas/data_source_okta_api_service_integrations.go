package idaas

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v5okta "github.com/okta/okta-sdk-golang/v5/okta"
	"github.com/okta/terraform-provider-okta/okta/config"
)

var (
	_ datasource.DataSource              = &apiServiceIntegrationsDataSource{}
	_ datasource.DataSourceWithConfigure = &apiServiceIntegrationsDataSource{}
)

func newAPIServiceIntegrationsDataSource() datasource.DataSource {
	return &apiServiceIntegrationsDataSource{}
}

type apiServiceIntegrationsDataSourceModel struct {
	APIServiceIntegrations []apiServiceIntegrationsItemModel `tfsdk:"api_service_integrations"`
}

type apiServiceIntegrationsItemModel struct {
	ID             types.String `tfsdk:"id"`
	Type           types.String `tfsdk:"type"`
	Name           types.String `tfsdk:"name"`
	ConfigGuideUrl types.String `tfsdk:"config_guide_url"`
	Created        types.String `tfsdk:"created"`
	CreatedBy      types.String `tfsdk:"created_by"`
	GrantedScopes  types.List   `tfsdk:"granted_scopes"`
}

type apiServiceIntegrationsDataSource struct {
	*config.Config
}

func (d *apiServiceIntegrationsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_service_integrations"
}

func (d *apiServiceIntegrationsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.Config = dataSourceConfiguration(req, resp)
}

func (d *apiServiceIntegrationsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Get a list of API service integrations from Okta.",
		Attributes: map[string]schema.Attribute{
			"api_service_integrations": schema.ListAttribute{
				Description: "The list of API service integrations.",
				Computed:    true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":               types.StringType,
						"type":             types.StringType,
						"name":             types.StringType,
						"config_guide_url": types.StringType,
						"created":          types.StringType,
						"created_by":       types.StringType,
						"granted_scopes":   types.ListType{ElemType: types.StringType},
					},
				},
			},
		},
	}
}

func (d *apiServiceIntegrationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state apiServiceIntegrationsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	integrations, apiResp, err := d.OktaIDaaSClient.OktaSDKClientV5().ApiServiceIntegrationsAPI.ListApiServiceIntegrationInstances(ctx).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to list API service integrations",
			fmt.Sprintf("Error retrieving API service integrations: %s", err.Error()),
		)
		return
	}

	for apiResp.HasNextPage() {
		var nextIntegrations []v5okta.APIServiceIntegrationInstance
		apiResp, err = apiResp.Next(&nextIntegrations)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to paginate API service integrations",
				fmt.Sprintf("Error paginating API service integrations: %s", err.Error()),
			)
			return
		}
		integrations = append(integrations, nextIntegrations...)
	}

	state.APIServiceIntegrations = make([]apiServiceIntegrationsItemModel, 0, len(integrations))
	for _, integration := range integrations {
		scopes := make([]attr.Value, 0, len(integration.GetGrantedScopes()))
		for _, scope := range integration.GetGrantedScopes() {
			scopes = append(scopes, types.StringValue(scope))
		}
		scopesList, diags := types.ListValue(types.StringType, scopes)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		state.APIServiceIntegrations = append(state.APIServiceIntegrations, apiServiceIntegrationsItemModel{
			ID:             types.StringValue(integration.GetId()),
			Type:           types.StringValue(integration.GetType()),
			Name:           types.StringValue(integration.GetName()),
			ConfigGuideUrl: types.StringValue(integration.GetConfigGuideUrl()),
			Created:        types.StringValue(integration.GetCreatedAt()),
			CreatedBy:      types.StringValue(integration.GetCreatedBy()),
			GrantedScopes:  scopesList,
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
