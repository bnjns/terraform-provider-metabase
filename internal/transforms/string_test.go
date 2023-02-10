package transforms

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestToTerraformString(t *testing.T) {
	t.Parallel()

	t.Run("nil", func(t *testing.T) {
		tfStr := ToTerraformString(nil)

		assert.True(t, tfStr.IsNull())
		assert.Empty(t, tfStr.ValueString())
	})

	t.Run("non-nil", func(t *testing.T) {
		str := "non-nil"
		tfStr := ToTerraformString(&str)

		assert.False(t, tfStr.IsNull())
		assert.Equal(t, "non-nil", tfStr.ValueString())
	})
}

func TestFromTerraformString(t *testing.T) {
	t.Parallel()

	t.Run("nil", func(t *testing.T) {
		str := FromTerraformString(types.StringNull())

		assert.Nil(t, str)
	})

	t.Run("non-nil", func(t *testing.T) {
		str := FromTerraformString(types.StringValue("non-nil"))

		assert.Equal(t, "non-nil", *str)
	})
}
