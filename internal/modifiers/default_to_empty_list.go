package modifiers

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type defaultToEmptyListModifier struct {
	elemType attr.Type
}

func DefaultToEmptyListModifier(elemType attr.Type) tfsdk.AttributePlanModifier {
	return defaultToEmptyListModifier{
		elemType: elemType,
	}
}

func (r defaultToEmptyListModifier) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
	var config types.List
	diags := tfsdk.ValueAs(ctx, req.AttributeConfig, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.Null {
		return
	}

	resp.AttributePlan = types.List{
		ElemType: r.elemType,
		Elems:    []attr.Value{},
		Null:     false,
		Unknown:  false,
	}
}

func (r defaultToEmptyListModifier) Description(ctx context.Context) string {
	return "Defaults a null or unknown value to an empty list of the specified type."
}

func (r defaultToEmptyListModifier) MarkdownDescription(ctx context.Context) string {
	return r.Description(ctx)
}
