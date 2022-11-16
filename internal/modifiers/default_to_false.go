package modifiers

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type defaultToFalseModifier struct{}

func DefaultToFalseModifier() tfsdk.AttributePlanModifier {
	return defaultToFalseModifier{}
}

func (r defaultToFalseModifier) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
	var plan types.Bool
	diags := tfsdk.ValueAs(ctx, req.AttributePlan, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.IsNull() {
		return
	}

	resp.AttributePlan = types.BoolValue(false)
}

func (r defaultToFalseModifier) Description(ctx context.Context) string {
	return "Defaults a null or unknown value to false."
}

func (r defaultToFalseModifier) MarkdownDescription(ctx context.Context) string {
	return r.Description(ctx)
}
