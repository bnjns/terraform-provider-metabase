package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pkg/errors"
	"strconv"
	"terraform-provider-metabase/internal/client"
	"terraform-provider-metabase/internal/transforms"
	"terraform-provider-metabase/internal/validators"
)

var (
	errMissingConnString     = errors.New("you must provide the connection string in the 'db' property")
	errMissingDbName         = errors.New("you must provide the database name in the 'db' property")
	errMissingHost           = errors.New("you must provide the database hostname/ip in the 'host' property")
	errMissingPort           = errors.New("you must provide the database port in the 'port' property")
	errMissingUsername       = errors.New("you must provide the auth username in the 'username' property")
	errMissingPassword       = errors.New("you must provide the auth password in the 'password' property")
	errMissingGcpCredentials = errors.New("you must provide the service account credentials as a stringified JSON object in the 'service-account-json' property")
)

type DatabaseModel struct {
	Id          types.Int64  `tfsdk:"id"`
	Engine      types.String `tfsdk:"engine"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`

	Features  types.List `tfsdk:"features"`
	Details   types.Map  `tfsdk:"details"`
	Schedules types.Map  `tfsdk:"schedules"`
}

type blockTypeDatabase int

const (
	blockTypeDatabaseResource blockTypeDatabase = iota
	blockTypeDatabaseDataSource
)

// Ensure provider fully satisfies the framework interfaces
var _ resource.Resource = &DatabaseResource{}
var _ resource.ResourceWithImportState = &DatabaseResource{}

type DatabaseResource struct {
	provider *MetabaseProvider
}

var typeDatabaseSchedule = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"day":    types.StringType,
		"frame":  types.StringType,
		"hour":   types.Int64Type,
		"minute": types.Int64Type,
		"type":   types.StringType,
	},
}

func (d *DatabaseResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_database"
}

func (d *DatabaseResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Allows you to manage a database configuration",
		Attributes:  getDatabaseAttributes(blockTypeDatabaseResource),
	}, nil
}

