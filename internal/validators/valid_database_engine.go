package validators

import (
	"context"
	"fmt"
	"github.com/bnjns/metabase-sdk-go/service/database"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/exp/slices"
)

var databaseAllowedEngines = []database.Engine{
	database.EngineAmazonAthena,
	database.EngineAmazonRedshift,
	database.EngineBigQuery,
	database.EngineDruid,
	database.EngineGoogleAnalytics,
	database.EngineMongoDB,
	database.EngineMySQL,
	database.EnginePostgres,
	database.EnginePresto,
	database.EngineSnowflake,
	database.EngineSparkSQL,
	database.EngineSQLServer,
	database.EngineSQLite,
}

type isKnownDatabaseEngineValidator struct {
	validator.String
}

func IsKnownDatabaseEngineValidator() validator.String {
	return isKnownDatabaseEngineValidator{}
}

func (v isKnownDatabaseEngineValidator) Description(ctx context.Context) string {
	return "Checks whether the database engine specified is an accepted value."
}

func (v isKnownDatabaseEngineValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v isKnownDatabaseEngineValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	var engine types.String
	diags := tfsdk.ValueAs(ctx, request.ConfigValue, &engine)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	if engine.IsUnknown() || engine.IsNull() {
		return
	}

	dbEngine := database.Engine(engine.ValueString())

	if !slices.Contains(databaseAllowedEngines, dbEngine) {
		response.Diagnostics.AddAttributeWarning(
			request.Path,
			"Not a recognised database engine",
			fmt.Sprintf("Database engine '%s' is not a recognised type: %s. Applying is still possible, but the provider will not be able to validate the configuration.", engine.ValueString(), databaseAllowedEngines),
		)
	}
}
