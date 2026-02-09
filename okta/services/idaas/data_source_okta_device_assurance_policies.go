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
	_ datasource.DataSource              = &deviceAssurancePoliciesDataSource{}
	_ datasource.DataSourceWithConfigure = &deviceAssurancePoliciesDataSource{}
)

func newDeviceAssurancePoliciesDataSource() datasource.DataSource {
	return &deviceAssurancePoliciesDataSource{}
}

type deviceAssurancePoliciesDataSource struct {
	*config.Config
}

type deviceAssurancePoliciesModel struct {
	Policies []deviceAssurancePolicySummaryModel `tfsdk:"policies"`
}

type deviceAssurancePolicySummaryModel struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Platform types.String `tfsdk:"platform"`
}

func (d *deviceAssurancePoliciesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_assurance_policies"
}

func (d *deviceAssurancePoliciesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Get a list of device assurance policies from Okta.",
		Attributes: map[string]schema.Attribute{
			"policies": schema.ListAttribute{
				Description: "The list of device assurance policies.",
				Computed:    true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":       types.StringType,
						"name":     types.StringType,
						"platform": types.StringType,
					},
				},
			},
		},
	}
}

func (d *deviceAssurancePoliciesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.Config = dataSourceConfiguration(req, resp)
}

func (d *deviceAssurancePoliciesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state deviceAssurancePoliciesModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policies, _, err := d.OktaIDaaSClient.OktaSDKClientV5().DeviceAssuranceAPI.ListDeviceAssurancePolicies(ctx).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Device Assurance Policies",
			fmt.Sprintf("Error retrieving device assurance policies: %s", err.Error()),
		)
		return
	}

	state.Policies = make([]deviceAssurancePolicySummaryModel, 0, len(policies))
	for _, policy := range policies {
		data := policy.GetActualInstance()
		if data == nil {
			continue
		}
		sharedData, ok := data.(OktaDeviceAssurancePolicy)
		if !ok {
			continue
		}
		state.Policies = append(state.Policies, deviceAssurancePolicySummaryModel{
			ID:       types.StringValue(sharedData.GetId()),
			Name:     types.StringValue(sharedData.GetName()),
			Platform: types.StringValue(sharedData.GetPlatform()),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
