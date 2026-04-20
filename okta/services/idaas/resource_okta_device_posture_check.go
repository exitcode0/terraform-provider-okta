package idaas

import (
	"context"
	"net/http"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v6okta "github.com/okta/okta-sdk-golang/v6/okta"
	"github.com/okta/terraform-provider-okta/okta/config"
)

var regexpAlphanumeric = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

var (
	_ resource.Resource                = &devicePostureCheckResource{}
	_ resource.ResourceWithConfigure   = &devicePostureCheckResource{}
	_ resource.ResourceWithImportState = &devicePostureCheckResource{}
)

func newDevicePostureCheckResource() resource.Resource {
	return &devicePostureCheckResource{}
}

type devicePostureCheckResource struct {
	*config.Config
}

type devicePostureCheckResourceModel struct {
	ID                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	Description         types.String `tfsdk:"description"`
	Type                types.String `tfsdk:"type"`
	Platform            types.String `tfsdk:"platform"`
	Query               types.String `tfsdk:"query"`
	VariableName        types.String `tfsdk:"variable_name"`
	MappingType         types.String `tfsdk:"mapping_type"`
	RemediationSettings types.Object `tfsdk:"remediation_settings"`
	CreatedBy           types.String `tfsdk:"created_by"`
	CreatedDate         types.String `tfsdk:"created_date"`
	LastUpdated         types.String `tfsdk:"last_updated"`
	LastUpdatedBy       types.String `tfsdk:"last_updated_by"`
}

var remediationSettingsLinkAttrTypes = map[string]attr.Type{
	"default_url": types.StringType,
	"custom_url":  types.StringType,
}

var remediationSettingsMessageAttrTypes = map[string]attr.Type{
	"default_i18n_key": types.StringType,
	"custom_text":      types.StringType,
}

var remediationSettingsAttrTypes = map[string]attr.Type{
	"link":    types.ObjectType{AttrTypes: remediationSettingsLinkAttrTypes},
	"message": types.ObjectType{AttrTypes: remediationSettingsMessageAttrTypes},
}

func (r *devicePostureCheckResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_posture_check"
}

func (r *devicePostureCheckResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a device posture check. Device posture checks use OSQuery to evaluate device compliance and can be referenced in device assurance policies.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the device posture check.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Display name of the device posture check.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the device posture check.",
				Optional:    true,
			},
			"type": schema.StringAttribute{
				Description: "The type of the device posture check. User-created checks are always `CUSTOM`. Okta-provided defaults are `BUILTIN`.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"platform": schema.StringAttribute{
				Description: "Platform (OS) the check applies to. Valid values: `MACOS`, `WINDOWS`.",
				Required:    true,
			},
			"query": schema.StringAttribute{
				Description: "OSQuery SQL string for the device posture check.",
				Required:    true,
			},
			"variable_name": schema.StringAttribute{
				Description: "Unique variable name used to reference this check in device assurance policies. Must contain only alphanumeric characters.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexpAlphanumeric,
						"must contain only alphanumeric characters",
					),
				},
			},
			"mapping_type": schema.StringAttribute{
				Description: "How the check is rendered in device assurance policies.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"remediation_settings": schema.ObjectAttribute{
				Description: "Remediation instructions shown to the end user when the device posture check fails.",
				Optional:    true,
				Computed:    true,
				AttributeTypes: remediationSettingsAttrTypes,
			},
			"created_by": schema.StringAttribute{
				Description: "User who created the device posture check.",
				Computed:    true,
			},
			"created_date": schema.StringAttribute{
				Description: "Timestamp when the device posture check was created.",
				Computed:    true,
			},
			"last_updated": schema.StringAttribute{
				Description: "Timestamp when the device posture check was last updated.",
				Computed:    true,
			},
			"last_updated_by": schema.StringAttribute{
				Description: "User who last updated the device posture check.",
				Computed:    true,
			},
		},
	}
}

func (r *devicePostureCheckResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.Config = resourceConfiguration(req, resp)
}

