package modifiers

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type defaultToEmptyListModifier struct {
	planmodifier.List
	elemType attr.Type
}

func DefaultToEmptyListModifier(elemType attr.Type) planmodifier.List {
	return defaultToEmptyListModifier{
		elemType: elemType,
	}
}

func (r defaultToEmptyListModifier) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	var config types.List
	diags := tfsdk.ValueAs(ctx, req.ConfigValue, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.IsNull() {
		return
	}

	resp.PlanValue, diags = types.ListValue(r.elemType, []attr.Value{})
	resp.Diagnostics.Append(diags...)
}

func (r defaultToEmptyListModifier) Description(ctx context.Context) string {
	return "Defaults a null or unknown value to an empty list of the specified type."
}

func (r defaultToEmptyListModifier) MarkdownDescription(ctx context.Context) string {
	return r.Description(ctx)
}
