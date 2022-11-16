package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"os"

	"terraform-provider-metabase/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ provider.Provider = &MetabaseProvider{}

type MetabaseProvider struct {
	client     *client.Client
	configured bool
	version    string
}

type MetabaseProviderModel struct {
	Host     types.String `tfsdk:"host"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
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

	var host string
	if config.Host.IsNull() || config.Host.IsUnknown() {
		test := os.Environ()
		fmt.Print(test)
		host = os.Getenv("METABASE_HOST")
	} else {
		host = config.Host.ValueString()
	}

	if host == "" {
		resp.Diagnostics.AddError(
			"Unable to create client",
			"Missing required value for host. Either provide explicitly in the provider config or set the METABASE_HOST environment variable.",
		)
	}

	var username string
	if config.Username.IsNull() || config.Username.IsUnknown() {
		username = os.Getenv("METABASE_USERNAME")
	} else {
		username = config.Username.ValueString()
	}

	if username == "" {
		resp.Diagnostics.AddError(
			"Unable to create client",
			"Missing required value for username. Either provide explicitly in the provider config or set the METABASE_USERNAME environment variable.",
		)
	}

	var password string
	if config.Password.IsNull() || config.Password.IsUnknown() {
		password = os.Getenv("METABASE_PASSWORD")
	} else {
		password = config.Password.ValueString()
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

func (p *MetabaseProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
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
			return &PermissionsGroupDataSource{provider: p}
		},
		func() datasource.DataSource {
			return &UserDataSource{provider: p}
		},
	}
}

func (p *MetabaseProvider) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
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

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &MetabaseProvider{
			version: version,
		}
	}
}
