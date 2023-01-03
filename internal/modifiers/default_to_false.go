package modifiers

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type defaultToFalseModifier struct {
	planmodifier.Bool
}

func DefaultToFalseModifier() planmodifier.Bool {
	return defaultToFalseModifier{}
}

func (r defaultToFalseModifier) PlanModifyBool(ctx context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	var plan types.Bool
	diags := tfsdk.ValueAs(ctx, req.PlanValue, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.IsNull() {
		return
	}

	resp.PlanValue = types.BoolValue(false)
}

func (r defaultToFalseModifier) Description(ctx context.Context) string {
	return "Defaults a null or unknown value to false."
}

func (r defaultToFalseModifier) MarkdownDescription(ctx context.Context) string {
	return r.Description(ctx)
}
