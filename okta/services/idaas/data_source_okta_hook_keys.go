package idaas

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/okta/terraform-provider-okta/okta/config"
)

var (
	_ datasource.DataSource              = &hookKeysDataSource{}
	_ datasource.DataSourceWithConfigure = &hookKeysDataSource{}
)

type hookKeysDataSource struct {
	*config.Config
}

type hookKeysDataSourceModel struct {
	HookKeys []hookKeysItemModel `tfsdk:"hook_keys"`
}

type hookKeysItemModel struct {
	Id          types.String `tfsdk:"id"`
	KeyId       types.String `tfsdk:"key_id"`
	Name        types.String `tfsdk:"name"`
	Created     types.String `tfsdk:"created"`
	LastUpdated types.String `tfsdk:"last_updated"`
	IsUsed      types.Bool   `tfsdk:"is_used"`
}

func newHookKeysDataSource() datasource.DataSource {
	return &hookKeysDataSource{}
}

func (d *hookKeysDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_hook_keys"
}

func (d *hookKeysDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.Config = dataSourceConfiguration(req, resp)
}

func (d *hookKeysDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Get a list of hook keys from Okta.",
		Attributes: map[string]schema.Attribute{
			"hook_keys": schema.ListAttribute{
				Computed:    true,
				Description: "The list of hook keys.",
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":           types.StringType,
						"key_id":       types.StringType,
						"name":         types.StringType,
						"created":      types.StringType,
						"last_updated": types.StringType,
						"is_used":      types.BoolType,
					},
				},
			},
		},
	}
}

func (d *hookKeysDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state hookKeysDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hookKeys, _, err := d.OktaIDaaSClient.OktaSDKClientV5().HookKeyAPI.ListHookKeys(ctx).Execute()
	if err != nil {
		resp.Diagnostics.AddError("Error listing hook keys", "Could not list hook keys, unexpected error: "+err.Error())
		return
	}

	state.HookKeys = make([]hookKeysItemModel, 0, len(hookKeys))
	for i := range hookKeys {
		hk := &hookKeys[i]
		state.HookKeys = append(state.HookKeys, hookKeysItemModel{
			Id:          types.StringPointerValue(hk.Id),
			KeyId:       types.StringPointerValue(hk.KeyId),
			Name:        types.StringPointerValue(hk.Name),
			Created:     types.StringValue(hk.GetCreated().Format(time.RFC3339)),
			LastUpdated: types.StringValue(hk.GetLastUpdated().Format(time.RFC3339)),
			IsUsed:      types.BoolPointerValue(hk.IsUsed),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
