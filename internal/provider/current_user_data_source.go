package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &CurrentUserDataSource{}

type CurrentUserDataSource struct {
	provider *MetabaseProvider
}

func (t *CurrentUserDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_current_user"
}

func (t *CurrentUserDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Gets the details of the currently logged-in user.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The ID of the user.",
				Computed:    true,
			},
			"email": schema.StringAttribute{
				Description: "The email address of the user.",
				Computed:    true,
			},
			"first_name": schema.StringAttribute{
				Description: "The first name of the user.",
				Computed:    true,
			},
			"last_name": schema.StringAttribute{
				Description: "The last name of the user.",
				Computed:    true,
			},
			"common_name": schema.StringAttribute{
				Description: "The user's common name, which is a combination of their first and last names.",
				Computed:    true,
			},
			"locale": schema.StringAttribute{
				Description: "The locale the user has configured for themselves. The site default is used if this is nil.",
				Computed:    true,
			},
			"group_ids": schema.ListAttribute{
				ElementType: types.Int64Type,
				Description: "The IDs of the user groups the user is a member of.",
				Computed:    true,
			},
			"google_auth": schema.BoolAttribute{
				Description: "Whether the user was created via Google SSO. Note, if this is enabled then username/password log-in will not be possible.",
				Computed:    true,
			},
			"ldap_auth": schema.BoolAttribute{
				Description: "Whether the user was created via LDAP. Note, if this is enabled then username/password log-in will not be possible.",
				Computed:    true,
			},
			"is_active": schema.BoolAttribute{
				Description: "Used to indicate whether a user is active or if they've been deleted.",
				Computed:    true,
			},
			"is_installer": schema.BoolAttribute{
				Computed: true,
			},
			"is_qbnewb": schema.BoolAttribute{
				Description: "If false then the user has been introduced to how the Query Builder works.",
				Computed:    true,
			},
			"is_superuser": schema.BoolAttribute{
				Description: "Whether the user is a member of the built-in Admin group.",
				Computed:    true,
			},
			"has_invited_second_user": schema.BoolAttribute{
				Computed: true,
			},
			"has_question_and_dashboard": schema.BoolAttribute{
				Computed: true,
			},
			"date_joined": schema.StringAttribute{
				Description: "The timestamp of when the user was created.",
				Computed:    true,
			},
			"first_login": schema.StringAttribute{
				Description: "The timestamp of when the user first logged into Metabase.",
				Computed:    true,
			},
			"last_login": schema.StringAttribute{
				Description: "The timestamp of the user's most recent login to Metabase.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "The timestamp of when the user was last updated.",
				Computed:    true,
			},
		},
	}
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
