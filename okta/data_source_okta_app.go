package okta

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &AppDataSource{}

func NewAppDataSource() datasource.DataSource { return &AppDataSource{} }

type AppDataSource struct{ config *Config }

type AppDataSourceModel struct {
	ID types.String `tfsdk:"id"`
}

func (d *AppDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app"
}

func (d *AppDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":           schema.StringAttribute{Computed: true},
			"created":      schema.StringAttribute{Computed: true},
			"last_updated": schema.StringAttribute{Computed: true},
			"name":         schema.StringAttribute{Computed: true},
			"label":        schema.StringAttribute{Computed: true},
			"status":       schema.StringAttribute{Computed: true},
			"sign_on_mode": schema.StringAttribute{Computed: true},
			"features": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"admin_note":   schema.StringAttribute{Computed: true},
			"enduser_note": schema.StringAttribute{Computed: true},
			"visibility": schema.ObjectAttribute{
				Computed: true,
				AttributeTypes: map[string]attr.Type{
					"auto_launch":         types.BoolType,
					"auto_submit_toolbar": types.BoolType,
					"hide": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"ios": types.BoolType,
							"web": types.BoolType,
						},
					},
				},
			},
		},
	}
}

func (d *AppDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.config = dataSourceConfiguration(req, resp)
}

func (d *AppDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state AppDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	client := d.config.oktaSDKClientV5
	apiRequest := client.ApplicationAPI.GetApplication(ctx, state.ID.ValueString())
	app, _, err := apiRequest.Execute()
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Okta App", fmt.Sprintf("Error retrieving app: %s", err.Error()))
		return
	}

	oktaApp, ok := app.GetActualInstance().(OktaApp)
	if !ok {
		resp.Diagnostics.AddError("Unable to Read Okta App", "App is not an OktaApp")
		return
	}

	oktaAppModel := buildOktaAppModel(ctx, oktaApp)

	resp.Diagnostics.Append(resp.State.Set(ctx, &oktaAppModel)...)
}
