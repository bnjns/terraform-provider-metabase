package schema

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/stretchr/testify/assert"
	"terraform-provider-metabase/internal/validators"
	"testing"
)

func TestDatabaseResource(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("database schema should return expected fields", func(t *testing.T) {
		resourceSchema := DatabaseResource()

		assert.NotEmpty(t, resourceSchema.Description)
		assert.Equal(t, 7, len(resourceSchema.Attributes))

		t.Run("id should be configured", func(t *testing.T) {
			assert.IsType(t, schema.Int64Attribute{}, resourceSchema.Attributes["id"])

			id := resourceSchema.Attributes["id"].(schema.Int64Attribute)
			assert.NotEmpty(t, id.Description)
			assert.True(t, id.IsComputed())
			assert.False(t, id.IsRequired())
			assert.False(t, id.IsOptional())
			assert.Contains(t, id.Int64PlanModifiers(), int64planmodifier.UseStateForUnknown())
		})

		t.Run("engine should be configured", func(t *testing.T) {
			assert.IsType(t, schema.StringAttribute{}, resourceSchema.Attributes["engine"])

			engine := resourceSchema.Attributes["engine"].(schema.StringAttribute)
			assert.NotEmpty(t, engine.Description)
			assert.True(t, engine.IsRequired())
			assert.Contains(t, engine.StringValidators(), validators.IsKnownDatabaseEngineValidator())
			assert.Equal(t, "If the value of this attribute changes, Terraform will destroy and recreate the resource.", engine.StringPlanModifiers()[0].Description(ctx))
		})

		t.Run("name should be configured", func(t *testing.T) {
			assert.IsType(t, schema.StringAttribute{}, resourceSchema.Attributes["name"])

			name := resourceSchema.Attributes["name"].(schema.StringAttribute)
			assert.NotEmpty(t, name.Description)
			assert.True(t, name.IsRequired())
		})

		t.Run("features should be configured", func(t *testing.T) {
			assert.IsType(t, schema.ListAttribute{}, resourceSchema.Attributes["features"])

			features := resourceSchema.Attributes["features"].(schema.ListAttribute)
			assert.NotEmpty(t, features.Description)
			assert.True(t, features.IsComputed())
			assert.False(t, features.IsRequired())
			assert.False(t, features.IsOptional())
			assert.Contains(t, features.ListPlanModifiers(), listplanmodifier.UseStateForUnknown())
		})

		t.Run("details should be configured", func(t *testing.T) {
			assert.IsType(t, schema.StringAttribute{}, resourceSchema.Attributes["details"])

			details := resourceSchema.Attributes["details"].(schema.StringAttribute)
			assert.NotEmpty(t, details.Description)
			assert.True(t, details.IsOptional())
		})

		t.Run("details_secure should be configured", func(t *testing.T) {
			assert.IsType(t, schema.StringAttribute{}, resourceSchema.Attributes["details_secure"])

			detailsSecure := resourceSchema.Attributes["details_secure"].(schema.StringAttribute)
			assert.NotEmpty(t, detailsSecure.Description)
			assert.True(t, detailsSecure.IsOptional())
			assert.True(t, detailsSecure.IsSensitive())
		})

		t.Run("schedules should be configured", func(t *testing.T) {
			assert.IsType(t, schema.MapAttribute{}, resourceSchema.Attributes["schedules"])

			schedules := resourceSchema.Attributes["schedules"].(schema.MapAttribute)
			assert.NotEmpty(t, schedules.Description)
			assert.True(t, schedules.IsComputed())
			assert.False(t, schedules.IsRequired())
			assert.False(t, schedules.IsOptional())
		})
	})
}
