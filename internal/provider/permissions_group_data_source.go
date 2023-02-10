package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"terraform-provider-metabase/internal/schema"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &PermissionsGroupDataSource{}

type PermissionsGroupDataSource struct {
	provider *MetabaseProvider
}

func (g *PermissionsGroupDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_permissions_group"
}

func (g *PermissionsGroupDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.PermissionsGroupDataSource()
}

func (g *PermissionsGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Get the current state
	var state PermissionsGroupModel
	diags := req.Config.Get(ctx, &state)
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
