package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	env_var_base     = "FOXOPS_"
	endpoint_env_var = env_var_base + "ENDPOINT"
	token_env_var    = env_var_base + "TOKEN"
)

type Version string

type ClientEndpoint string
type ClientToken string
type ClientConstructor func(ClientEndpoint, ClientToken, Version) FoxopsClient

type foxopsProvider struct {
	version     Version
	clientCtor  ClientConstructor
	datasources []func() datasource.DataSource
	resources   []func() resource.Resource
}

type FoxopsProviderModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	Token    types.String `tfsdk:"token"`
}

func New(
	version Version,
	clientCtor ClientConstructor,
	datasources []func() datasource.DataSource,
	resources []func() resource.Resource,
) func() provider.Provider {
	return func() provider.Provider {
		return &foxopsProvider{
			version:     version,
			clientCtor:  clientCtor,
			datasources: datasources,
			resources:   resources,
		}
	}
}

func (p *foxopsProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "foxops"
	resp.Version = string(p.version)
}

func (p *foxopsProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `foxops` provider allows you to configure and manage your Foxops incarnation using infrastructure as code.",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "The base endpoint at which your Foxops instance can be reached.",
				Optional:            true,
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "The token used to authenticate to your Foxops instance.",
				Optional:            true,
			},
		},
	}
}

func (p *foxopsProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(
		ctx,
		"Configuring FoxOps provider version",
		map[string]interface{}{
			"version": p.version,
		},
	)

	var data FoxopsProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.Endpoint.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("endpoint"),
			"Unknown Foxops API endpoint",
			fmt.Sprintf(
				"The provider cannot create the Foxops API client as there is an unknown configuration value for the API endpoint."+
					"Either target apply the source of the value first, set the value statically in the configuration or use the %s environment variable.",
				endpoint_env_var,
			),
		)
	}

	if data.Token.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Unknown Foxops API token",
			fmt.Sprintf(
				"The provider cannot create the Foxops API client as there is an unknown configuration value for the API token."+
					"Either target apply the source of the value first, set the value statically in the configuration or use the %s environment variable.",
				token_env_var,
			),
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := os.Getenv(endpoint_env_var)
	token := os.Getenv(token_env_var)

	if !data.Endpoint.IsNull() {
		endpoint = data.Endpoint.ValueString()
	}

	if !data.Token.IsNull() {
		token = data.Token.ValueString()
	}

	if endpoint == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("endpoint"),
			"Missing Foxops API endpoint",
			fmt.Sprintf(
				"The provider cannot create the Foxops API client as there is a missing configuration value for the API endpoint."+
					"Set the endpoint value in the configuration or use the %s environment variable. IF either is already set, ensure the value is not empty.",
				endpoint_env_var,
			),
		)
	}

	if token == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Missing Foxops API token",
			fmt.Sprintf(
				"The provider cannot create the Foxops API client as there is a missing configuration value for the API token."+
					"Set the token value in the configuration or use the %s environment variable. IF either is already set, ensure the value is not empty.",
				token_env_var,
			),
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	client := p.clientCtor(
		ClientEndpoint(endpoint),
		ClientToken(token),
		p.version,
	)

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *foxopsProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return p.datasources
}

func (p *foxopsProvider) Resources(context.Context) []func() resource.Resource {
	return p.resources
}
