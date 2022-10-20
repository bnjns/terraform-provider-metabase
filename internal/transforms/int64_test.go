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
		intList := FromTerraformInt64List(types.List{
			ElemType: types.Int64Type,
			Null:     true,
		})

		assert.Nil(t, intList)
	})

	t.Run("non-nil", func(t *testing.T) {
		intList := FromTerraformInt64List(types.List{
			ElemType: types.Int64Type,
			Elems: []attr.Value{
				types.Int64{Value: 1},
				types.Int64{Value: 5},
				types.Int64{Value: 9},
			},
		})

		assert.Equal(t, []int64{1, 5, 9}, *intList)
	})
}

func TestToTerraformInt(t *testing.T) {
	t.Parallel()

	t.Run("nil", func(t *testing.T) {
		tfInt := ToTerraformInt(nil)

		assert.True(t, tfInt.Null)
		assert.Zero(t, tfInt.Value)
	})

	t.Run("non-nil", func(t *testing.T) {
		num := int64(12)
		tfInt := ToTerraformInt(&num)

		assert.False(t, tfInt.Null)
		assert.Equal(t, int64(12), tfInt.Value)
	})
}

func TestFromTerraformInt(t *testing.T) {
	t.Parallel()

	t.Run("nil", func(t *testing.T) {
		num := FromTerraformInt(types.Int64{
			Null: true,
		})

		assert.Nil(t, num)
	})

	t.Run("non-nil", func(t *testing.T) {
		num := FromTerraformInt(types.Int64{
			Value: int64(12),
		})

		assert.Equal(t, int64(12), *num)
	})
}
