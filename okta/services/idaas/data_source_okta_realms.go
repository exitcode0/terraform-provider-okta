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

type realmsDataSource struct {
	config *config.Config
}

type realmsDataSourceModel struct {
	Search types.String      `tfsdk:"search"`
	Realms []realmsItemModel `tfsdk:"realms"`
}

type realmsItemModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	RealmType types.String `tfsdk:"realm_type"`
	IsDefault types.Bool   `tfsdk:"is_default"`
}

func newRealmsDataSource() datasource.DataSource {
	return &realmsDataSource{}
}

func (d *realmsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_realms"
}

func (d *realmsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.config = dataSourceConfiguration(req, resp)
}

func (d *realmsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Get a list of realms from Okta.",
		Attributes: map[string]schema.Attribute{
			"search": schema.StringAttribute{
				Optional:    true,
				Description: "Filter realms by profile.name using the contains operator.",
			},
			"realms": schema.ListAttribute{
				Description: "The list of realms.",
				Computed:    true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":         types.StringType,
						"name":       types.StringType,
						"realm_type": types.StringType,
						"is_default": types.BoolType,
					},
				},
			},
		},
	}
}

func (d *realmsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state realmsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	listReq := d.config.OktaIDaaSClient.OktaSDKClientV5().RealmAPI.ListRealms(ctx)
	if !state.Search.IsNull() && state.Search.ValueString() != "" {
		searchString := fmt.Sprintf(`profile.name co "%s"`, state.Search.ValueString())
		listReq = listReq.Search(searchString)
	}

	realms, _, err := listReq.Execute()
	if err != nil {
		resp.Diagnostics.AddError("Unable to list Okta Realms", fmt.Sprintf("Error retrieving realms: %s", err.Error()))
		return
	}

	state.Realms = make([]realmsItemModel, 0, len(realms))
	for _, realm := range realms {
		state.Realms = append(state.Realms, realmsItemModel{
			ID:        types.StringPointerValue(realm.Id),
			Name:      types.StringValue(realm.Profile.Name),
			RealmType: types.StringPointerValue(realm.Profile.RealmType),
			IsDefault: types.BoolPointerValue(realm.IsDefault),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
