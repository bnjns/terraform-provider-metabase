package provider

import (
	"context"
	"fmt"
	"github.com/bnjns/metabase-sdk-go/service/permissions"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strconv"
	"terraform-provider-metabase/internal/schema"
	"terraform-provider-metabase/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &PermissionsGroupResource{}
var _ resource.ResourceWithImportState = &PermissionsGroupResource{}

type PermissionsGroupResource struct {
	provider *MetabaseProvider
}

type PermissionsGroupModel struct {
	Id   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (g *PermissionsGroupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_permissions_group"
}

func (g *PermissionsGroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.PermissionsGroupResource()
}

func (g *PermissionsGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PermissionsGroupModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupId, err := g.provider.client.Permissions.CreateGroup(ctx, &permissions.CreateGroupRequest{
		Name: plan.Name.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating permissions group",
			fmt.Sprintf("Unexpected error occured: %s", err.Error()),
		)
		return
	}

	// Refresh the state
	var group PermissionsGroupModel
	group.Id = types.Int64Value(groupId)
	diags = g.provider.syncPermissionsGroupWithApi(ctx, &group)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Ensure we have a consistent plan for any known plan values
	group.ensureConsistentPlan(&plan)

	// Update the state
	diags = resp.State.Set(ctx, group)
	resp.Diagnostics.Append(diags...)
}

func (g *PermissionsGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PermissionsGroupModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupId := state.Id.ValueInt64()
	group, err := g.provider.client.Permissions.GetGroup(ctx, groupId)
	if err != nil {
		diags = utils.HandleResourceReadError(ctx, "permissions group", groupId, err, resp)
		resp.Diagnostics.Append(diags...)
		return
	}

	mapPermissionsGroupToState(group, &state)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (g *PermissionsGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan PermissionsGroupModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state PermissionsGroupModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update the permissions group
	groupId := state.Id.ValueInt64()
	err := g.provider.client.Permissions.UpdateGroup(ctx, groupId, &permissions.UpdateGroupRequest{
		Name: plan.Name.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error updating permissions group with ID %d", groupId),
			fmt.Sprintf("Unexpected error occured: %s", err.Error()),
		)
		return
	}

	// Refresh the state
	diags = g.provider.syncPermissionsGroupWithApi(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Ensure we have a consistent plan for any known plan values
	state.ensureConsistentPlan(&plan)

	// Update the state
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (g *PermissionsGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var group PermissionsGroupModel
	diags := req.State.Get(ctx, &group)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupId := group.Id.ValueInt64()
	err := g.provider.client.Permissions.DeleteGroup(ctx, groupId)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error deleting permissions group with ID %d", groupId),
			fmt.Sprintf("Unexpected error occurred: %s", err.Error()),
		)
		return
	}
}

func (g *PermissionsGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	groupId, _ := strconv.ParseInt(req.ID, 10, 64)

	// Refresh the state from the API
	var state PermissionsGroupModel
	state.Id = types.Int64Value(groupId)
	diags := g.provider.syncPermissionsGroupWithApi(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Store the state
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func mapPermissionsGroupToState(group *permissions.Group, target *PermissionsGroupModel) {
	target.Id = types.Int64Value(group.Id)
	target.Name = types.StringValue(group.Name)
}

func (p *MetabaseProvider) syncPermissionsGroupWithApi(ctx context.Context, state *PermissionsGroupModel) diag.Diagnostics {
	groupId := state.Id.ValueInt64()

	groupDetails, err := p.client.Permissions.GetGroup(ctx, groupId)
	if err != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				fmt.Sprintf("Failed to get permissions group with ID %d", groupId),
				fmt.Sprintf("An error occurred: %s", err.Error()),
			),
		}
	}

	mapPermissionsGroupToState(groupDetails, state)
	return diag.Diagnostics{}
}

func (state *PermissionsGroupModel) ensureConsistentPlan(plan *PermissionsGroupModel) {
	if !plan.Name.IsUnknown() {
		state.Name = plan.Name
	}
}
