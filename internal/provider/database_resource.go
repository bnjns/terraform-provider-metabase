package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/exp/slices"
	"terraform-provider-metabase/internal/client"
	"terraform-provider-metabase/internal/schema"
	"terraform-provider-metabase/internal/transforms"
	"terraform-provider-metabase/internal/utils"
)

type DatabaseModel struct {
	Id     types.Int64  `tfsdk:"id"`
	Engine types.String `tfsdk:"engine"`
	Name   types.String `tfsdk:"name"`

	Features      types.List   `tfsdk:"features"`
	Details       types.String `tfsdk:"details"`
	DetailsSecure types.String `tfsdk:"details_secure"`
	Schedules     types.Map    `tfsdk:"schedules"`
}

// Ensure provider fully satisfies the framework interfaces
var _ resource.Resource = &DatabaseResource{}

//var _ resource.ResourceWithImportState = &DatabaseResource{}

var (
	errMissingConnString = errors.New("you must provide the connection string in the 'db' property")
	errMissingDbName     = errors.New("you must provide the database name in the 'dbname' property")
	errMissingHost       = errors.New("you must provide the database hostname/ip in the 'host' property")
	errMissingPort       = errors.New("you must provide the database port in the 'port' property")
	errMissingUser       = errors.New("you must provide the auth username in the 'user' property")
	errMissingPassword   = errors.New("you must provide the auth password in the 'password' property")
)
var databaseSensitiveProperties = []string{"password"}

type DatabaseResource struct {
	provider *MetabaseProvider
}

func (d *DatabaseResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database"
}

func (d *DatabaseResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.DatabaseResource()
}

func (d *DatabaseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DatabaseModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createRequest, diags := plan.toRequest()
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	databaseId, err := d.provider.client.CreateDatabase(*createRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating database",
			fmt.Sprintf("Unexpected error occurred: %s", err.Error()),
		)
		return
	}

	state, diags := d.fetchDatabaseState(ctx, databaseId, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (d *DatabaseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DatabaseModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	databaseId := state.Id.ValueInt64()
	database, err := d.provider.client.GetDatabase(databaseId)
	if err != nil {
		diags = utils.HandleResourceReadError(ctx, "database", databaseId, err, resp)
		resp.Diagnostics.Append(diags...)
		return
	}
	diags = mapDatabaseToState(ctx, database, &state)
	resp.Diagnostics.Append(diags...)
}

func (d *DatabaseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan DatabaseModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	databaseId := plan.Id.ValueInt64()
	updateRequest, diags := plan.toRequest()
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := d.provider.client.UpdateDatabase(databaseId, *updateRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error updating database with ID %d", databaseId),
			fmt.Sprintf("An unexpected error occurred: %s", err.Error()),
		)
		return
	}

	state, diags := d.fetchDatabaseState(ctx, databaseId, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (d *DatabaseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DatabaseModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	databaseId := state.Id.ValueInt64()
	err := d.provider.client.DeleteDatabase(databaseId)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error deleting database: %d", databaseId),
			fmt.Sprintf("Unexpected error occurred: %s", err.Error()),
		)
		return
	}
}

func (d *DatabaseModel) toRequest() (*client.DatabaseRequest, diag.Diagnostics) {
	engine := client.DatabaseEngine(d.Engine.ValueString())
	var diags diag.Diagnostics

	// Map the JSON-encoded details string into a map
	details, errDetails := unmarshallConfig(d.Details)
	detailsSecure, errDetailsSecure := unmarshallConfig(d.DetailsSecure)

	if errDetails != nil {
		diags.AddError(
			"Configuration error",
			fmt.Sprintf("Error processing details configuration: %s", errDetails.Error()),
		)
	}
	if errDetailsSecure != nil {
		diags.AddError(
			"Configuration error",
			fmt.Sprintf("Error processing details_secure configuration: %s", errDetails.Error()),
		)
	}

	detailsCombined := make(map[string]interface{})
	if details != nil {
		for k, v := range details {
			detailsCombined[k] = v
		}
	}
	if detailsSecure != nil {
		for k, v := range detailsSecure {
			detailsCombined[k] = v
		}
	}

	errs := checkDatabaseDetails(engine, detailsCombined)
	if len(errs) > 0 {
		diags := make([]diag.Diagnostic, len(errs))
		for i, err := range errs {
			diags[i] = diag.NewErrorDiagnostic(
				"Missing required database configuration",
				err.Error(),
			)
		}
		return nil, diags
	}

	// TODO: add schedules to the request
	return &client.DatabaseRequest{
		Engine:  engine,
		Name:    d.Name.ValueString(),
		Details: detailsCombined,
	}, nil
}

