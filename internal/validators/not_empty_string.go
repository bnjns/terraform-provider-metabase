package validators

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type notEmptyStringValidator struct {
	tfsdk.AttributeValidator
}

func NotEmptyStringValidator() tfsdk.AttributeValidator {
	return notEmptyStringValidator{}
}

func (v notEmptyStringValidator) Description(ctx context.Context) string {
	return "string must be nil or empty"
}

func (v notEmptyStringValidator) Validate(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
	var str types.String
	diags := tfsdk.ValueAs(ctx, req.AttributeConfig, &str)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	if str.Unknown || str.Null {
		return
	}

	if len(str.Value) == 0 {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Must not be empty string",
			"You must provide a non-empty string.",
		)
	}
}