func (d *DatabaseResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan DatabaseModel
	diags := request.Plan.Get(ctx, &plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	createRequest, diags := plan.toRequest()
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	databaseId, err := d.provider.client.CreateDatabase(*createRequest)
	if err != nil {
		response.Diagnostics.AddError(
			"Error creating database",
			fmt.Sprintf("Unexpected error occurred: %s", err.Error()),
		)
		return
	}

	// Refresh the state
	var database DatabaseModel
	database.Id = types.Int64Value(databaseId)
	diags = d.provider.syncDatabaseWithApi(&database)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// Update the state
	diags = response.State.Set(ctx, database)
	response.Diagnostics.Append(diags...)
}

func (d *DatabaseResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state DatabaseModel
	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	diags = d.provider.syncDatabaseWithApi(&state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	diags = response.State.Set(ctx, &state)
	response.Diagnostics.Append(diags...)
}

func (d *DatabaseResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan DatabaseModel
	diags := request.Plan.Get(ctx, &plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	var state DatabaseModel
	diags = request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	databaseId := state.Id.ValueInt64()
	updateRequest, diags := plan.toRequest()
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	err := d.provider.client.UpdateDatabase(databaseId, *updateRequest)
	if err != nil {
		response.Diagnostics.AddError(
			fmt.Sprintf("Error updating database with ID %d", databaseId),
			fmt.Sprintf("Unexpected error occurred: %s", err.Error()),
		)
		return
	}

	// Refresh the state
	var database DatabaseModel
	database.Id = types.Int64Value(databaseId)
	diags = d.provider.syncDatabaseWithApi(&database)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// Update the state
	diags = response.State.Set(ctx, database)
	response.Diagnostics.Append(diags...)
}

func (d *DatabaseResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var database DatabaseModel
	diags := request.State.Get(ctx, &database)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	databaseId := database.Id.ValueInt64()
	err := d.provider.client.DeleteDatabase(databaseId)
	if err != nil {
		response.Diagnostics.AddError(
			"Error deleting database",
			fmt.Sprintf("Unexpected error occurred: %s", err.Error()),
		)
		return
	}
}

func (d *DatabaseResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	databaseId, _ := strconv.ParseInt(request.ID, 10, 64)

	// Refresh the state from the API
	var state DatabaseModel
	state.Id = types.Int64Value(databaseId)
	diags := d.provider.syncDatabaseWithApi(&state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// Set the state
	diags = response.State.Set(ctx, &state)
	response.Diagnostics.Append(diags...)
}

func (d DatabaseModel) toRequest() (*client.DatabaseRequest, diag.Diagnostics) {
	engine := client.DatabaseEngine(d.Engine.ValueString())
	errs := checkDatabaseDetails(engine, d.Details)
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

	details := make(map[string]*string, len(d.Details.Elements()))
	for name, detail := range d.Details.Elements() {
		details[name] = transforms.FromTerraformString(detail.(types.String))
	}
	return &client.DatabaseRequest{
		Engine:      engine,
		Name:        d.Name.ValueString(),
		Description: transforms.FromTerraformString(d.Description),
		Details:     details,
	}, nil
}

func getDatabaseAttributes(t blockTypeDatabase) map[string]tfsdk.Attribute {
	return map[string]tfsdk.Attribute{
		"id": {
			Type:        types.Int64Type,
			Description: "The ID of the database.",
			Required:    t == blockTypeDatabaseDataSource,
			Computed:    t == blockTypeDatabaseResource,
		},
		"engine": {
			Type:        types.StringType,
			Description: "The engine type of the database.",
			Required:    t == blockTypeDatabaseResource,
			Computed:    t == blockTypeDatabaseDataSource,
			Validators: []tfsdk.AttributeValidator{
				validators.IsKnownDatabaseEngineValidator(),
			},
		},
		"name": {
			Type:        types.StringType,
			Description: "The name of the database.",
			Required:    t == blockTypeDatabaseResource,
			Computed:    t == blockTypeDatabaseDataSource,
		},
		"description": {
			Type:        types.StringType,
			Description: "An optional description of the database.",
			Optional:    t == blockTypeDatabaseResource,
			Computed:    true,
		},
		"features": {
			Type: types.ListType{
				ElemType: types.StringType,
			},
			Description: "The features this database engine supports.",
			Computed:    true,
		},
		"details": {
			Type: types.MapType{
				ElemType: types.StringType,
			},
			Description: "",
			Optional:    t == blockTypeDatabaseResource,
			Computed:    true,
			Sensitive:   true,
		},
		"schedules": {
			Type: types.MapType{
				ElemType: typeDatabaseSchedule,
			},
			Description: "The schedules used to sync the database.",
			Optional:    t == blockTypeDatabaseResource,
			Computed:    true,
		},
	}
}

func checkDatabaseDetails(engine client.DatabaseEngine, details types.Map) []error {
	var errs []error

	var requireDetail = func(detail string, errIfMissing error) {
		if _, exists := details.Elements()[detail]; !exists {
			errs = append(errs, errIfMissing)
		}
	}

	switch engine {
	case client.EngineAmazonRedshift, client.EngineMongoDB, client.EngineMySQL, client.EnginePostgres:
		requireDetail("db", errMissingDbName)
		requireDetail("host", errMissingHost)
		requireDetail("port", errMissingPort)
		requireDetail("username", errMissingUsername)
		requireDetail("password", errMissingPassword)
	case client.EngineBigQuery:
		requireDetail("service-account-json", errMissingGcpCredentials)
	case client.EngineDruid:
		requireDetail("host", errMissingHost)
		requireDetail("port", errMissingPort)
	case client.EngineGoogleAnalytics:
		requireDetail("service-account-json", errMissingGcpCredentials)
	case client.EngineH2:
		requireDetail("db", errMissingConnString)
	case client.EngineOracle:
		// TODO
	case client.EnginePresto:
		// TODO
	case client.EnginePrestoDeprecated:
		// TODO
	case client.EngineSnowflake:
		// TODO
	case client.EngineSparkSQL:
		// TODO
	case client.EngineSQLServer:
		// TODO
	case client.EngineSQLite:
		// TODO
	}

	return errs
}

func (p *MetabaseProvider) syncDatabaseWithApi(state *DatabaseModel) diag.Diagnostics {
	databaseId := state.Id.ValueInt64()

	database, err := p.client.GetDatabase(databaseId)
	if err != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				fmt.Sprintf("Failed to get database with ID %d", databaseId),
				fmt.Sprintf("An error occurred: %s", err.Error()),
			),
		}
	}

	mapDatabaseToState(database, state)
	return diag.Diagnostics{}
}

func mapDatabaseToState(database *client.Database, target *DatabaseModel) {
	target.Id = types.Int64Value(database.Id)
	target.Engine = types.StringValue(database.Engine)
	target.Name = types.StringValue(database.Name)
	target.Description = transforms.ToTerraformString(database.Description)

	// Set the feature list
	features := make([]attr.Value, len(database.Features))
	for i, feature := range database.Features {
		features[i] = types.StringValue(feature)
	}
	target.Features, _ = types.ListValue(types.StringType, features)

	// Set the details map
	if database.Details != nil {
		details := make(map[string]attr.Value, len(*database.Details))
		for name, detail := range *database.Details {
			details[name] = transforms.ToTerraformString(detail)
		}
		target.Details, _ = types.MapValue(types.StringType, details)
	} else {
		target.Details, _ = types.MapValue(types.StringType, map[string]attr.Value{})
	}

	// Set the schedule map
	if database.Schedules != nil {
		schedules := make(map[string]attr.Value, len(*database.Schedules))
		for name, schedule := range *database.Schedules {
			schedules[name], _ = types.ObjectValue(typeDatabaseSchedule.AttributeTypes(), map[string]attr.Value{
				"day":    transforms.ToTerraformString(schedule.Day),
				"frame":  transforms.ToTerraformString(schedule.Frame),
				"hour":   transforms.ToTerraformInt(schedule.Hour),
				"minute": transforms.ToTerraformInt(schedule.Minute),
				"type":   types.StringValue(schedule.Type),
			})
		}
		target.Schedules, _ = types.MapValue(typeDatabaseSchedule, schedules)
	} else {
		target.Schedules, _ = types.MapValue(typeDatabaseSchedule, map[string]attr.Value{})
	}
}
