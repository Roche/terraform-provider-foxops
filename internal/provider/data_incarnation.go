package provider

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type incarnationDataSource struct {
	client FoxopsClient
}

var _ datasource.DataSourceWithConfigure = (*incarnationDataSource)(nil)

func NewIncarnationDataSource() datasource.DataSource {
	return &incarnationDataSource{}
}

type waitForStatusMRModel struct {
	Status  types.String `tfsdk:"status"`
	Timeout types.String `tfsdk:"timeout"`
}

var waitForSchema = schema.SingleNestedAttribute{
	MarkdownDescription: "Wait for the status of the last merge request to reach a status before completing the current operation. " +
		"This field only affects incarnation that have been updated as it requires a merge request to exist.",
	Optional: true,
	Attributes: map[string]schema.Attribute{
		"status": schema.StringAttribute{
			MarkdownDescription: "The expected status for the merge request. Can be one of `open`, `merge`, `closed` or `unknown`.",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.OneOf(
					"open",
					"merged",
					"closed",
					"unknown",
				),
			},
		},
		"timeout": schema.StringAttribute{
			MarkdownDescription: "The amount of time to wait for the expected status to be reached. " +
				"It should be a sequence of numbers followed by a unit suffix (`s`, `m` or `h`). " +
				"Example: `1m30s`. Default: `10s`.",
			Optional: true,
			Validators: []validator.String{
				stringvalidator.RegexMatches(
					regexp.MustCompile(`(\d+[smh])+`),
					`must be a sequence of numbers with a unit suffix. Valid unit suffixes are "s", "m" and "h". Example: "1m30s"`,
				),
			},
		},
	},
}

type incarnationDatasourceModel struct {
	Id                        types.String          `tfsdk:"id"`
	IncarnationRepository     types.String          `tfsdk:"incarnation_repository"`
	TargetDirectory           types.String          `tfsdk:"target_directory"`
	TemplateData              types.Map             `tfsdk:"template_data"`
	TemplateRepository        types.String          `tfsdk:"template_repository"`
	TemplateRepositoryVersion types.String          `tfsdk:"template_repository_version"`
	MergeRequestUrl           types.String          `tfsdk:"merge_request_url"`
	CommitSha                 types.String          `tfsdk:"commit_sha"`
	CommitUrl                 types.String          `tfsdk:"commit_url"`
	MergeRequestStatus        types.String          `tfsdk:"merge_request_status"`
	MergeRequestId            types.String          `tfsdk:"merge_request_id"`
	WaitForMRStatus           *waitForStatusMRModel `tfsdk:"wait_for_mr_status"`
}

func (ds *incarnationDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_incarnation"
}

func (ds *incarnationDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(FoxopsClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected provider.FoxopsClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	ds.client = client
}

func (ds *incarnationDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Use this data source to get information about an incarnation.",
		MarkdownDescription: "Use this data source to get information about an incarnation.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The `id` of the incarnation.",
				Required:            true,
			},
			"incarnation_repository": schema.StringAttribute{
				MarkdownDescription: "The repository in which the incarnation will be created.",
				Computed:            true,
			},
			"target_directory": schema.StringAttribute{
				MarkdownDescription: "The folder in which the incarnation will be created. Default: `.`.",
				Computed:            true,
			},
			"template_data": schema.MapAttribute{
				MarkdownDescription: "An object containing variables used to generate the incarnation. " +
					"These variables should match those declared in the `fengine.yaml` file of the template",
				ElementType: types.StringType,
				Computed:    true,
			},
			"template_repository": schema.StringAttribute{
				MarkdownDescription: "The repository containing the template used to create the incarnation.",
				Computed:            true,
			},
			"template_repository_version": schema.StringAttribute{
				MarkdownDescription: "A tag, commit or branch of the template repository to use for the incarnation.",
				Computed:            true,
			},
			"merge_request_url": schema.StringAttribute{
				MarkdownDescription: "The url of the latest merge request created for the incarnation. " +
					"This property will be `null` after the creation of the incarnation and only populated after updates.",
				Computed: true,
			},
			"commit_sha": schema.StringAttribute{
				MarkdownDescription: "The hash of the last commit created for the incarnation.",
				Computed:            true,
			},
			"commit_url": schema.StringAttribute{
				MarkdownDescription: "The url of the last commit created for the incarnation.",
				Computed:            true,
			},
			"merge_request_status": schema.StringAttribute{
				MarkdownDescription: "The status of the last merge request created for the incarnation. " +
					"This property will be `null` after the creation of the incarnation and only populated after updates. " +
					"It will be one of `open`, `merged`, `closed` or `unknown`.",
				Computed: true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						"open",
						"merged",
						"closed",
						"unknown",
					),
				},
			},
			"merge_request_id": schema.StringAttribute{
				MarkdownDescription: "The id of the last merge request created for the incarnation. " +
					"This property will be `null` after the creation of the incarnation and only populated after updates.",
				Computed: true,
			},
			"wait_for_mr_status": waitForSchema,
		},
	}
}

func (ds *incarnationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data incarnationDatasourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := data.Id.ValueString()

	inc := getIncarnation(
		ctx,
		ds.client,
		resp.Diagnostics,
		IncarnationId(id),
		data.WaitForMRStatus,
	)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Id = types.StringValue(string(inc.Id))
	data.IncarnationRepository = types.StringValue(inc.IncarnationRepository)
	data.TargetDirectory = types.StringValue(inc.TargetDirectory)
	data.CommitSha = types.StringValue(inc.CommitSha)
	data.CommitUrl = types.StringValue(inc.CommitUrl)
	data.TemplateRepository = types.StringValue(inc.TemplateRepository)
	data.TemplateRepositoryVersion = types.StringValue(inc.TemplateRepositoryVersion)

	if inc.MergeRequestId != nil {
		data.MergeRequestId = types.StringValue(*inc.MergeRequestId)
	}

	if inc.MergeRequestUrl != nil {
		data.MergeRequestUrl = types.StringValue(*inc.MergeRequestUrl)
	}

	if inc.MergeRequestStatus != nil {
		data.MergeRequestStatus = types.StringValue(*inc.MergeRequestStatus)
	}

	var diags diag.Diagnostics
	templateData := map[string]string{}
	for key, value := range inc.TemplateData {
		switch value := value.(type) {
		case string:
			templateData[key] = value
		case int:
			templateData[key] = fmt.Sprintf("%d", value)
		case float64:
			templateData[key] = fmt.Sprintf("%f", value)
		}
	}
	data.TemplateData, diags = types.MapValueFrom(ctx, types.StringType, templateData)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
