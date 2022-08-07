package utils

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ToTerraformString(str *string) types.String {
	if str == nil {
		return types.String{
			Null:  true,
			Value: "",
		}
	} else {
		return types.String{
			Value: *str,
		}
	}
}

func FromTerraformString(str types.String) *string {
	if str.Null {
		return nil
	} else {
		return &str.Value
	}
}
