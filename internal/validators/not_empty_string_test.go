package validators

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNotEmptyStringValidator(t *testing.T) {
	t.Parallel()

	notEmptyStringValidator := NotEmptyStringValidator()
	ctx := context.Background()

	t.Run("description", func(t *testing.T) {
		assert.NotEmpty(t, notEmptyStringValidator.Description(ctx))
	})

	t.Run("markdown description", func(t *testing.T) {
		assert.NotEmpty(t, notEmptyStringValidator.MarkdownDescription(ctx))
	})

	t.Run("an empty string should return an error", func(t *testing.T) {
		request := validator.StringRequest{
			Path:        path.Empty(),
			ConfigValue: types.StringValue(""),
		}
		response := validator.StringResponse{}

		notEmptyStringValidator.ValidateString(ctx, request, &response)

		assert.Equal(t, "Must not be empty string", response.Diagnostics[0].Summary())
	})

	t.Run("a non-empty string should pass", func(t *testing.T) {
		request := validator.StringRequest{
			Path:        path.Empty(),
			ConfigValue: types.StringValue("non-empty value"),
		}
		response := validator.StringResponse{}

		notEmptyStringValidator.ValidateString(ctx, request, &response)

		assert.Empty(t, response.Diagnostics)
	})

	t.Run("a nil string should pass", func(t *testing.T) {
		request := validator.StringRequest{
			Path:        path.Empty(),
			ConfigValue: types.StringNull(),
		}
		response := validator.StringResponse{}

		notEmptyStringValidator.ValidateString(ctx, request, &response)

		assert.Empty(t, response.Diagnostics)
	})
}
