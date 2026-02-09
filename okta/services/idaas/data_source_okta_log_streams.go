package idaas

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/okta/okta-sdk-golang/v4/okta"
	"github.com/okta/terraform-provider-okta/okta/config"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &logStreamsDataSource{}
	_ datasource.DataSourceWithConfigure = &logStreamsDataSource{}
)

func newLogStreamsDataSource() datasource.DataSource {
	return &logStreamsDataSource{}
}

type logStreamsDataSource struct {
	*config.Config
}

type logStreamsDataSourceModel struct {
	Filter     types.String          `tfsdk:"filter"`
	LogStreams []logStreamsItemModel `tfsdk:"log_streams"`
}

type logStreamsItemModel struct {
	ID     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Type   types.String `tfsdk:"type"`
	Status types.String `tfsdk:"status"`
}

func (d *logStreamsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_log_streams"
}

func (d *logStreamsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Get a list of log streams from Okta.",
		Attributes: map[string]schema.Attribute{
			"filter": schema.StringAttribute{
				Optional:    true,
				Description: "Filter expression to limit results (e.g. `type eq \"aws_eventbridge\"` or `status eq \"ACTIVE\"`).",
			},
			"log_streams": schema.ListAttribute{
				Description: "The list of log streams.",
				Computed:    true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":     types.StringType,
						"name":   types.StringType,
						"type":   types.StringType,
						"status": types.StringType,
					},
				},
			},
		},
	}
}

func (d *logStreamsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.Config = dataSourceConfiguration(req, resp)
}

func (d *logStreamsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state logStreamsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client := d.OktaIDaaSClient.OktaSDKClientV3()
	apiReq := client.LogStreamAPI.ListLogStreams(ctx)
	if filter := state.Filter.ValueString(); filter != "" {
		apiReq = apiReq.Filter(filter)
	}

	logStreamList, apiResp, err := apiReq.Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Okta Log Streams",
			fmt.Sprintf("Error retrieving log streams: %s", err.Error()),
		)
		return
	}

	// Handle pagination
	for apiResp.HasNextPage() {
		var nextLogStreams []okta.ListLogStreams200ResponseInner
		apiResp, err = apiResp.Next(&nextLogStreams)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Read Okta Log Streams",
				fmt.Sprintf("Error retrieving next page of log streams: %s", err.Error()),
			)
			return
		}
		logStreamList = append(logStreamList, nextLogStreams...)
	}

	for _, ls := range logStreamList {
		item := logStreamsItemModel{}
		if ls.LogStreamAws != nil {
			item.ID = types.StringValue(ls.LogStreamAws.Id)
			item.Name = types.StringValue(ls.LogStreamAws.Name)
			item.Type = types.StringValue(string(ls.LogStreamAws.Type))
			item.Status = types.StringValue(ls.LogStreamAws.Status)
		} else if ls.LogStreamSplunk != nil {
			item.ID = types.StringValue(ls.LogStreamSplunk.Id)
			item.Name = types.StringValue(ls.LogStreamSplunk.Name)
			item.Type = types.StringValue(string(ls.LogStreamSplunk.Type))
			item.Status = types.StringValue(ls.LogStreamSplunk.Status)
		} else {
			continue
		}
		state.LogStreams = append(state.LogStreams, item)
	}

	// Ensure log_streams is an empty list rather than null when no results
	if state.LogStreams == nil {
		state.LogStreams = []logStreamsItemModel{}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
