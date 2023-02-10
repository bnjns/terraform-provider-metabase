package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"terraform-provider-metabase/internal/schema"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &UserDataSource{}

type UserDataSource struct {
	provider *MetabaseProvider
}

func (t *UserDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (t *UserDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.UserDataSource(schema.DataSourceTypeUser)
}

func (t *UserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Get the current state, in the desired struct
	var data UserResourceModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = t.provider.syncUserWithApi(&data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
