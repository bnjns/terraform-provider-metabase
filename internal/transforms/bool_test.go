package transforms

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFromTerraformBool(t *testing.T) {
	t.Parallel()

	t.Run("nil", func(t *testing.T) {
		result := FromTerraformBool(types.BoolNull())

		assert.Nil(t, result)
	})

	t.Run("non-nil", func(t *testing.T) {
		result := FromTerraformBool(types.BoolValue(true))

		assert.True(t, *result)
	})
}
