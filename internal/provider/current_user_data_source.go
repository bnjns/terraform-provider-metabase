package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ provider.DataSourceType = currentUserDataSourceType{}
var _ datasource.DataSource = currentUserDataSource{}

type currentUserDataSourceType struct{}
type currentUserDataSource struct {
	provider metabaseProvider
}

func (t currentUserDataSourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Gets the details of the currently logged-in user.",
		Attributes:  UserAttributes(false),
	}, nil
}

func (t currentUserDataSourceType) NewDataSource(ctx context.Context, in provider.Provider) (datasource.DataSource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)

	return currentUserDataSource{
		provider: provider,
	}, diags
}

func (t currentUserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	currentUserDetails, err := t.provider.client.GetCurrentUser()
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get current user",
			"An error occurred: "+err.Error(),
		)
		return
	}

	var data userDataSourceData
	MapToState(currentUserDetails, &data)

	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
