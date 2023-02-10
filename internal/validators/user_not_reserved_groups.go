package validators

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/exp/slices"
	"strconv"
)

const (
	GroupIdAllUsers       = 1
	GroupIdAdministrators = 2
)

var ReservedGroupIds = []int64{GroupIdAllUsers, GroupIdAdministrators}

type userNotInReservedGroupsValidator struct {
	validator.String
}

func UserNotInReservedGroupsValidator() validator.List {
	return userNotInReservedGroupsValidator{}
}

func (u userNotInReservedGroupsValidator) Description(ctx context.Context) string {
	return "groups list must not contain a reserved group"
}

func (u userNotInReservedGroupsValidator) MarkdownDescription(ctx context.Context) string {
	return u.Description(ctx)
}

func (u userNotInReservedGroupsValidator) ValidateList(ctx context.Context, request validator.ListRequest, response *validator.ListResponse) {
	var groupIds types.List
	diags := tfsdk.ValueAs(ctx, request.ConfigValue, &groupIds)
	response.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	if groupIds.IsUnknown() || groupIds.IsNull() {
		return
	}

	for _, groupId := range groupIds.Elements() {
		gId, _ := strconv.ParseInt(groupId.String(), 10, 64)

		if slices.Contains(ReservedGroupIds, gId) {
			response.Diagnostics.AddAttributeError(
				request.Path,
				"Must not contain reserved group ID",
				fmt.Sprintf("Config contains reserved group ID %d which must not be explicitly set.", gId),
			)
		}
	}
}
