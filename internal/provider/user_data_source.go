package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"terraform-provider-metabase/internal/client"
	"terraform-provider-metabase/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ tfsdk.DataSourceType = userDataSourceType{}
var _ tfsdk.DataSource = userDataSource{}

type userDataSourceType struct{}
type userDataSourceData struct {
	Id         types.Int64  `tfsdk:"id"`
	Email      types.String `tfsdk:"email"`
	FirstName  types.String `tfsdk:"first_name"`
	LastName   types.String `tfsdk:"last_name"`
	CommonName types.String `tfsdk:"common_name"`
	Locale     types.String `tfsdk:"locale"`
	GroupIds   types.List   `tfsdk:"group_ids"`

	GoogleAuth types.Bool `tfsdk:"google_auth"`
	LdapAuth   types.Bool `tfsdk:"ldap_auth"`

	IsActive                types.Bool `tfsdk:"is_active"`
	IsInstaller             types.Bool `tfsdk:"is_installer"`
	IsQbnewb                types.Bool `tfsdk:"is_qbnewb"`
	IsSuperuser             types.Bool `tfsdk:"is_superuser"`
	HasInvitedSecondUser    types.Bool `tfsdk:"has_invited_second_user"`
	HasQuestionAndDashboard types.Bool `tfsdk:"has_question_and_dashboard"`

	DateJoined types.String `tfsdk:"date_joined"`
	FirstLogin types.String `tfsdk:"first_login"`
	LastLogin  types.String `tfsdk:"last_login"`
	UpdatedAt  types.String `tfsdk:"updated_at"`
}
type userDataSource struct {
	provider provider
}

func UserAttributes(userSpecifiedId bool) map[string]tfsdk.Attribute {
	return map[string]tfsdk.Attribute{
		"id": {
			Type:        types.Int64Type,
			Description: "The ID of the user.",
			Required:    userSpecifiedId,
			Computed:    !userSpecifiedId,
		},
		"email": {
			Type:        types.StringType,
			Description: "The email address of the user.",
			Computed:    true,
		},
		"first_name": {
			Type:        types.StringType,
			Description: "The first name of the user.",
			Computed:    true,
		},
		"last_name": {
			Type:        types.StringType,
			Description: "The last name of the user.",
			Computed:    true,
		},
		"common_name": {
			Type:        types.StringType,
			Description: "The user's common name, which is a combination of their first and last names.",
			Computed:    true,
		},
		"locale": {
			Type:        types.StringType,
			Description: "The locale the user has configured for themselves. The site default is used if this is nil.",
			Computed:    true,
		},
		"group_ids": {
			Type:        types.ListType{ElemType: types.Int64Type},
			Description: "The IDs of the user groups the user is a member of.",
			Computed:    true,
		},
		"google_auth": {
			Type:     types.BoolType,
			Computed: true,
		},
		"ldap_auth": {
			Type:     types.BoolType,
			Computed: true,
		},
		"is_active": {
			Type:     types.BoolType,
			Computed: true,
		},
		"is_installer": {
			Type:     types.BoolType,
			Computed: true,
		},
		"is_qbnewb": {
			Type:     types.BoolType,
			Computed: true,
		},
		"is_superuser": {
			Type:     types.BoolType,
			Computed: true,
		},
		"has_invited_second_user": {
			Type:     types.BoolType,
			Computed: true,
		},
		"has_question_and_dashboard": {
			Type:     types.BoolType,
			Computed: true,
		},
		"date_joined": {
			Type:     types.StringType,
			Computed: true,
		},
		"first_login": {
			Type:     types.StringType,
			Computed: true,
		},
		"last_login": {
			Type:     types.StringType,
			Computed: true,
		},
		"updated_at": {
			Type:     types.StringType,
			Computed: true,
		},
	}
}

func (t userDataSourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Gets the details of the provided user.",
		Attributes:  UserAttributes(true),
	}, nil
}

func (t userDataSourceType) NewDataSource(ctx context.Context, in tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)

	return userDataSource{
		provider: provider,
	}, diags
}

func (t userDataSource) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	// Get the current state, in the desired struct
	var data userDataSourceData
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	userId := data.Id.Value

	userDetails, err := t.provider.client.GetUser(userId)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Failed to get user with ID %d", userId),
			fmt.Sprintf("An error occurred: %s", err.Error()),
		)
		return
	}

	MapToState(userDetails, &data)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func MapToState(user *client.User, target *userDataSourceData) {
	groupIds := make([]attr.Value, len(user.Groups))
	for i, groupId := range user.Groups {
		groupIds[i] = types.Int64{Value: groupId}
	}

	target.Id = types.Int64{Value: user.Id}
	target.Email = types.String{Value: user.Email}
	target.FirstName = types.String{Value: user.FirstName}
	target.LastName = types.String{Value: user.LastName}
	target.CommonName = types.String{Value: user.CommonName}
	target.Locale = utils.ToTerraformString(user.Locale)
	target.GroupIds = types.List{
		ElemType: types.Int64Type,
		Elems:    groupIds,
	}
	target.GoogleAuth = types.Bool{Value: user.GoogleAuth}
	target.LdapAuth = types.Bool{Value: user.LdapAuth}
	target.IsActive = types.Bool{Value: user.IsActive}
	target.IsInstaller = types.Bool{Value: user.IsInstaller}
	target.IsQbnewb = types.Bool{Value: user.IsQbnewb}
	target.IsSuperuser = types.Bool{Value: user.IsSuperuser}
	target.HasInvitedSecondUser = types.Bool{Value: user.HasInvitedSecondUser}
	target.HasQuestionAndDashboard = types.Bool{Value: user.HasQuestionAndDashboard}
	target.DateJoined = types.String{Value: user.DateJoined}
	target.FirstLogin = utils.ToTerraformString(user.FirstLogin)
	target.LastLogin = utils.ToTerraformString(user.LastLogin)
	target.UpdatedAt = utils.ToTerraformString(user.UpdatedAt)
}
