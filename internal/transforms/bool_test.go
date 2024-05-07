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

func TestToTerraformBool(t *testing.T) {
	t.Parallel()

	t.Run("nil", func(t *testing.T) {
		tfBool := ToTerraformBool(nil)

		assert.True(t, tfBool.IsNull())
	})

	t.Run("non-nil", func(t *testing.T) {
		b := true
		tfBool := ToTerraformBool(&b)

		assert.False(t, tfBool.IsNull())
		assert.True(t, tfBool.ValueBool())
	})
}
