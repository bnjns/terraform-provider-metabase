package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/exp/slices"
	"strconv"
	"terraform-provider-metabase/internal/client"
	"terraform-provider-metabase/internal/modifiers"
	"terraform-provider-metabase/internal/transforms"
	"terraform-provider-metabase/internal/validators"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &UserResource{}
var _ resource.ResourceWithImportState = &UserResource{}

type UserResource struct {
	provider *MetabaseProvider
}

type UserResourceModel struct {
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

type blockTypeUser int

const (
	blockTypeResourceUser blockTypeUser = iota
	blockTypeDataSourceUser
	blockTypeDataSourceCurrentUser
)

func getUserAttributes(t blockTypeUser) map[string]tfsdk.Attribute {
	groupDescription := "The IDs of the user groups the user is a member of."
	if t == blockTypeResourceUser {
		groupDescription += " The 'All Users' group is automatically added by Metabase and you can use `is_superuser` to add the user to the 'Administrators' group."
	}

	return map[string]tfsdk.Attribute{
		"id": {
			Type:        types.Int64Type,
			Description: "The ID of the user.",
			Required:    t == blockTypeDataSourceUser,
			Computed:    t != blockTypeDataSourceUser,
		},
		"email": {
			Type:        types.StringType,
			Description: "The email address of the user.",
			Required:    t == blockTypeResourceUser,
			Computed:    t != blockTypeResourceUser,
			Validators: []tfsdk.AttributeValidator{
				validators.NotEmptyStringValidator(),
			},
		},
		"first_name": {
			Type:        types.StringType,
			Description: "The first name of the user.",
			Optional:    t == blockTypeResourceUser,
			Computed:    t != blockTypeResourceUser,
			Validators: []tfsdk.AttributeValidator{
				validators.NotEmptyStringValidator(),
			},
		},
		"last_name": {
			Type:        types.StringType,
			Description: "The last name of the user.",
			Optional:    t == blockTypeResourceUser,
			Computed:    t != blockTypeResourceUser,
			Validators: []tfsdk.AttributeValidator{
				validators.NotEmptyStringValidator(),
			},
		},
		"common_name": {
			Type:        types.StringType,
			Description: "The user's common name, which is a combination of their first and last names.",
			Computed:    true,
		},
		"locale": {
			Type:        types.StringType,
			Description: "The locale the user has configured for themselves. The site default is used if this is nil.",
			Optional:    t == blockTypeResourceUser,
			Computed:    t != blockTypeResourceUser,
		},
		"group_ids": {
			Type:        types.ListType{ElemType: types.Int64Type},
			Description: groupDescription,
			Optional:    t == blockTypeResourceUser,
			Computed:    true,
			Validators: []tfsdk.AttributeValidator{
				validators.UserNotInReservedGroupsValidator(),
			},
			PlanModifiers: tfsdk.AttributePlanModifiers{
				modifiers.DefaultToEmptyListModifier(types.Int64Type),
			},
		},
		"google_auth": {
			Type:        types.BoolType,
			Description: "Whether the user was created via Google SSO. Note, if this is enabled then username/password log-in will not be possible.",
			Computed:    true,
		},
		"ldap_auth": {
			Type:        types.BoolType,
			Description: "Whether the user was created via LDAP. Note, if this is enabled then username/password log-in will not be possible.",
			Computed:    true,
		},
		"is_active": {
			Type:        types.BoolType,
			Description: "Used to indicate whether a user is active or if they've been deleted.",
			Computed:    true,
		},
		"is_installer": {
			Type:     types.BoolType,
			Computed: true,
		},
		"is_qbnewb": {
			Type:        types.BoolType,
			Description: "If false then the user has been introduced to how the Query Builder works.",
			Computed:    true,
		},
		"is_superuser": {
			Type:        types.BoolType,
			Description: "Whether the user is a member of the built-in Admin group.",
			Optional:    t == blockTypeResourceUser,
			Computed:    true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				modifiers.DefaultToFalseModifier(),
			},
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
			Type:        types.StringType,
			Description: "The timestamp of when the user was created.",
			Computed:    true,
		},
		"first_login": {
			Type:        types.StringType,
			Description: "The timestamp of when the user first logged into Metabase.",
			Computed:    true,
		},
		"last_login": {
			Type:        types.StringType,
			Description: "The timestamp of the user's most recent login to Metabase.",
			Computed:    true,
		},
		"updated_at": {
			Type:        types.StringType,
			Description: "The timestamp of when the user was last updated.",
			Computed:    true,
		},
	}
}

