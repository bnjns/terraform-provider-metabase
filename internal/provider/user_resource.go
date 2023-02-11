package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/exp/slices"
	"strconv"
	"terraform-provider-metabase/internal/client"
	"terraform-provider-metabase/internal/schema"
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

func (u *UserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (u *UserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.UserResource()
}

func (u *UserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan UserResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupIds := addReservedUserGroups(plan)
	groupMemberships := mapToGroupMemberships(groupIds)
	var createReq = client.CreateUserRequest{
		Email:            plan.Email.ValueString(),
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
	if !plan.IsSuperuser.IsNull() && !plan.IsSuperuser.IsUnknown() && plan.IsSuperuser.ValueBool() {
		updateReq := client.UpdateUserRequest{
			Email:            transforms.FromTerraformString(plan.Email),
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
	user.Id = types.Int64Value(userId)
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
	addUserReadError := func(userId int64, err error) {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Failed to get user with ID: %d", userId),
			fmt.Sprintf("An unexpected error occurred: %s", err.Error()),
		)
	}

	var state UserResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	userId := state.Id.ValueInt64()
	user, err := u.provider.client.GetUser(userId)
	if err != nil {
		if err == client.ErrNotFound {
			// If the user is not found, attempt to reactivate in case they were manually deactivated
			err = u.provider.client.ReactivateUser(userId)

			if err == nil {
				// If no error when reactivating, then re-fetch the user and continue with the read
				user, err = u.provider.client.GetUser(userId)
				if err != nil {
					addUserReadError(userId, err)
					return
				}
			} else if err == client.ErrNotFound {
				// If reactivating returns a not found error, then remove the resource
				resp.State.RemoveResource(ctx)
				return
			} else {
				// Fall back to reporting an error to the user
				addUserReadError(userId, err)
				return
			}
		} else {
			addUserReadError(userId, err)
			return
		}
	}

	mapUserToState(user, &state)
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
	groupIds := addReservedUserGroups(plan)

	// Update the user
	userId := state.Id.ValueInt64()
	updateReq := client.UpdateUserRequest{
		Email:            transforms.FromTerraformString(plan.Email),
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

	userId := user.Id.ValueInt64()
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
	state.Id = types.Int64Value(userId)
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
			groupIds = append(groupIds, types.Int64Value(membership.Id))
		}
	}

	target.Id = types.Int64Value(user.Id)
	target.Email = types.StringValue(user.Email)
	target.FirstName = transforms.ToTerraformString(user.FirstName)
	target.LastName = transforms.ToTerraformString(user.LastName)
	target.CommonName = transforms.ToTerraformString(user.CommonName)
	target.Locale = transforms.ToTerraformString(user.Locale)
	target.GroupIds, _ = types.ListValue(types.Int64Type, groupIds)
	target.GoogleAuth = types.BoolValue(user.GoogleAuth)
	target.LdapAuth = types.BoolValue(user.LdapAuth)
	target.IsActive = types.BoolValue(user.IsActive)
	target.IsInstaller = types.BoolValue(user.IsInstaller)
	target.IsQbnewb = types.BoolValue(user.IsQbnewb)
	target.IsSuperuser = types.BoolValue(user.IsSuperuser)
	target.HasInvitedSecondUser = types.BoolValue(user.HasInvitedSecondUser)
	target.HasQuestionAndDashboard = types.BoolValue(user.HasQuestionAndDashboard)
	target.DateJoined = types.StringValue(user.DateJoined)
	target.FirstLogin = transforms.ToTerraformString(user.FirstLogin)
	target.LastLogin = transforms.ToTerraformString(user.LastLogin)
	target.UpdatedAt = transforms.ToTerraformString(user.UpdatedAt)
}

func (p *MetabaseProvider) syncUserWithApi(state *UserResourceModel) diag.Diagnostics {
	userId := state.Id.ValueInt64()

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

func addReservedUserGroups(plan UserResourceModel) *[]int64 {
	// Add the reserved groups so we don't upset Metabase
	groupIds := transforms.FromTerraformInt64List(plan.GroupIds)
	if groupIds != nil {
		if !slices.Contains(*groupIds, validators.GroupIdAllUsers) {
			*groupIds = append(*groupIds, validators.GroupIdAllUsers)
		}
		if !slices.Contains(*groupIds, validators.GroupIdAdministrators) && plan.IsSuperuser.ValueBool() {
			*groupIds = append(*groupIds, validators.GroupIdAdministrators)
		}
	}
	return groupIds
}

func (state *UserResourceModel) ensureConsistentCreate(plan *UserResourceModel) {
	if !plan.Email.IsUnknown() {
		state.Email = plan.Email
	}
	if !plan.FirstName.IsUnknown() {
		state.FirstName = plan.FirstName
	}
	if !plan.LastName.IsUnknown() {
		state.LastName = plan.LastName
	}
	if !plan.GroupIds.IsUnknown() {
		state.GroupIds = plan.GroupIds
	}
	if !plan.IsSuperuser.IsUnknown() {
		state.IsSuperuser = plan.IsSuperuser
	}
}

func (state *UserResourceModel) ensureConsistentUpdate(plan *UserResourceModel) {
	if !plan.Email.IsUnknown() {
		state.Email = plan.Email
	}
	if !plan.FirstName.IsUnknown() {
		state.FirstName = plan.FirstName
	}
	if !plan.LastName.IsUnknown() {
		state.LastName = plan.LastName
	}
	if !plan.GroupIds.IsUnknown() {
		state.GroupIds = plan.GroupIds
	}
	if !plan.IsSuperuser.IsUnknown() {
		state.IsSuperuser = plan.IsSuperuser
	}
	if !plan.Locale.IsUnknown() {
		state.Locale = plan.Locale
	}
	if !plan.IsActive.IsUnknown() {
		state.IsActive = plan.IsActive
	}
}
