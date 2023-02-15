package transforms

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFromTerraformInt64List(t *testing.T) {
	t.Parallel()

	t.Run("nil", func(t *testing.T) {
		intList := FromTerraformInt64List(types.ListNull(types.Int64Type))

		assert.Nil(t, intList)
	})

	t.Run("non-nil", func(t *testing.T) {
		tfIntList, _ := types.ListValue(
			types.Int64Type,
			[]attr.Value{
				types.Int64Value(1),
				types.Int64Value(5),
				types.Int64Value(9),
			},
		)

		intList := FromTerraformInt64List(tfIntList)

		assert.Equal(t, []int64{1, 5, 9}, *intList)
	})
}

func TestToTerraformInt(t *testing.T) {
	t.Parallel()

	t.Run("nil", func(t *testing.T) {
		tfInt := ToTerraformInt(nil)

		assert.True(t, tfInt.IsNull())
		assert.Zero(t, tfInt.ValueInt64())
	})

	t.Run("non-nil", func(t *testing.T) {
		num := int64(12)
		tfInt := ToTerraformInt(&num)

		assert.False(t, tfInt.IsNull())
		assert.Equal(t, int64(12), tfInt.ValueInt64())
	})
}

func TestFromTerraformInt(t *testing.T) {
	t.Parallel()

	t.Run("nil", func(t *testing.T) {
		num := FromTerraformInt(types.Int64Null())

		assert.Nil(t, num)
	})

	t.Run("non-nil", func(t *testing.T) {
		num := FromTerraformInt(types.Int64Value(12))

		assert.Equal(t, int64(12), *num)
	})
}
