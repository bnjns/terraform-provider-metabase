package transforms

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strconv"
)

func FromTerraformInt64List(l types.List) *[]int64 {
	if l.Null {
		return nil
	} else {
		var newList []int64
		for _, item := range l.Elems {
			newItem, _ := strconv.ParseInt(item.String(), 10, 64)
			newList = append(newList, newItem)
		}

		return &newList
	}
}
