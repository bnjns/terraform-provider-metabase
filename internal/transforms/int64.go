package transforms

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strconv"
)

func FromTerraformInt64List(l types.List) *[]int64 {
	if l.IsNull() {
		return nil
	} else {
		var newList []int64
		for _, item := range l.Elements() {
			newItem, _ := strconv.ParseInt(item.String(), 10, 64)
			newList = append(newList, newItem)
		}

		return &newList
	}
}

func ToTerraformInt(i *int64) types.Int64 {
	if i == nil {
		return types.Int64{
			Null:  true,
			Value: 0,
		}
	} else {
		return types.Int64{
			Value: *i,
		}
	}
}

func FromTerraformInt(i types.Int64) *int64 {
	if i.Null {
		return nil
	} else {
		return &i.Value
	}
}
