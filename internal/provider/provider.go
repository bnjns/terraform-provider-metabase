package provider

import (
	"context"
	"fmt"
	"github.com/bnjns/metabase-sdk-go/metabase"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-metabase/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ provider.Provider = &MetabaseProvider{}

type MetabaseProvider struct {
	client     *metabase.Client
	configured bool
	version    string
}

type MetabaseProviderModel struct {
	Host     types.String `tfsdk:"host"`
	ApiKey   types.String `tfsdk:"api_key"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	Headers  types.Map    `tfsdk:"headers"`
}

func (p *MetabaseProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "metabase"
}

func (p *MetabaseProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config MetabaseProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	host := utils.GetConfigValue(config.Host, "METABASE_HOST")
	if host == "" {
		resp.Diagnostics.AddError(
			"Unable to create client",
			"Missing required value for host. Either provide explicitly in the provider config or set the METABASE_HOST environment variable.",
		)
	}

	headers := make(map[string]string)
	if !config.Headers.IsNull() {
		config.Headers.ElementsAs(ctx, &headers, true)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	metabaseAuth, err := createAuth(config)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create client",
			fmt.Sprintf("An error occurred when configuring the authentication: %s", err.Error()),
		)
		return
	}
	client, err := metabase.NewClient(host, metabaseAuth, metabase.WithHeaders(headers))
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create client",
			fmt.Sprintf("An error occurred when creating the client: %s", err.Error()),
		)
		return
	}

	p.client = client
	p.configured = true
}

func (p *MetabaseProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		func() resource.Resource {
			return &DatabaseResource{provider: p}
		},
		func() resource.Resource {
			return &PermissionsGroupResource{provider: p}
		},
		func() resource.Resource {
			return &UserResource{provider: p}
		},
	}
}

func (p *MetabaseProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		func() datasource.DataSource {
			return &CurrentUserDataSource{provider: p}
		},
		func() datasource.DataSource {
			return &DatabaseDataSource{provider: p}
		},
		func() datasource.DataSource {
			return &PermissionsGroupDataSource{provider: p}
		},
		func() datasource.DataSource {
			return &UserDataSource{provider: p}
		},
	}
}

func (p *MetabaseProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Description: "The Host URL of the Metabase instance to manage. Can also be set with the METABASE_HOST environment variable.",
				Optional:    true,
			},
			"api_key": schema.StringAttribute{
				Description: "The API key to use for authenticating with Metabase. Can also be set with the METABASE_API_KEY environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"username": schema.StringAttribute{
				Description: "The username of the super user to use when interacting with Metabase. Can also be set with the METABASE_USERNAME environment variable.",
				Optional:    true,
			},
			"password": schema.StringAttribute{
				Description: "The password of the super user to use when interacting with Metabase. Can also be set with the METABASE_PASSWORD environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"headers": schema.MapAttribute{
				ElementType: types.StringType,
				Description: "Optional headers to attach to every request to Metabase.",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &MetabaseProvider{
			version: version,
		}
	}
}

func createAuth(config MetabaseProviderModel) (metabase.Authenticator, error) {
	apiKey := utils.GetConfigValue(config.ApiKey, "METABASE_API_KEY")
	username := utils.GetConfigValue(config.Username, "METABASE_USERNAME")
	password := utils.GetConfigValue(config.Password, "METABASE_PASSWORD")

	if apiKey != "" {
		return metabase.NewApiKeyAuthenticator(apiKey)
	} else if username != "" && password != "" {
		return metabase.NewSessionAuthenticator(username, password)
	} else {
		return nil, fmt.Errorf("you must set either the API key (via the api_key attribute or METABASE_API_KEY environment variable) or username and password (via the username and password attributes, or METABASE_USERNAME and METABASE_PASSWORD environment variables)")
	}
}
