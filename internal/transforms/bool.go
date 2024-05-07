package transforms

import "github.com/hashicorp/terraform-plugin-framework/types"

func FromTerraformBool(b types.Bool) *bool {
	if b.IsNull() {
		return nil
	} else {
		val := b.ValueBool()
		return &val
	}
}

func ToTerraformBool(b *bool) types.Bool {
	if b == nil {
		return types.BoolNull()
	} else {
		return types.BoolValue(*b)
	}
}
