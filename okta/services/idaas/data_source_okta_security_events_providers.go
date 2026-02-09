package idaas

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/okta/terraform-provider-okta/okta/config"
)

var (
	_ datasource.DataSource              = &securityEventsProvidersDataSource{}
	_ datasource.DataSourceWithConfigure = &securityEventsProvidersDataSource{}
)

func newSecurityEventsProvidersDataSource() datasource.DataSource {
	return &securityEventsProvidersDataSource{}
}

type securityEventsProvidersDataSource struct {
	*config.Config
}

type securityEventsProvidersDataSourceModel struct {
	SecurityEventsProviders []securityEventsProviderItemModel `tfsdk:"security_events_providers"`
}

type securityEventsProviderItemModel struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Type     types.String `tfsdk:"type"`
	Status   types.String `tfsdk:"status"`
	Settings types.Object `tfsdk:"settings"`
}

func (d *securityEventsProvidersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_events_providers"
}

func (d *securityEventsProvidersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Get a list of security events providers from Okta.",
		Attributes: map[string]schema.Attribute{
			"security_events_providers": schema.ListAttribute{
				Description: "The list of security events providers.",
				Computed:    true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":     types.StringType,
						"name":   types.StringType,
						"type":   types.StringType,
						"status": types.StringType,
						"settings": types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"well_known_url": types.StringType,
								"issuer":         types.StringType,
								"jwks_url":       types.StringType,
							},
						},
					},
				},
			},
		},
	}
}

func (d *securityEventsProvidersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.Config = dataSourceConfiguration(req, resp)
}

func (d *securityEventsProvidersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state securityEventsProvidersDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	providers, _, err := d.OktaIDaaSClient.OktaSDKClientV5().SSFReceiverAPI.ListSecurityEventsProviderInstances(ctx).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to list Security Events Providers",
			fmt.Sprintf("Error retrieving security events providers: %s", err.Error()),
		)
		return
	}

	state.SecurityEventsProviders = make([]securityEventsProviderItemModel, 0, len(providers))

	settingsAttrTypes := map[string]attr.Type{
		"well_known_url": types.StringType,
		"issuer":         types.StringType,
		"jwks_url":       types.StringType,
	}

	for _, provider := range providers {
		settingsValues := map[string]attr.Value{
			"well_known_url": types.StringPointerValue(nil),
			"issuer":         types.StringPointerValue(nil),
			"jwks_url":       types.StringPointerValue(nil),
		}

		settings := provider.GetSettings()
		if settings.HasWellKnownUrl() {
			wellKnownUrl := settings.GetWellKnownUrl()
			settingsValues["well_known_url"] = types.StringValue(wellKnownUrl)
		}
		if settings.HasIssuer() {
			issuer := settings.GetIssuer()
			settingsValues["issuer"] = types.StringValue(issuer)
		}
		if settings.HasJwksUrl() {
			jwksUrl := settings.GetJwksUrl()
			settingsValues["jwks_url"] = types.StringValue(jwksUrl)
		}

		settingsObj, diags := types.ObjectValue(settingsAttrTypes, settingsValues)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		state.SecurityEventsProviders = append(state.SecurityEventsProviders, securityEventsProviderItemModel{
			ID:       types.StringValue(provider.GetId()),
			Name:     types.StringValue(provider.GetName()),
			Type:     types.StringValue(provider.GetType()),
			Status:   types.StringValue(provider.GetStatus()),
			Settings: settingsObj,
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
