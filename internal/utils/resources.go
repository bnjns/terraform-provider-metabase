package utils

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"strings"
)

func HandleResourceReadError(ctx context.Context, resourceName string, resourceId int64, err error, response *resource.ReadResponse) diag.Diagnostics {
	if strings.Contains(err.Error(), "not found") {
		response.State.RemoveResource(ctx)
		return diag.Diagnostics{}
	} else {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				fmt.Sprintf("Failed to get %s with ID %d", resourceName, resourceId),
				fmt.Sprintf("An error occurred: %s", err.Error()),
			),
		}
	}
}
