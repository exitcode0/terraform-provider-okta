package idaas

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/okta/terraform-provider-okta/okta/config"
)

var (
	_ datasource.DataSource              = &pushProvidersDataSource{}
	_ datasource.DataSourceWithConfigure = &pushProvidersDataSource{}
)

func newPushProvidersDataSource() datasource.DataSource {
	return &pushProvidersDataSource{}
}

type pushProvidersDataSource struct {
	*config.Config
}

type pushProvidersDataSourceModel struct {
	Type          types.String            `tfsdk:"type"`
	PushProviders []pushProviderItemModel `tfsdk:"push_providers"`
}

type pushProviderItemModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	ProviderType    types.String `tfsdk:"provider_type"`
	LastUpdatedDate types.String `tfsdk:"last_updated_date"`
}

func (d *pushProvidersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_push_providers"
}

func (d *pushProvidersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Get a list of push providers from Okta.",
		Attributes: map[string]schema.Attribute{
			"type": schema.StringAttribute{
				Description: "Filters push providers by `providerType`. Valid values: `APNS`, `FCM`.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("APNS", "FCM"),
				},
			},
			"push_providers": schema.ListAttribute{
				Description: "The list of push providers.",
				Computed:    true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":                types.StringType,
						"name":              types.StringType,
						"provider_type":     types.StringType,
						"last_updated_date": types.StringType,
					},
				},
			},
		},
	}
}

func (d *pushProvidersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.Config = dataSourceConfiguration(req, resp)
}

func (d *pushProvidersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state pushProvidersDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	listReq := d.OktaIDaaSClient.OktaSDKClientV5().PushProviderAPI.ListPushProviders(ctx)
	if !state.Type.IsNull() && !state.Type.IsUnknown() {
		listReq = listReq.Type_(state.Type.ValueString())
	}

	providers, _, err := listReq.Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to list Okta Push Providers",
			fmt.Sprintf("Error retrieving push providers: %s", err.Error()),
		)
		return
	}

	state.PushProviders = make([]pushProviderItemModel, 0, len(providers))

	for _, provider := range providers {
		item := pushProviderItemModel{}
		if provider.APNSPushProvider != nil {
			item.ID = types.StringValue(provider.APNSPushProvider.GetId())
			item.Name = types.StringValue(provider.APNSPushProvider.GetName())
			item.ProviderType = types.StringValue(provider.APNSPushProvider.GetProviderType())
			item.LastUpdatedDate = types.StringValue(provider.APNSPushProvider.GetLastUpdatedDate())
		} else if provider.FCMPushProvider != nil {
			item.ID = types.StringValue(provider.FCMPushProvider.GetId())
			item.Name = types.StringValue(provider.FCMPushProvider.GetName())
			item.ProviderType = types.StringValue(provider.FCMPushProvider.GetProviderType())
			item.LastUpdatedDate = types.StringValue(provider.FCMPushProvider.GetLastUpdatedDate())
		} else {
			continue
		}
		state.PushProviders = append(state.PushProviders, item)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
