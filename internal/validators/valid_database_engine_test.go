package validators

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"testing"
)

func TestIsKnownDatabaseEngineValidator(t *testing.T) {
	t.Parallel()

	dbEngineValidator := IsKnownDatabaseEngineValidator()
	ctx := context.Background()

	t.Run("description", func(t *testing.T) {
		assert.NotEmpty(t, dbEngineValidator.Description(ctx))
	})

	t.Run("markdown description", func(t *testing.T) {
		assert.NotEmpty(t, dbEngineValidator.MarkdownDescription(ctx))
	})

	t.Run("an invalid engine should add an error", func(t *testing.T) {
		request := validator.StringRequest{
			Path:        path.Empty(),
			ConfigValue: types.StringValue("invalid"),
		}
		response := validator.StringResponse{}

		dbEngineValidator.ValidateString(ctx, request, &response)

		assert.NotEmpty(t, response.Diagnostics)
		assert.Equal(t, "Not a recognised database engine", response.Diagnostics[0].Summary())
		assert.Equal(t, diag.SeverityWarning, response.Diagnostics[0].Severity())
	})

	t.Run("a valid engine should pass", func(t *testing.T) {
		request := validator.StringRequest{
			Path:        path.Empty(),
			ConfigValue: types.StringValue("postgres"),
		}
		response := validator.StringResponse{}

		dbEngineValidator.ValidateString(ctx, request, &response)

		assert.Empty(t, response.Diagnostics)
	})
}
