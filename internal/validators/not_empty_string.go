package validators

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type notEmptyStringValidator struct {
	validator.String
}

func NotEmptyStringValidator() validator.String {
	return notEmptyStringValidator{}
}

func (v notEmptyStringValidator) Description(ctx context.Context) string {
	return "string must be nil or empty"
}

func (v notEmptyStringValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	var str types.String
	diags := tfsdk.ValueAs(ctx, req.ConfigValue, &str)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	if str.IsUnknown() || str.IsNull() {
		return
	}

	if len(str.ValueString()) == 0 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Must not be empty string",
			"You must provide a non-empty string.",
		)
	}
}
