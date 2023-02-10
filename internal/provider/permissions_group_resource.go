package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strconv"
	"terraform-provider-metabase/internal/client"
	"terraform-provider-metabase/internal/schema"
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

type blockTypePermissionsGroup int

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

	createReq := client.PermissionsGroupRequest{
		Name: plan.Name.ValueString(),
	}
	groupId, err := g.provider.client.CreatePermissionsGroup(createReq)
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
	diags = g.provider.syncPermissionsGroupWithApi(&group)
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

	diags = g.provider.syncPermissionsGroupWithApi(&state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

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
	updateReq := client.PermissionsGroupRequest{
		Name: plan.Name.ValueString(),
	}
	err := g.provider.client.UpdatePermissionsGroup(groupId, updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error updating permissions group with ID %d", groupId),
			fmt.Sprintf("Unexpected error occured: %s", err.Error()),
		)
		return
	}

	// Refresh the state
	diags = g.provider.syncPermissionsGroupWithApi(&state)
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
	err := g.provider.client.DeletePermissionsGroup(groupId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting permissions group",
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
	diags := g.provider.syncPermissionsGroupWithApi(&state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Store the state
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func mapPermissionsGroupToState(group *client.PermissionsGroup, target *PermissionsGroupModel) {
	target.Id = types.Int64Value(group.Id)
	target.Name = types.StringValue(group.Name)
}

func (p *MetabaseProvider) syncPermissionsGroupWithApi(state *PermissionsGroupModel) diag.Diagnostics {
	groupId := state.Id.ValueInt64()

	groupDetails, err := p.client.GetPermissionsGroup(groupId)
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
