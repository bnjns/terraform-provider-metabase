package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bnjns/metabase-sdk-go/service/database"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/exp/slices"
	"regexp"
	"terraform-provider-metabase/internal/schema"
)

var sensitiveDatabaseDetails = []string{"password", "service-account-json"}
var redactedPattern = regexp.MustCompile(`^\*\*.+\*\*$`)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &DatabaseDataSource{}

type DatabaseDataSourceModel struct {
	Id     types.Int64  `tfsdk:"id"`
	Engine types.String `tfsdk:"engine"`
	Name   types.String `tfsdk:"name"`

	Features  types.List   `tfsdk:"features"`
	Details   types.String `tfsdk:"details"`
	Schedules types.Object `tfsdk:"schedules"`
}

type DatabaseDataSource struct {
	provider *MetabaseProvider
}

func (d DatabaseDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database"
}

func (d DatabaseDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.DatabaseDataSource()
}

func (d DatabaseDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state DatabaseDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	databaseId := state.Id.ValueInt64()
	db, err := d.provider.client.Database.Get(ctx, databaseId)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error fetching database with ID: %d", databaseId),
			fmt.Sprintf("An error occurred: %s", err.Error()),
		)
		return
	}

	diags = mapDatabaseToDataSource(ctx, db, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func mapDatabaseToDataSource(ctx context.Context, db *database.Database, target *DatabaseDataSourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	target.Engine = types.StringValue(string(db.Engine))
	target.Name = types.StringValue(db.Name)
	target.Features, _ = types.ListValueFrom(ctx, types.StringType, db.Features)

	schedules, scheduleDiags := buildSchedules(db)
	target.Schedules = schedules
	diags.Append(scheduleDiags...)

	details := make(map[string]interface{})
	for k, v := range *db.Details {
		if shouldIncludeDatabaseDetail(k, v) {
			details[k] = v
		}
	}
	detailsBytes, detailsErr := json.Marshal(details)
	if detailsErr != nil {
		diags.AddError(
			fmt.Sprintf("Error fetching database with ID: %d", db.Id),
			fmt.Sprintf("An error occurred when serialising the details: %s", detailsErr.Error()),
		)
	}
	target.Details = types.StringValue(string(detailsBytes))

	return diags
}

func shouldIncludeDatabaseDetail(key string, value interface{}) bool {
	if slices.Contains(sensitiveDatabaseDetails, key) {
		return false
	}

	valueStr, isString := value.(string)
	if !isString {
		return true
	}

	if redactedPattern.MatchString(valueStr) {
		return false
	}

	return true
}
