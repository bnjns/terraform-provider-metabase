package provider

import (
	"context"
	"fmt"
	"os"

	"terraform-provider-metabase/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ tfsdk.Provider = &provider{}

type provider struct {
	client     *client.Client
	configured bool
	version    string
}

type providerData struct {
	Host     types.String `tfsdk:"host"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

func (p *provider) Configure(ctx context.Context, req tfsdk.ConfigureProviderRequest, resp *tfsdk.ConfigureProviderResponse) {
	var config providerData
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	var host string
	if config.Host.Null || config.Host.Unknown {
		test := os.Environ()
		fmt.Print(test)
		host = os.Getenv("METABASE_HOST")
	} else {
		host = config.Host.Value
	}

	if host == "" {
		resp.Diagnostics.AddError(
			"Unable to create client",
			"Missing required value for host. Either provide explicitly in the provider config or set the METABASE_HOST environment variable.",
		)
	}

	var username string
	if config.Username.Null || config.Username.Unknown {
		username = os.Getenv("METABASE_USERNAME")
	} else {
		username = config.Username.Value
	}

	if username == "" {
		resp.Diagnostics.AddError(
			"Unable to create client",
			"Missing required value for username. Either provide explicitly in the provider config or set the METABASE_USERNAME environment variable.",
		)
	}

	var password string
	if config.Password.Null || config.Password.Unknown {
		password = os.Getenv("METABASE_PASSWORD")
	} else {
		password = config.Password.Value
	}

	if password == "" {
		resp.Diagnostics.AddError(
			"Unable to create client",
			"Missing required value for password. Either provide explicitly in the provider config or set the METABASE_PASSWORD environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	c, err := client.NewClient(host, username, password)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create client",
			"An error occurred when creating the client:"+err.Error(),
		)
		return
	}

	p.client = c
	p.configured = true
}

func (p *provider) GetResources(ctx context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
	return map[string]tfsdk.ResourceType{}, nil
}

func (p *provider) GetDataSources(ctx context.Context) (map[string]tfsdk.DataSourceType, diag.Diagnostics) {
	return map[string]tfsdk.DataSourceType{
		"metabase_current_user": currentUserDataSourceType{},
		"metabase_user":         userDataSourceType{},
	}, nil
}

func (p *provider) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"host": {
				Type:        types.StringType,
				Description: "The Host URL of the Metabase instance to manage. Can also be set with the METABASE_HOST environment variable.",
				Optional:    true,
			},
			"username": {
				Type:        types.StringType,
				Description: "The username of the super user to use when interacting with Metabase. Can also be set with the METABASE_USERNAME environment variable.",
				Optional:    true,
			},
			"password": {
				Type:        types.StringType,
				Description: "The password of the super user to use when interacting with Metabase. Can also be set with the METABASE_PASSWORD environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}, nil
}

func New(version string) func() tfsdk.Provider {
	return func() tfsdk.Provider {
		return &provider{
			version: version,
		}
	}
}

// convertProviderType is a helper function for NewResource and NewDataSource
// implementations to associate the concrete provider type. Alternatively,
// this helper can be skipped and the provider type can be directly type
// asserted (e.g. provider: in.(*provider)), however using this can prevent
// potential panics.
func convertProviderType(in tfsdk.Provider) (provider, diag.Diagnostics) {
	var diags diag.Diagnostics

	p, ok := in.(*provider)

	if !ok {
		diags.AddError(
			"Unexpected Provider Instance Type",
			fmt.Sprintf("While creating the data source or resource, an unexpected provider type (%T) was received. This is always a bug in the provider code and should be reported to the provider developers.", p),
		)
		return provider{}, diags
	}

	if p == nil {
		diags.AddError(
			"Unexpected Provider Instance Type",
			"While creating the data source or resource, an unexpected empty provider instance was received. This is always a bug in the provider code and should be reported to the provider developers.",
		)
		return provider{}, diags
	}

	return *p, diags
}
