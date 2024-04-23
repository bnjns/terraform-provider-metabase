package provider

import (
	"context"
	"fmt"
	"github.com/bnjns/metabase-sdk-go/service/permissions"
	"github.com/bnjns/metabase-sdk-go/service/user"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/exp/slices"
	"strconv"
	"strings"
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

	groupMemberships := mapToGroupMemberships(transforms.FromTerraformInt64List(plan.GroupIds))

	userId, err := u.provider.client.User.Create(ctx, &user.CreateRequest{
		Email:            plan.Email.ValueString(),
		FirstName:        transforms.FromTerraformString(plan.FirstName),
		LastName:         transforms.FromTerraformString(plan.LastName),
		GroupMemberships: groupMemberships,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating user",
			fmt.Sprintf("Unexpected error occured: %s", err.Error()),
		)
		return
	}

	// If the `is_superuser` attribute is set to true we need to update the user
	if !plan.IsSuperuser.IsNull() && !plan.IsSuperuser.IsUnknown() && plan.IsSuperuser.ValueBool() {
		err := u.provider.client.User.Update(ctx, userId, &user.UpdateRequest{
			Email:            transforms.FromTerraformString(plan.Email),
			FirstName:        transforms.FromTerraformString(plan.FirstName),
			LastName:         transforms.FromTerraformString(plan.LastName),
			IsSuperuser:      transforms.FromTerraformBool(plan.IsSuperuser),
			GroupMemberships: groupMemberships,
		})
		if err != nil {
			resp.Diagnostics.AddWarning(
				"User partially created",
				fmt.Sprintf("User with ID %d was created but an error occurred when marking them as a superuser: %s. Try re-applying.", userId, err.Error()),
			)
		}
	}

	// Refresh the state
	var userState UserResourceModel
	userState.Id = types.Int64Value(userId)
	diags = u.provider.syncUserWithApi(ctx, &userState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Ensure we have a consistent plan for any known plan values
	userState.ensureConsistentCreate(&plan)

	// Update the state
	diags = resp.State.Set(ctx, userState)
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
	usr, err := u.provider.client.User.Get(ctx, userId)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			// If the user is not found, attempt to reactivate in case they were manually deactivated
			err = u.provider.client.User.Reactivate(ctx, userId)

			if err == nil {
				// If no error when reactivating, then re-fetch the user and continue with the read
				usr, err = u.provider.client.User.Get(ctx, userId)
				if err != nil {
					addUserReadError(userId, err)
					return
				}
			} else if strings.Contains(err.Error(), "not found") {
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

	mapUserToState(usr, &state)
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

	// Update the user
	userId := state.Id.ValueInt64()
	err := u.provider.client.User.Update(ctx, userId, &user.UpdateRequest{
		Email:            transforms.FromTerraformString(plan.Email),
		FirstName:        transforms.FromTerraformString(plan.FirstName),
		LastName:         transforms.FromTerraformString(plan.LastName),
		Locale:           transforms.FromTerraformString(plan.Locale),
		IsSuperuser:      transforms.FromTerraformBool(plan.IsSuperuser),
		GroupMemberships: mapToGroupMemberships(transforms.FromTerraformInt64List(plan.GroupIds)),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error updating user with ID %d", userId),
			fmt.Sprintf("Unexpected error occured: %s", err.Error()),
		)
		return
	}

	// Refresh the state
	diags = u.provider.syncUserWithApi(ctx, &state)
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
	var userState UserResourceModel
	diags := req.State.Get(ctx, &userState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	userId := userState.Id.ValueInt64()
	err := u.provider.client.User.Disable(ctx, userId)
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
	err := u.provider.client.User.Reactivate(ctx, userId)
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
	diags := u.provider.syncUserWithApi(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Store the state
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func buildGroupIdList(user *user.User, state *UserResourceModel) []int64 {
	var isReservedGroup = func(groupId int64) bool {
		return slices.Contains(validators.ReservedGroupIds, groupId)
	}

	groupIds := make([]int64, 0)

	// Convert the group memberships from the API into a list of IDs
	apiGroupIds := make([]int64, len(user.GroupMemberships))
	for i, membership := range user.GroupMemberships {
		apiGroupIds[i] = membership.Id
	}

	// If the target (aka the current state) has the group IDs set, then use this to initialise the group IDs.
	// This will ensure that we retain the ordering.
	if !state.GroupIds.IsUnknown() && !state.GroupIds.IsNull() {
		for _, groupIdEl := range state.GroupIds.Elements() {
			groupId := groupIdEl.(types.Int64).ValueInt64()
			// Only add this group from state if it's also in the API response and isn't reserved
			if slices.Contains(apiGroupIds, groupId) && !isReservedGroup(groupId) {
				groupIds = append(groupIds, groupId)
			}
		}
	}

	// Now iterate through the group IDs from the API and add any that are missing
	for _, groupId := range apiGroupIds {
		if !slices.Contains(groupIds, groupId) && !isReservedGroup(groupId) {
			groupIds = append(groupIds, groupId)
		}
	}

	return groupIds
}

func mapUserToState(user *user.User, target *UserResourceModel) {
	groupIds := buildGroupIdList(user, target)

	target.Id = types.Int64Value(user.Id)
	target.FirstName = transforms.ToTerraformString(user.FirstName)
	target.LastName = transforms.ToTerraformString(user.LastName)
	target.CommonName = transforms.ToTerraformString(user.CommonName)
	target.Email = types.StringValue(user.Email)
	target.Locale = transforms.ToTerraformString(user.Locale)

	target.IsActive = types.BoolValue(user.IsActive)
	target.IsQbnewb = types.BoolValue(user.IsQbnewb)
	target.IsSuperuser = types.BoolValue(user.IsSuperuser)
	target.IsInstaller = transforms.ToTerraformBool(user.IsInstaller)

	target.GroupIds = transforms.ToTerraformInt64List(&groupIds)

	target.GoogleAuth = types.BoolValue(user.GoogleAuth)
	// TODO: sso source?

	target.HasInvitedSecondUser = types.BoolValue(user.HasInvitedSecondUser)
	target.HasQuestionAndDashboard = types.BoolValue(user.HasQuestionAndDashboard)
	// TODO: personal collection ID?

	target.DateJoined = types.StringValue(user.DateJoined)
	target.FirstLogin = transforms.ToTerraformString(user.FirstLogin)
	target.LastLogin = transforms.ToTerraformString(user.LastLogin)
	target.UpdatedAt = transforms.ToTerraformString(user.UpdatedAt)
}

func (p *MetabaseProvider) syncUserWithApi(ctx context.Context, state *UserResourceModel) diag.Diagnostics {
	userId := state.Id.ValueInt64()

	userDetails, err := p.client.User.Get(ctx, userId)
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

func mapToGroupMemberships(groupIds *[]int64) *[]user.GroupMembership {
	if groupIds == nil || len(*groupIds) == 0 {
		return nil
	}

	groupMemberships := make([]user.GroupMembership, len(*groupIds))
	for i, groupId := range *groupIds {
		groupMemberships[i] = user.GroupMembership{
			Id: groupId,
		}
	}
	return &groupMemberships
}

func addReservedUserGroups(plan UserResourceModel) *[]int64 {
	// Add the reserved groups so we don't upset Metabase
	groupIds := transforms.FromTerraformInt64List(plan.GroupIds)
	if groupIds != nil {
		if !slices.Contains(*groupIds, permissions.GroupAllUsers) {
			*groupIds = append(*groupIds, permissions.GroupAllUsers)
		}
		if !slices.Contains(*groupIds, permissions.GroupAdministrators) && plan.IsSuperuser.ValueBool() {
			*groupIds = append(*groupIds, permissions.GroupAdministrators)
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
