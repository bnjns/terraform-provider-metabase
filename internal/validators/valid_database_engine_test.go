package validators

import (
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"testing"
)

func TestIsKnownDatabaseEngineValidator(t *testing.T) {
	t.Parallel()

	validator := IsKnownDatabaseEngineValidator()
	ctx := context.Background()

	t.Run("description", func(t *testing.T) {
		assert.NotEmpty(t, validator.Description(ctx))
	})

	t.Run("markdown description", func(t *testing.T) {
		assert.NotEmpty(t, validator.MarkdownDescription(ctx))
	})

	t.Run("an invalid engine should add an error", func(t *testing.T) {
		request := tfsdk.ValidateAttributeRequest{
			AttributePath:   path.Empty(),
			AttributeConfig: types.String{Value: "invalid"},
		}
		response := tfsdk.ValidateAttributeResponse{}

		validator.Validate(ctx, request, &response)

		assert.NotEmpty(t, response.Diagnostics)
		assert.Equal(t, "Must be a valid database engine", response.Diagnostics[0].Summary())
	})

	t.Run("a valid engine should pass", func(t *testing.T) {
		request := tfsdk.ValidateAttributeRequest{
			AttributePath:   path.Empty(),
			AttributeConfig: types.String{Value: "h2"},
		}
		response := tfsdk.ValidateAttributeResponse{}

		validator.Validate(ctx, request, &response)

		assert.Empty(t, response.Diagnostics)
	})
}
