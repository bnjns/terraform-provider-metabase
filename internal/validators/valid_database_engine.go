package validators

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/exp/slices"
	"terraform-provider-metabase/internal/client"
)

var _ tfsdk.AttributeValidator = isKnownDatabaseEngineValidator{}

type isKnownDatabaseEngineValidator struct{}

func (v isKnownDatabaseEngineValidator) Description(ctx context.Context) string {
	return "Checks whether the database engine specified is an accepted value."
}

func (v isKnownDatabaseEngineValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v isKnownDatabaseEngineValidator) Validate(ctx context.Context, request tfsdk.ValidateAttributeRequest, response *tfsdk.ValidateAttributeResponse) {
	var engine types.String
	diags := tfsdk.ValueAs(ctx, request.AttributeConfig, &engine)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	if engine.IsUnknown() || engine.IsNull() {
		return
	}

	dbEngine := client.DatabaseEngine(engine.ValueString())

	if !slices.Contains(client.DatabaseAllowedEngines, dbEngine) {
		response.Diagnostics.AddAttributeError(
			request.AttributePath,
			"Must be a valid database engine",
			fmt.Sprintf("Database engine '%s' is not a recognised type: %s", engine.ValueString(), client.DatabaseAllowedEngines),
		)
	}
}

func IsKnownDatabaseEngineValidator() tfsdk.AttributeValidator {
	return isKnownDatabaseEngineValidator{}
}