func (u *UserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (u *UserResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Allows for creating and managing users in Metabase.",
		Attributes:  getUserAttributes(blockTypeResourceUser),
	}, nil
}

func (u *UserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan UserResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupMemberships := mapToGroupMemberships(transforms.FromTerraformInt64List(plan.GroupIds))
	var createReq = client.CreateUserRequest{
		Email:            plan.Email.Value,
		FirstName:        transforms.FromTerraformString(plan.FirstName),
		LastName:         transforms.FromTerraformString(plan.LastName),
		GroupMemberships: groupMemberships,
	}

	userId, err := u.provider.client.CreateUser(createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating user",
			fmt.Sprintf("Unexpected error occured: %s", err.Error()),
		)
		return
	}

	// If the `is_superuser` attribute is set to true we need to update the user
	if !plan.IsSuperuser.Null && !plan.IsSuperuser.Unknown && plan.IsSuperuser.Value {
		updateReq := client.UpdateUserRequest{
			Email:            &plan.Email.Value,
			FirstName:        transforms.FromTerraformString(plan.FirstName),
			LastName:         transforms.FromTerraformString(plan.LastName),
			GroupMemberships: groupMemberships,
			IsSuperuser:      transforms.FromTerraformBool(plan.IsSuperuser),
			Locale:           transforms.FromTerraformString(plan.Locale),
		}
		err := u.provider.client.UpdateUser(userId, updateReq)
		if err != nil {
			resp.Diagnostics.AddWarning(
				"User partially created",
				fmt.Sprintf("User with ID %d was created but an error occurred when marking them as a superuser: %s. Try re-applying.", userId, err.Error()),
			)
		}
	}

	// Refresh the state
	var user UserResourceModel
	user.Id = types.Int64{Value: userId}
	diags = u.provider.syncUserWithApi(&user)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Ensure we have a consistent plan for any known plan values
	user.ensureConsistentCreate(&plan)

	// Update the state
	diags = resp.State.Set(ctx, user)
	resp.Diagnostics.Append(diags...)
}

