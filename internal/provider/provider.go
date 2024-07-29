package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/terraform-provider-wundergraph/internal/resources"
	"github.com/labd/terraform-provider-wundergraph/internal/utils"
	"github.com/labd/terraform-provider-wundergraph/sdk/wg/cosmo/platform/v1/platformv1connect"
	"net/http"
	"os"
)

// Ensure WundergraphProvider satisfies various provider interfaces.
var _ provider.Provider = &WundergraphProvider{}

// WundergraphProvider defines the provider implementation.
type WundergraphProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// WundergraphProviderModel describes the provider data model.
type WundergraphProviderModel struct {
	ApiKey types.String `tfsdk:"api_key"`
	ApiUrl types.String `tfsdk:"api_url"`
}

func (p *WundergraphProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "wundergraph"
	resp.Version = p.version
}

func (p *WundergraphProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "The API key for the provider.",
				Optional:            true,
			},
			"api_url": schema.StringAttribute{
				MarkdownDescription: "The API URL for the provider.",
				Optional:            true,
			},
		},
	}
}

func (p *WundergraphProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data WundergraphProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	//TODO: simplify input validation

	var apiKey string
	if data.ApiKey.IsUnknown() || data.ApiKey.IsNull() {
		apiKey = os.Getenv("WGC_API_KEY")
	} else {
		apiKey = data.ApiKey.ValueString()
	}
	if apiKey == "" {
		resp.Diagnostics.AddError("api_key must be set", "Expected a non-empty value for api_key")
		return
	}

	var apiUrl string
	if data.ApiUrl.IsUnknown() || data.ApiUrl.IsNull() {
		apiUrl = os.Getenv("WGC_API_URL")
	} else {
		apiUrl = data.ApiUrl.ValueString()
	}
	if apiUrl == "" {
		apiUrl = "https://cosmo-cp.wundergraph.com"
	}

	var httpClient = http.DefaultClient
	httpClient.Transport = &utils.HeaderTransport{Headers: map[string]string{
		"User-Agent":    utils.GetUserAgent(p.version),
		"Authorization": fmt.Sprintf("Bearer %s", apiKey),
	}}

	// Example client configuration for data sources and resources
	client := platformv1connect.NewPlatformServiceClient(httpClient, apiUrl)

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *WundergraphProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		resources.NewNamespaceResource,
		resources.NewFederatedSubgraphResource,
		resources.NewFederatedGraphResource,
	}
}

func (p *WundergraphProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &WundergraphProvider{
			version: version,
		}
	}
}
