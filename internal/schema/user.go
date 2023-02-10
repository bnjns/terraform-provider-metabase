package schema

import (
	dSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-metabase/internal/modifiers"
	"terraform-provider-metabase/internal/validators"
)

type dataSourceType int

const (
	DataSourceTypeUser dataSourceType = iota
	DataSourceTypeCurrentUser
)

func UserResource() rSchema.Schema {
	return rSchema.Schema{
		Description: "Allows for creating and managing users in Metabase.",
		Attributes: map[string]rSchema.Attribute{
			"id": rSchema.Int64Attribute{
				Description: "The ID of the user.",
				Computed:    true,
			},
			"email": rSchema.StringAttribute{
				Description: "The email address of the user.",
				Required:    true,
				Validators: []validator.String{
					validators.NotEmptyStringValidator(),
				},
			},
			"first_name": rSchema.StringAttribute{
				Description: "The first name of the user.",
				Optional:    true,
				Validators: []validator.String{
					validators.NotEmptyStringValidator(),
				},
			},
			"last_name": rSchema.StringAttribute{
				Description: "The last name of the user.",
				Optional:    true,
				Validators: []validator.String{
					validators.NotEmptyStringValidator(),
				},
			},
			"common_name": rSchema.StringAttribute{
				Description: "The user's common name, which is a combination of their first and last names.",
				Computed:    true,
			},
			"locale": rSchema.StringAttribute{
				Description: "The locale the user has configured for themselves. The site default is used if this is nil.",
				Optional:    true,
			},
			"group_ids": rSchema.ListAttribute{
				ElementType: types.Int64Type,
				Description: "The IDs of the user groups the user is a member of. The 'All Users' group is automatically added by Metabase and you can use `is_superuser` to add the user to the 'Administrators' group.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.List{
					validators.UserNotInReservedGroupsValidator(),
				},
				PlanModifiers: []planmodifier.List{
					modifiers.DefaultToEmptyListModifier(types.Int64Type),
				},
			},
			"google_auth": rSchema.BoolAttribute{
				Description: "Whether the user was created via Google SSO. Note, if this is enabled then username/password log-in will not be possible.",
				Computed:    true,
			},
			"ldap_auth": rSchema.BoolAttribute{
				Description: "Whether the user was created via LDAP. Note, if this is enabled then username/password log-in will not be possible.",
				Computed:    true,
			},
			"is_active": rSchema.BoolAttribute{
				Description: "Used to indicate whether a user is active or if they've been deleted.",
				Computed:    true,
			},
			"is_installer": rSchema.BoolAttribute{
				Computed: true,
			},
			"is_qbnewb": rSchema.BoolAttribute{
				Description: "If false then the user has been introduced to how the Query Builder works.",
				Computed:    true,
			},
			"is_superuser": rSchema.BoolAttribute{
				Description: "Whether the user is a member of the built-in Admin group.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					modifiers.DefaultToFalseModifier(),
				},
			},
			"has_invited_second_user": rSchema.BoolAttribute{
				Computed: true,
			},
			"has_question_and_dashboard": rSchema.BoolAttribute{
				Computed: true,
			},
			"date_joined": rSchema.StringAttribute{
				Description: "The timestamp of when the user was created.",
				Computed:    true,
			},
			"first_login": rSchema.StringAttribute{
				Description: "The timestamp of when the user first logged into Metabase.",
				Computed:    true,
			},
			"last_login": rSchema.StringAttribute{
				Description: "The timestamp of the user's most recent login to Metabase.",
				Computed:    true,
			},
			"updated_at": rSchema.StringAttribute{
				Description: "The timestamp of when the user was last updated.",
				Computed:    true,
			},
		},
	}
}

func (t dataSourceType) makeDescription() string {
	if t == DataSourceTypeUser {
		return "Gets the details of the provided user."
	} else if t == DataSourceTypeCurrentUser {
		return "Gets the details of the currently logged-in user."
	} else {
		return ""
	}
}

func UserDataSource(dataSourceType dataSourceType) dSchema.Schema {

	return dSchema.Schema{
		Description: dataSourceType.makeDescription(),
		Attributes: map[string]dSchema.Attribute{
			"id": dSchema.Int64Attribute{
				Description: "The ID of the user.",
				Required:    dataSourceType == DataSourceTypeUser,
				Computed:    dataSourceType == DataSourceTypeCurrentUser,
			},
			"email": dSchema.StringAttribute{
				Description: "The email address of the user.",
				Computed:    true,
			},
			"first_name": dSchema.StringAttribute{
				Description: "The first name of the user.",
				Computed:    true,
			},
			"last_name": dSchema.StringAttribute{
				Description: "The last name of the user.",
				Computed:    true,
			},
			"common_name": dSchema.StringAttribute{
				Description: "The user's common name, which is a combination of their first and last names.",
				Computed:    true,
			},
			"locale": dSchema.StringAttribute{
				Description: "The locale the user has configured for themselves. The site default is used if this is nil.",
				Computed:    true,
			},
			"group_ids": dSchema.ListAttribute{
				ElementType: types.Int64Type,
				Description: "The IDs of the user groups the user is a member of.",
				Computed:    true,
			},
			"google_auth": dSchema.BoolAttribute{
				Description: "Whether the user was created via Google SSO. Note, if this is enabled then username/password log-in will not be possible.",
				Computed:    true,
			},
			"ldap_auth": dSchema.BoolAttribute{
				Description: "Whether the user was created via LDAP. Note, if this is enabled then username/password log-in will not be possible.",
				Computed:    true,
			},
			"is_active": dSchema.BoolAttribute{
				Description: "Used to indicate whether a user is active or if they've been deleted.",
				Computed:    true,
			},
			"is_installer": dSchema.BoolAttribute{
				Computed: true,
			},
			"is_qbnewb": dSchema.BoolAttribute{
				Description: "If false then the user has been introduced to how the Query Builder works.",
				Computed:    true,
			},
			"is_superuser": dSchema.BoolAttribute{
				Description: "Whether the user is a member of the built-in Admin group.",
				Computed:    true,
			},
			"has_invited_second_user": dSchema.BoolAttribute{
				Computed: true,
			},
			"has_question_and_dashboard": dSchema.BoolAttribute{
				Computed: true,
			},
			"date_joined": dSchema.StringAttribute{
				Description: "The timestamp of when the user was created.",
				Computed:    true,
			},
			"first_login": dSchema.StringAttribute{
				Description: "The timestamp of when the user first logged into Metabase.",
				Computed:    true,
			},
			"last_login": dSchema.StringAttribute{
				Description: "The timestamp of the user's most recent login to Metabase.",
				Computed:    true,
			},
			"updated_at": dSchema.StringAttribute{
				Description: "The timestamp of when the user was last updated.",
				Computed:    true,
			},
		},
	}
}