func (r *devicePostureCheckResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var state devicePostureCheckResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	reqBody := buildDevicePostureCheckRequest(ctx, state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	created, _, err := r.OktaIDaaSClient.OktaSDKClientV6().DevicePostureCheckAPI.CreateDevicePostureCheck(ctx).DevicePostureCheck(*reqBody).Execute()
	if err != nil {
		resp.Diagnostics.AddError("failed to create device posture check", err.Error())
		return
	}

	resp.Diagnostics.Append(mapDevicePostureCheckToState(ctx, created, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *devicePostureCheckResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state devicePostureCheckResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	check, apiResp, err := r.OktaIDaaSClient.OktaSDKClientV6().DevicePostureCheckAPI.GetDevicePostureCheck(ctx, state.ID.ValueString()).Execute()
	if err != nil {
		if apiResp != nil && apiResp.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("failed to read device posture check", err.Error())
		return
	}

	resp.Diagnostics.Append(mapDevicePostureCheckToState(ctx, check, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *devicePostureCheckResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state devicePostureCheckResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read current state to get the ID
	var currentState devicePostureCheckResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &currentState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	reqBody := buildDevicePostureCheckRequest(ctx, state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, _, err := r.OktaIDaaSClient.OktaSDKClientV6().DevicePostureCheckAPI.ReplaceDevicePostureCheck(ctx, currentState.ID.ValueString()).DevicePostureCheck(*reqBody).Execute()
	if err != nil {
		resp.Diagnostics.AddError("failed to update device posture check", err.Error())
		return
	}

	resp.Diagnostics.Append(mapDevicePostureCheckToState(ctx, updated, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *devicePostureCheckResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state devicePostureCheckResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiResp, err := r.OktaIDaaSClient.OktaSDKClientV6().DevicePostureCheckAPI.DeleteDevicePostureCheck(ctx, state.ID.ValueString()).Execute()
	if err != nil {
		if apiResp != nil && apiResp.StatusCode == http.StatusNotFound {
			return
		}
		resp.Diagnostics.AddError("failed to delete device posture check", err.Error())
		return
	}
}

func (r *devicePostureCheckResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func buildDevicePostureCheckRequest(ctx context.Context, state devicePostureCheckResourceModel, diags *diag.Diagnostics) *v6okta.DevicePostureCheck {
	reqBody := v6okta.NewDevicePostureCheck()
	reqBody.SetName(state.Name.ValueString())
	reqBody.SetPlatform(state.Platform.ValueString())
	reqBody.SetQuery(state.Query.ValueString())
	reqBody.SetType("CUSTOM")

	if !state.Description.IsNull() && !state.Description.IsUnknown() {
		reqBody.SetDescription(state.Description.ValueString())
	}

	if !state.VariableName.IsNull() && !state.VariableName.IsUnknown() {
		reqBody.SetVariableName(state.VariableName.ValueString())
	}

	if !state.MappingType.IsNull() && !state.MappingType.IsUnknown() {
		reqBody.SetMappingType(state.MappingType.ValueString())
	}

	if !state.RemediationSettings.IsNull() && !state.RemediationSettings.IsUnknown() {
		remSettings := v6okta.NewDevicePostureChecksRemediationSettings()

		remAttrs := state.RemediationSettings.Attributes()

		linkObj, ok := remAttrs["link"].(types.Object)
		if ok && !linkObj.IsNull() && !linkObj.IsUnknown() {
			link := v6okta.NewDevicePostureChecksRemediationSettingsLink()
			linkAttrs := linkObj.Attributes()
			if v, ok := linkAttrs["default_url"].(types.String); ok && !v.IsNull() && !v.IsUnknown() {
				link.SetDefaultUrl(v.ValueString())
			}
			if v, ok := linkAttrs["custom_url"].(types.String); ok && !v.IsNull() && !v.IsUnknown() {
				link.SetCustomUrl(v.ValueString())
			}
			remSettings.SetLink(*link)
		}

		msgObj, ok := remAttrs["message"].(types.Object)
		if ok && !msgObj.IsNull() && !msgObj.IsUnknown() {
			msg := v6okta.NewDevicePostureChecksRemediationSettingsMessage()
			msgAttrs := msgObj.Attributes()
			if v, ok := msgAttrs["default_i18n_key"].(types.String); ok && !v.IsNull() && !v.IsUnknown() {
				msg.SetDefaultI18nKey(v.ValueString())
			}
			if v, ok := msgAttrs["custom_text"].(types.String); ok && !v.IsNull() && !v.IsUnknown() {
				msg.SetCustomText(v.ValueString())
			}
			remSettings.SetMessage(*msg)
		}

		reqBody.SetRemediationSettings(*remSettings)
	}

	return reqBody
}

func mapDevicePostureCheckToState(ctx context.Context, check *v6okta.DevicePostureCheck, state *devicePostureCheckResourceModel) (diags diag.Diagnostics) {
	state.ID = types.StringPointerValue(check.Id)
	state.Name = types.StringPointerValue(check.Name)
	state.Type = types.StringPointerValue(check.Type)
	state.Platform = types.StringPointerValue(check.Platform)
	state.Query = types.StringPointerValue(check.Query)
	state.CreatedBy = types.StringPointerValue(check.CreatedBy)
	state.CreatedDate = types.StringPointerValue(check.CreatedDate)
	state.LastUpdatedBy = types.StringPointerValue(check.LastUpdatedBy)

	// The API returns "lastUpdated" but the SDK model maps "lastUpdate" (mismatched json tag).
	// Read from AdditionalProperties to get the correct value.
	if v, ok := check.AdditionalProperties["lastUpdated"].(string); ok {
		state.LastUpdated = types.StringValue(v)
	} else {
		state.LastUpdated = types.StringNull()
	}

	if check.Description != nil {
		state.Description = types.StringPointerValue(check.Description)
	} else {
		state.Description = types.StringNull()
	}

	if check.VariableName != nil {
		state.VariableName = types.StringPointerValue(check.VariableName)
	} else {
		state.VariableName = types.StringNull()
	}

	if check.MappingType != nil {
		state.MappingType = types.StringPointerValue(check.MappingType)
	} else {
		state.MappingType = types.StringNull()
	}

	if check.RemediationSettings != nil {
		remSettings := check.RemediationSettings

		var linkVal attr.Value
		if remSettings.Link != nil {
			linkAttrs := map[string]attr.Value{
				"default_url": types.StringPointerValue(remSettings.Link.DefaultUrl),
				"custom_url":  types.StringPointerValue(remSettings.Link.CustomUrl),
			}
			linkObj, d := types.ObjectValue(remediationSettingsLinkAttrTypes, linkAttrs)
			diags.Append(d...)
			linkVal = linkObj
		} else {
			linkVal = types.ObjectNull(remediationSettingsLinkAttrTypes)
		}

		var msgVal attr.Value
		if remSettings.Message != nil {
			msgAttrs := map[string]attr.Value{
				"default_i18n_key": types.StringPointerValue(remSettings.Message.DefaultI18nKey),
				"custom_text":      types.StringPointerValue(remSettings.Message.CustomText),
			}
			msgObj, d := types.ObjectValue(remediationSettingsMessageAttrTypes, msgAttrs)
			diags.Append(d...)
			msgVal = msgObj
		} else {
			msgVal = types.ObjectNull(remediationSettingsMessageAttrTypes)
		}

		remAttrs := map[string]attr.Value{
			"link":    linkVal,
			"message": msgVal,
		}
		remObj, d := types.ObjectValue(remediationSettingsAttrTypes, remAttrs)
		diags.Append(d...)
		state.RemediationSettings = remObj
	} else {
		state.RemediationSettings = types.ObjectNull(remediationSettingsAttrTypes)
	}

	return diags
}