func (d *DatabaseResource) fetchDatabaseState(ctx context.Context, databaseId int64, plan DatabaseModel) (DatabaseModel, diag.Diagnostics) {
	database, err := d.provider.client.GetDatabase(databaseId)
	if err != nil {
		return DatabaseModel{}, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				fmt.Sprintf("Error fetching database with ID: %d", databaseId),
				fmt.Sprintf("An unexpected error occurred: %s", err.Error()),
			),
		}
	}

	var state DatabaseModel
	diags := mapDatabaseToState(ctx, database, &state)

	// Override both details and details_secure
	// The API can return additional keys in details, and anything in details_secure is redacted so the response is useless
	state.Details = plan.Details
	state.DetailsSecure = plan.DetailsSecure

	return state, diags
}

func mapDatabaseToState(ctx context.Context, database *client.Database, target *DatabaseModel) diag.Diagnostics {
	var diags diag.Diagnostics

	target.Id = types.Int64Value(database.Id)
	target.Engine = types.StringValue(database.Engine)
	target.Name = types.StringValue(database.Name)
	target.Features, _ = types.ListValueFrom(ctx, types.StringType, database.Features)

	details, detailsSecure, detailsDiags := splitDatabaseDetails(database)
	target.Details = details
	target.DetailsSecure = detailsSecure
	diags.Append(detailsDiags...)
	if diags.HasError() {
		return diags
	}

	schedules, scheduleDiags := buildSchedules(database)
	target.Schedules = schedules
	diags.Append(scheduleDiags...)
	return diags
}

// TODO: unit test this
func splitDatabaseDetails(database *client.Database) (types.String, types.String, diag.Diagnostics) {
	var diags diag.Diagnostics

	if database.Details == nil {
		return types.StringNull(), types.StringNull(), diags
	}

	details := make(map[string]interface{})
	detailsSecure := make(map[string]interface{})
	for k, v := range *database.Details {
		if isSensitiveDetail(k) {
			detailsSecure[k] = v
		} else {
			details[k] = v
		}
	}

	detailsStr, errDetails := json.Marshal(details)
	detailsSecureStr, errDetailsSecure := json.Marshal(detailsSecure)

	if errDetails != nil {
		diags.AddError(
			fmt.Sprintf("Error processing database: %d", database.Id),
			fmt.Sprintf("Error occurred when converting details to JSON string: %e", errDetails),
		)
	}
	if errDetailsSecure != nil {
		diags.AddError(
			fmt.Sprintf("Error processing database: %d", database.Id),
			fmt.Sprintf("Error occurred when converting details_secure to JSON string: %e", errDetailsSecure),
		)
	}
	if diags.HasError() {
		return types.StringNull(), types.StringNull(), diags
	}

	return types.StringValue(string(detailsStr)), types.StringValue(string(detailsSecureStr)), diag.Diagnostics{}

}

func isSensitiveDetail(key string) bool {
	return slices.Contains(databaseSensitiveProperties, key)
}

// TODO: unit test this
func buildSchedules(database *client.Database) (types.Map, diag.Diagnostics) {
	if database.Schedules == nil {
		return types.MapValue(schema.DatabaseScheduleType, map[string]attr.Value{})
	}

	schedules := make(map[string]attr.Value, len(*database.Schedules))
	for name, schedule := range *database.Schedules {
		schedules[name], _ = types.ObjectValue(schema.DatabaseScheduleType.AttributeTypes(), map[string]attr.Value{
			"day":    transforms.ToTerraformString(schedule.Day),
			"frame":  transforms.ToTerraformString(schedule.Frame),
			"hour":   transforms.ToTerraformInt(schedule.Hour),
			"minute": transforms.ToTerraformInt(schedule.Minute),
			"type":   types.StringValue(schedule.Type),
		})
	}

	return types.MapValue(schema.DatabaseScheduleType, schedules)
}

func checkDatabaseDetails(engine client.DatabaseEngine, details map[string]interface{}) []error {
	var errs []error

	var requireDetail = func(detail string, errIfMissing error) {
		if _, exists := details[detail]; !exists {
			errs = append(errs, errIfMissing)
		}
	}

	switch engine {
	case client.EngineH2:
		requireDetail("db", errMissingConnString)
	case client.EnginePostgres:
		requireDetail("dbname", errMissingDbName)
		requireDetail("host", errMissingHost)
		requireDetail("user", errMissingUser)
		requireDetail("password", errMissingPassword)
	}

	return errs
}

// TODO: move this to a utility?
func unmarshallConfig(config types.String) (map[string]interface{}, error) {
	if config.IsNull() {
		return nil, nil
	} else {
		configUnmarshalled := make(map[string]interface{})
		err := json.Unmarshal([]byte(config.ValueString()), &configUnmarshalled)
		if err != nil {
			return nil, fmt.Errorf("error processing database configuration: %e", err)
		}

		return configUnmarshalled, nil
	}
}
