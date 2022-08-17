package transforms

import "github.com/hashicorp/terraform-plugin-framework/types"

func FromTerraformBool(b types.Bool) *bool {
	if b.Null {
		return nil
	} else {
		return &b.Value
	}
}
