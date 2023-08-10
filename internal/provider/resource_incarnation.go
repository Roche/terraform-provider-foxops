package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type incarnationResource struct {
	client FoxopsClient
}

var _ resource.ResourceWithConfigure = (*incarnationResource)(nil)

func NewIncarnationResource() resource.Resource {
	return &incarnationResource{}
}

type incarnationStateSetter interface {
	Set(ctx context.Context, val interface{}) diag.Diagnostics
}

type incarnationResourceModel struct {
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
	WaitForMRStatus           *waitForStatusMRModel `tfsdk:"wait_for_mr_status_on_update"`
	AutoMerge                 types.Bool            `tfsdk:"auto_merge_on_update"`
}

func (ds *incarnationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_incarnation"
}

func (ds *incarnationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *incarnationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Use this resource to create and manage incarnations.",
		MarkdownDescription: "Use this resource to create and manage incarnations.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The `id` of the incarnation.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"incarnation_repository": schema.StringAttribute{
				MarkdownDescription: "The repository in which the incarnation will be created.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"target_directory": schema.StringAttribute{
				MarkdownDescription: "The folder in which the incarnation will be created. Default: `.`.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"template_data": schema.MapAttribute{
				MarkdownDescription: "An object containing variables used to generate the incarnation. " +
					"These variables should match those declared in the `fengine.yaml` file of the template",
				ElementType: types.StringType,
				Optional:    true,
			},
			"template_repository": schema.StringAttribute{
				MarkdownDescription: "The repository containing the template used to create the incarnation.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"template_repository_version": schema.StringAttribute{
				MarkdownDescription: "A tag, commit or branch of the template repository to use for the incarnation.",
				Required:            true,
			},
			"auto_merge_on_update": schema.BoolAttribute{
				MarkdownDescription: "Whether merge request should automatically merged after update of the incarnation.",
				Optional:            true,
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
			"wait_for_mr_status_on_update": waitForSchema,
		},
	}
}

func (r *incarnationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data incarnationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.Id.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	id := data.Id.ValueString()

	inc := getIncarnation(
		ctx,
		r.client,
		resp.Diagnostics,
		IncarnationId(id),
		data.WaitForMRStatus,
	)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.setState(ctx, &resp.State, inc)...)
}

func (r *incarnationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data incarnationResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createIncarnationRequest := CreateIncarnationRequest{
		IncarnationRepository: data.IncarnationRepository.ValueString(),
		TargetDirectory:       data.TargetDirectory.ValueStringPointer(),
		TemplateRepository:    data.TemplateRepository.ValueString(),
		UpdateIncarnationRequest: UpdateIncarnationRequest{
			TemplateData:              map[string]interface{}{},
			TemplateRepositoryVersion: data.TemplateRepositoryVersion.ValueString(),
		},
	}

	if !data.TemplateData.IsNull() {
		for key, value := range data.TemplateData.Elements() {
			if value, ok := value.(types.String); ok {
				createIncarnationRequest.TemplateData[key] = value.ValueString()
			}
		}
	}

	inc, err := r.client.CreateIncarnation(ctx, createIncarnationRequest)
	if err != nil {
		resp.Diagnostics.AddError("failed to create incarnation", err.Error())
		return
	}

	resp.Diagnostics.Append(r.setState(ctx, &resp.State, inc)...)
}

func (r *incarnationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data incarnationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateIncarnationRequest := UpdateIncarnationRequest{
		AutoMerge:                 true,
		TemplateData:              map[string]interface{}{},
		TemplateRepositoryVersion: data.TemplateRepositoryVersion.ValueString(),
	}

	if !data.AutoMerge.IsNull() {
		updateIncarnationRequest.AutoMerge = data.AutoMerge.ValueBool()
	}

	if !data.TemplateData.IsNull() {
		for key, value := range data.TemplateData.Elements() {
			if value, ok := value.(types.String); ok {
				updateIncarnationRequest.TemplateData[key] = value.ValueString()
			}
		}
	}

	inc, err := r.client.UpdateIncarnation(ctx, IncarnationId(data.Id.ValueString()), updateIncarnationRequest)
	if err != nil {
		resp.Diagnostics.AddError("failed to update incarnation", err.Error())
		return
	}

	resp.Diagnostics.Append(r.setState(ctx, &resp.State, inc)...)
	if resp.Diagnostics.HasError() {
		return
	}

	inc = getIncarnation(ctx, r.client, resp.Diagnostics, inc.Id, data.WaitForMRStatus)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.setState(ctx, &resp.State, inc)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *incarnationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data incarnationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteIncarnation(ctx, IncarnationId(data.Id.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("failed to delete incarnation", err.Error())
		return
	}
}

func (r *incarnationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *incarnationResource) setState(ctx context.Context, setter incarnationStateSetter, inc Incarnation) (diags diag.Diagnostics) {
	var data incarnationResourceModel

	data.Id = types.StringValue(string(inc.Id))
	data.IncarnationRepository = types.StringValue(inc.IncarnationRepository)
	data.TemplateRepositoryVersion = types.StringValue(inc.TemplateRepositoryVersion)
	data.TemplateRepository = types.StringValue(inc.TemplateRepository)
	data.TargetDirectory = types.StringValue(inc.TargetDirectory)
	data.CommitSha = types.StringValue(inc.CommitSha)
	data.CommitUrl = types.StringValue(inc.CommitUrl)

	if inc.MergeRequestId != nil {
		data.MergeRequestId = types.StringValue(*inc.MergeRequestId)
	}

	if inc.MergeRequestUrl != nil {
		data.MergeRequestUrl = types.StringValue(*inc.MergeRequestUrl)
	}

	if inc.MergeRequestStatus != nil {
		data.MergeRequestStatus = types.StringValue(*inc.MergeRequestStatus)
	}

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

	diags.Append(diags...)
	if diags.HasError() {
		return
	}

	diags.Append(setter.Set(ctx, data)...)

	return
}
