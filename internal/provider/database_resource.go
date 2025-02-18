package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bnjns/metabase-sdk-go/service/database"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"golang.org/x/exp/slices"
	"regexp"
	"strconv"
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
	Schedules     types.Object `tfsdk:"schedules"`
}

// Ensure provider fully satisfies the framework interfaces
var _ resource.Resource = &DatabaseResource{}
var _ resource.ResourceWithImportState = &DatabaseResource{}

var (
	errMissingConnString = errors.New("you must provide the connection string in the 'db' property")
	errMissingDbName     = errors.New("you must provide the database name in the 'dbname' property")
	errMissingHost       = errors.New("you must provide the database hostname/ip in the 'host' property")
	errMissingPort       = errors.New("you must provide the database port in the 'port' property")
	errMissingUser       = errors.New("you must provide the auth username in the 'user' property")
	errMissingPassword   = errors.New("you must provide the auth password in the 'password' property")
)
var sensitiveDatabaseDetails = []string{"password", "service-account-json"}
var redactedPattern = regexp.MustCompile(`^\*\*.+\*\*$`)

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

	databaseDetails, diags := plan.buildDatabaseDetails()
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	databaseId, err := d.provider.client.Database.Create(ctx, &database.CreateRequest{
		Name:    plan.Name.ValueString(),
		Engine:  database.Engine(plan.Engine.ValueString()),
		Details: databaseDetails,
	})
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
	db, err := d.provider.client.Database.Get(ctx, databaseId)
	if err != nil {
		diags = utils.HandleResourceReadError(ctx, "database", databaseId, err, resp)
		resp.Diagnostics.Append(diags...)
		return
	}
	diags = mapDatabaseToState(ctx, db, &state)
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
	databaseDetails, diags := plan.buildDatabaseDetails()
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	db, err := d.provider.client.Database.Get(ctx, databaseId)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error updating database with ID %d", databaseId),
			fmt.Sprintf("An error occurred when fetching the database: %s", err.Error()),
		)
		return
	}

	err = d.provider.client.Database.Update(ctx, databaseId, &database.UpdateRequest{
		Name:             plan.Name.ValueStringPointer(),
		Engine:           &db.Engine,
		Refingerprint:    &db.Refingerprint,
		Details:          &databaseDetails,
		Schedules:        db.Schedules,
		Caveats:          db.Caveats,
		PointsOfInterest: db.PointsOfInterest,
		AutoRunQueries:   &db.AutoRunQueries,
		CacheTTL:         db.CacheTTL,
		Settings:         db.Settings,
	})
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
	err := d.provider.client.Database.Delete(ctx, databaseId)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error deleting database: %d", databaseId),
			fmt.Sprintf("Unexpected error occurred: %s", err.Error()),
		)
		return
	}
}

func (d *DatabaseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	databaseId, _ := strconv.ParseInt(req.ID, 10, 64)

	plan := DatabaseModel{
		Details:       types.StringUnknown(),
		DetailsSecure: types.StringUnknown(),
	}
	state, diags := d.fetchDatabaseState(ctx, databaseId, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (d *DatabaseModel) buildDatabaseDetails() (database.Details, diag.Diagnostics) {
	engine := database.Engine(d.Engine.ValueString())
	var diags diag.Diagnostics

	// Map the JSON-encoded details string into a map
	details, errDetails := utils.UnmarshallJson(d.Details)
	detailsSecure, errDetailsSecure := utils.UnmarshallJson(d.DetailsSecure)

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

	return detailsCombined, nil
}

func (d *DatabaseResource) fetchDatabaseState(ctx context.Context, databaseId int64, plan DatabaseModel) (DatabaseModel, diag.Diagnostics) {
	db, err := d.provider.client.Database.Get(ctx, databaseId)
	if err != nil {
		return DatabaseModel{}, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				fmt.Sprintf("Error fetching database with ID: %d", databaseId),
				fmt.Sprintf("An unexpected error occurred: %s", err.Error()),
			),
		}
	}

	var state DatabaseModel
	diags := mapDatabaseToState(ctx, db, &state)

	// Override both details and details_secure if they're set in the plan as the API can return additional keys in
	// details and this produces an inconsistent result
	if !plan.Details.IsUnknown() {
		state.Details = plan.Details
	}
	if !plan.DetailsSecure.IsUnknown() {
		state.DetailsSecure = plan.DetailsSecure
	}

	return state, diags
}

