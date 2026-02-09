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
	"github.com/okta/terraform-provider-okta/okta/utils"
)

var _ datasource.DataSource = &devicesDataSource{}

func newDevicesDataSource() datasource.DataSource {
	return &devicesDataSource{}
}

type devicesDataSource struct {
	*config.Config
}

type devicesDataSourceModel struct {
	Search  types.String       `tfsdk:"search"`
	Devices []devicesItemModel `tfsdk:"devices"`
}

type devicesItemModel struct {
	Id                  types.String `tfsdk:"id"`
	Status              types.String `tfsdk:"status"`
	ResourceType        types.String `tfsdk:"resource_type"`
	Created             types.String `tfsdk:"created"`
	LastUpdated         types.String `tfsdk:"last_updated"`
	ResourceAlternateId types.String `tfsdk:"resource_alternate_id"`
	ResourceId          types.String `tfsdk:"resource_id"`
	Profile             types.Object `tfsdk:"profile"`
	ResourceDisplayName types.Object `tfsdk:"resource_display_name"`
}

var deviceProfileAttrTypes = map[string]attr.Type{
	"display_name":            types.StringType,
	"platform":                types.StringType,
	"registered":              types.BoolType,
	"disk_encryption_type":    types.StringType,
	"imei":                    types.StringType,
	"integrity_jail_break":    types.BoolType,
	"manufacturer":            types.StringType,
	"meid":                    types.StringType,
	"model":                   types.StringType,
	"os_version":              types.StringType,
	"secure_hardware_present": types.BoolType,
	"serial_number":           types.StringType,
	"sid":                     types.StringType,
	"tpm_public_key_hash":     types.StringType,
	"udid":                    types.StringType,
}

var deviceResourceDisplayNameAttrTypes = map[string]attr.Type{
	"sensitive": types.BoolType,
	"value":     types.StringType,
}

func (d *devicesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_devices"
}

func (d *devicesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.Config = dataSourceConfiguration(req, resp)
}

func (d *devicesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Get a list of devices from Okta.",
		Attributes: map[string]schema.Attribute{
			"search": schema.StringAttribute{
				Optional:    true,
				Description: "A SCIM filter expression that filters the results. Searches include all Device profile properties and the Device id, status, and lastUpdated properties.",
			},
			"devices": schema.ListAttribute{
				Description: "The list of devices.",
				Computed:    true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":                    types.StringType,
						"status":                types.StringType,
						"resource_type":         types.StringType,
						"created":               types.StringType,
						"last_updated":          types.StringType,
						"resource_alternate_id": types.StringType,
						"resource_id":           types.StringType,
						"profile": types.ObjectType{
							AttrTypes: deviceProfileAttrTypes,
						},
						"resource_display_name": types.ObjectType{
							AttrTypes: deviceResourceDisplayNameAttrTypes,
						},
					},
				},
			},
		},
	}
}

func (d *devicesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state devicesDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiRequest := d.OktaIDaaSClient.OktaSDKClientV5().DeviceAPI.ListDevices(ctx).Limit(int32(utils.DefaultPaginationLimit))
	if search := state.Search.ValueString(); search != "" {
		apiRequest = apiRequest.Search(search)
	}

	deviceList, apiResp, err := apiRequest.Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading devices",
			fmt.Sprintf("Could not list devices: %s", err.Error()),
		)
		return
	}

	for apiResp.HasNextPage() {
		var nextDevices []oktav5sdk.DeviceList
		apiResp, err = apiResp.Next(&nextDevices)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading devices",
				fmt.Sprintf("Error retrieving next page of devices: %s", err.Error()),
			)
			return
		}
		deviceList = append(deviceList, nextDevices...)
	}

	state.Devices = make([]devicesItemModel, 0, len(deviceList))
	for _, device := range deviceList {
		profile := device.GetProfile()
		profileValue, diags := types.ObjectValue(deviceProfileAttrTypes, map[string]attr.Value{
			"display_name":            types.StringValue(profile.GetDisplayName()),
			"platform":                types.StringValue(profile.GetPlatform()),
			"registered":              types.BoolValue(profile.GetRegistered()),
			"disk_encryption_type":    types.StringValue(profile.GetDiskEncryptionType()),
			"imei":                    types.StringValue(profile.GetImei()),
			"integrity_jail_break":    types.BoolValue(profile.GetIntegrityJailbreak()),
			"manufacturer":            types.StringValue(profile.GetManufacturer()),
			"meid":                    types.StringValue(profile.GetMeid()),
			"model":                   types.StringValue(profile.GetModel()),
			"os_version":              types.StringValue(profile.GetOsVersion()),
			"secure_hardware_present": types.BoolValue(profile.GetSecureHardwarePresent()),
			"serial_number":           types.StringValue(profile.GetSerialNumber()),
			"sid":                     types.StringValue(profile.GetSid()),
			"tpm_public_key_hash":     types.StringValue(profile.GetTpmPublicKeyHash()),
			"udid":                    types.StringValue(profile.GetUdid()),
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		resourceDisplayName := device.GetResourceDisplayName()
		resourceDisplayNameValue, diags := types.ObjectValue(deviceResourceDisplayNameAttrTypes, map[string]attr.Value{
			"sensitive": types.BoolValue(resourceDisplayName.GetSensitive()),
			"value":     types.StringValue(resourceDisplayName.GetValue()),
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		state.Devices = append(state.Devices, devicesItemModel{
			Id:                  types.StringValue(device.GetId()),
			Status:              types.StringValue(device.GetStatus()),
			ResourceType:        types.StringValue(device.GetResourceType()),
			Created:             types.StringValue(device.GetCreated().Format(time.RFC3339)),
			LastUpdated:         types.StringValue(device.GetLastUpdated().Format(time.RFC3339)),
			ResourceAlternateId: types.StringValue(device.GetResourceAlternateId()),
			ResourceId:          types.StringValue(device.GetResourceId()),
			Profile:             profileValue,
			ResourceDisplayName: resourceDisplayNameValue,
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
