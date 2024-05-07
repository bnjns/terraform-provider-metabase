package validators

import (
	"context"
	"github.com/bnjns/metabase-sdk-go/service/permissions"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUserNotInReservedGroupsValidator(t *testing.T) {
	t.Parallel()

	userNotInReservedGroupsValidator := UserNotInReservedGroupsValidator()
	ctx := context.Background()

	t.Run("description", func(t *testing.T) {
		assert.NotEmpty(t, userNotInReservedGroupsValidator.Description(ctx))
	})

	t.Run("markdown description", func(t *testing.T) {
		assert.NotEmpty(t, userNotInReservedGroupsValidator.MarkdownDescription(ctx))
	})

	t.Run("a null value should pass", func(t *testing.T) {
		request := validator.ListRequest{
			Path:        path.Empty(),
			ConfigValue: types.ListNull(types.Int64Type),
		}
		response := validator.ListResponse{}

		userNotInReservedGroupsValidator.ValidateList(ctx, request, &response)

		assert.Empty(t, response.Diagnostics)
	})

	t.Run("list of groups not including reserved should pass", func(t *testing.T) {
		tfGroupIds, _ := types.ListValueFrom(ctx, types.Int64Type, []int64{4, 5, 10, 8})

		request := validator.ListRequest{
			Path:        path.Empty(),
			ConfigValue: tfGroupIds,
		}
		response := validator.ListResponse{}

		userNotInReservedGroupsValidator.ValidateList(ctx, request, &response)

		assert.Empty(t, response.Diagnostics)
	})

	t.Run("a list of groups that contains the all users reserved group should return an error", func(t *testing.T) {
		tfGroupIds, _ := types.ListValueFrom(ctx, types.Int64Type, []int64{4, 5, permissions.GroupAllUsers, 8})

		request := validator.ListRequest{
			Path:        path.Empty(),
			ConfigValue: tfGroupIds,
		}
		response := validator.ListResponse{}

		userNotInReservedGroupsValidator.ValidateList(ctx, request, &response)

		assert.Equal(t, "Must not contain reserved group ID", response.Diagnostics[0].Summary())
	})

	t.Run("a list of groups that contains the administrators reserved group should return an error", func(t *testing.T) {
		tfGroupIds, _ := types.ListValueFrom(ctx, types.Int64Type, []int64{4, 5, permissions.GroupAdministrators, 8})

		request := validator.ListRequest{
			Path:        path.Empty(),
			ConfigValue: tfGroupIds,
		}
		response := validator.ListResponse{}

		userNotInReservedGroupsValidator.ValidateList(ctx, request, &response)

		assert.Equal(t, "Must not contain reserved group ID", response.Diagnostics[0].Summary())
	})
}