func mapDatabaseToState(ctx context.Context, db *database.Database, target *DatabaseModel) diag.Diagnostics {
	var diags diag.Diagnostics

	target.Id = types.Int64Value(db.Id)
	target.Engine = types.StringValue(string(db.Engine))
	target.Name = types.StringValue(db.Name)
	target.Features, _ = types.ListValueFrom(ctx, types.StringType, db.Features)

	schedules, scheduleDiags := buildSchedules(db)
	target.Schedules = schedules
	diags.Append(scheduleDiags...)

	details, detailsSecure, detailsDiags := buildDatabaseDetails(db)
	target.Details = details
	target.DetailsSecure = detailsSecure
	diags.Append(detailsDiags...)

	return diags
}

func isSensitiveDatabaseDetail(key string, value interface{}) bool {
	if slices.Contains(sensitiveDatabaseDetails, key) {
		return true
	}

	valueStr, isString := value.(string)
	if !isString {
		return false
	}

	if redactedPattern.MatchString(valueStr) {
		return true
	}

	return false
}

func buildDatabaseDetails(db *database.Database) (types.String, types.String, diag.Diagnostics) {
	var diags diag.Diagnostics

	if db.Details == nil {
		return types.StringNull(), types.StringNull(), diags
	}

	details := make(map[string]any)
	detailsSecure := make(map[string]any)
	for k, v := range *db.Details {
		if isSensitiveDatabaseDetail(k, v) {
			detailsSecure[k] = v
		} else {
			details[k] = v
		}
	}

	detailsStr, err := json.Marshal(details)
	if err != nil {
		diags = append(diags, diag.NewErrorDiagnostic(fmt.Sprintf("Error parsing details for database %d", db.Id), err.Error()))
	}

	detailsSecureStr, err := json.Marshal(detailsSecure)
	if err != nil {
		diags = append(diags, diag.NewErrorDiagnostic(fmt.Sprintf("Error parsing details_secure for database %d", db.Id), err.Error()))
	}

	return types.StringValue(string(detailsStr)), types.StringValue(string(detailsSecureStr)), diags
}

func buildScheduleSettings(settings *database.ScheduleSettings) basetypes.ObjectValue {
	if settings == nil {
		return types.ObjectNull(schema.DatabaseScheduleType.AttributeTypes())
	}

	scheduleSettings, _ := types.ObjectValue(schema.DatabaseScheduleType.AttributeTypes(), map[string]attr.Value{
		"type":   types.StringValue(string(settings.Type)),
		"day":    transforms.ToTerraformString((*string)(settings.Day)),
		"frame":  transforms.ToTerraformString((*string)(settings.Frame)),
		"hour":   transforms.ToTerraformInt(settings.Hour),
		"minute": transforms.ToTerraformInt(settings.Minute),
	})
	return scheduleSettings
}

func buildSchedules(db *database.Database) (types.Object, diag.Diagnostics) {
	if db.Schedules == nil {
		return types.ObjectNull(schema.DatabaseSchedulesType.AttributeTypes()), nil
	}

	schedules := map[string]attr.Value{
		"metadata_sync":      buildScheduleSettings(db.Schedules.MetadataSync),
		"cache_field_values": buildScheduleSettings(db.Schedules.CacheFieldValues),
	}

	return types.ObjectValue(schema.DatabaseSchedulesType.AttributeTypes(), schedules)
}

func checkDatabaseDetails(engine database.Engine, details map[string]interface{}) []error {
	var errs []error

	var requireDetail = func(detail string, errIfMissing error) {
		if _, exists := details[detail]; !exists {
			errs = append(errs, errIfMissing)
		}
	}

	switch engine {
	case database.EnginePostgres:
		requireDetail("dbname", errMissingDbName)
		requireDetail("host", errMissingHost)
		requireDetail("user", errMissingUser)
		requireDetail("password", errMissingPassword)
	}

	return errs
}