func (u *UserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state UserResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = u.provider.syncUserWithApi(&state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (u *UserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan UserResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state UserResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Add the reserved groups so we don't upset Metabase
	groupIds := transforms.FromTerraformInt64List(plan.GroupIds)
	if groupIds != nil {
		if !slices.Contains(*groupIds, validators.GroupIdAllUsers) {
			*groupIds = append(*groupIds, validators.GroupIdAllUsers)
		}
		if !slices.Contains(*groupIds, validators.GroupIdAdministrators) && plan.IsSuperuser.Value {
			*groupIds = append(*groupIds, validators.GroupIdAdministrators)
		}
	}

	// Update the user
	userId := state.Id.Value
	updateReq := client.UpdateUserRequest{
		Email:            &plan.Email.Value,
		FirstName:        transforms.FromTerraformString(plan.FirstName),
		LastName:         transforms.FromTerraformString(plan.LastName),
		GroupMemberships: mapToGroupMemberships(groupIds),
		IsSuperuser:      transforms.FromTerraformBool(plan.IsSuperuser),
		Locale:           transforms.FromTerraformString(plan.Locale),
	}
	err := u.provider.client.UpdateUser(userId, updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error updating user with ID %d", userId),
			fmt.Sprintf("Unexpected error occured: %s", err.Error()),
		)
		return
	}

	// Refresh the state
	diags = u.provider.syncUserWithApi(&state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Ensure we have a consistent plan for any known plan values
	state.ensureConsistentUpdate(&plan)

	// Update the state
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (u *UserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var user UserResourceModel
	diags := req.State.Get(ctx, &user)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	userId := user.Id.Value
	err := u.provider.client.DeleteUser(userId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting user",
			fmt.Sprintf("Unexpected error occurred: %s", err.Error()),
		)
		return
	}
}

func (u *UserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	userId, _ := strconv.ParseInt(req.ID, 10, 64)

	// TODO: the current approach is a bit hacky as we rely on reactivating the user throwing a 404 and erroring if the
	//  user doesn't exist, but we can't use client.GetUser() with deactivated users. Maybe it would be better to
	//  explicitly search using GET /api/user?include_deactivated=true first?

	// We'll need to reactivate the user if it exists
	err := u.provider.client.ReactivateUser(userId)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error importing user with ID %d", userId),
			fmt.Sprintf("Error occurred when reactivating user: %s", err.Error()),
		)
		return
	}

	// Refresh the state from the API
	var state UserResourceModel
	state.Id = types.Int64{Value: userId}
	diags := u.provider.syncUserWithApi(&state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Store the state
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func mapUserToState(user *client.User, target *UserResourceModel) {
	groupIds := make([]attr.Value, 0)
	for _, membership := range user.GroupMemberships {
		// We need to remove the restricted groups from state so they don't conflict
		if !slices.Contains(validators.ReservedGroupIds, membership.Id) {
			groupIds = append(groupIds, types.Int64{Value: membership.Id})
		}
	}

	target.Id = types.Int64{Value: user.Id}
	target.Email = types.String{Value: user.Email}
	target.FirstName = transforms.ToTerraformString(user.FirstName)
	target.LastName = transforms.ToTerraformString(user.LastName)
	target.CommonName = transforms.ToTerraformString(user.CommonName)
	target.Locale = transforms.ToTerraformString(user.Locale)
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
	target.FirstLogin = transforms.ToTerraformString(user.FirstLogin)
	target.LastLogin = transforms.ToTerraformString(user.LastLogin)
	target.UpdatedAt = transforms.ToTerraformString(user.UpdatedAt)
}

func (p *MetabaseProvider) syncUserWithApi(state *UserResourceModel) diag.Diagnostics {
	userId := state.Id.Value

	userDetails, err := p.client.GetUser(userId)
	if err != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				fmt.Sprintf("Failed to get user with ID %d", userId),
				fmt.Sprintf("An error occurred: %s", err.Error()),
			),
		}
	}

	mapUserToState(userDetails, state)
	return diag.Diagnostics{}
}

func mapToGroupMemberships(groupIds *[]int64) *[]client.GroupMembership {
	if groupIds == nil || len(*groupIds) == 0 {
		return nil
	}

	groupMemberships := make([]client.GroupMembership, len(*groupIds))
	for i, groupId := range *groupIds {
		groupMemberships[i] = client.GroupMembership{
			Id: groupId,
		}
	}
	return &groupMemberships
}

func (state *UserResourceModel) ensureConsistentCreate(plan *UserResourceModel) {
	if !plan.Email.Unknown {
		state.Email = plan.Email
	}
	if !plan.FirstName.Unknown {
		state.FirstName = plan.FirstName
	}
	if !plan.LastName.Unknown {
		state.LastName = plan.LastName
	}
	if !plan.GroupIds.Unknown {
		state.GroupIds = plan.GroupIds
	}
	if !plan.IsSuperuser.Unknown {
		state.IsSuperuser = plan.IsSuperuser
	}
}

func (state *UserResourceModel) ensureConsistentUpdate(plan *UserResourceModel) {
	if !plan.Email.Unknown {
		state.Email = plan.Email
	}
	if !plan.FirstName.Unknown {
		state.FirstName = plan.FirstName
	}
	if !plan.LastName.Unknown {
		state.LastName = plan.LastName
	}
	if !plan.GroupIds.Unknown {
		state.GroupIds = plan.GroupIds
	}
	if !plan.IsSuperuser.Unknown {
		state.IsSuperuser = plan.IsSuperuser
	}
	if !plan.Locale.Unknown {
		state.Locale = plan.Locale
	}
	if !plan.IsActive.Unknown {
		state.IsActive = plan.IsActive
	}
}
