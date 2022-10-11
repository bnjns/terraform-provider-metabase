package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &CurrentUserDataSource{}

type CurrentUserDataSource struct {
	provider *MetabaseProvider
}

func (t *CurrentUserDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_current_user"
}

func (t *CurrentUserDataSource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Gets the details of the currently logged-in user.",
		Attributes:  getUserAttributes(blockTypeDataSourceCurrentUser),
	}, nil
}

func (t *CurrentUserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	currentUserDetails, err := t.provider.client.GetCurrentUser()
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get current user",
			"An error occurred: "+err.Error(),
		)
		return
	}

	var data UserResourceModel
	mapUserToState(currentUserDetails, &data)

	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
