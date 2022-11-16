package transforms

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ToTerraformString(str *string) types.String {
	if str == nil {
		return types.StringNull()
	} else {
		return types.StringValue(*str)
	}
}

func FromTerraformString(str types.String) *string {
	if str.IsNull() {
		return nil
	} else {
		val := str.ValueString()
		return &val
	}
}
