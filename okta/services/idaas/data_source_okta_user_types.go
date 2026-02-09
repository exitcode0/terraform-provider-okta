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

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &userTypesDataSource{}
	_ datasource.DataSourceWithConfigure = &userTypesDataSource{}
)

type userTypesDataSourceModel struct {
	UserTypes []userTypeModel `tfsdk:"user_types"`
}

func newUserTypesDataSource() datasource.DataSource {
	return &userTypesDataSource{}
}

type userTypesDataSource struct {
	*config.Config
}

func (d *userTypesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_types"
}

func (d *userTypesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Get a list of user types from Okta.",
		Attributes: map[string]schema.Attribute{
			"user_types": schema.ListAttribute{
				Description: "The list of user types.",
				Computed:    true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":           types.StringType,
						"name":         types.StringType,
						"display_name": types.StringType,
						"description":  types.StringType,
					},
				},
			},
		},
	}
}

func (d *userTypesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.Config = dataSourceConfiguration(req, resp)
}

func (d *userTypesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state userTypesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	userTypeListResp, _, err := d.OktaIDaaSClient.OktaSDKClientV2().UserType.ListUserTypes(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Okta User Types",
			fmt.Sprintf("Error retrieving user types: %s", err.Error()),
		)
		return
	}

	for _, ut := range userTypeListResp {
		state.UserTypes = append(state.UserTypes, userTypeModel{
			ID:          types.StringValue(ut.Id),
			Name:        types.StringValue(ut.Name),
			DisplayName: types.StringValue(ut.DisplayName),
			Description: types.StringValue(ut.Description),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
