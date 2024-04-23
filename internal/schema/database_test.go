package schema

import (
	"context"
	dSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/stretchr/testify/assert"
	"terraform-provider-metabase/internal/validators"
	"testing"
)

func TestDatabaseResource(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("database rSchema should return expected fields", func(t *testing.T) {
		resourceSchema := DatabaseResource()

		assert.NotEmpty(t, resourceSchema.Description)
		assert.Equal(t, 7, len(resourceSchema.Attributes))

		t.Run("id should be configured", func(t *testing.T) {
			assert.IsType(t, rSchema.Int64Attribute{}, resourceSchema.Attributes["id"])

			id := resourceSchema.Attributes["id"].(rSchema.Int64Attribute)
			assert.NotEmpty(t, id.Description)
			assert.True(t, id.IsComputed())
			assert.False(t, id.IsRequired())
			assert.False(t, id.IsOptional())
			assert.Contains(t, id.Int64PlanModifiers(), int64planmodifier.UseStateForUnknown())
		})

		t.Run("engine should be configured", func(t *testing.T) {
			assert.IsType(t, rSchema.StringAttribute{}, resourceSchema.Attributes["engine"])

			engine := resourceSchema.Attributes["engine"].(rSchema.StringAttribute)
			assert.NotEmpty(t, engine.Description)
			assert.True(t, engine.IsRequired())
			assert.Contains(t, engine.StringValidators(), validators.IsKnownDatabaseEngineValidator())
			assert.Equal(t, "If the value of this attribute changes, Terraform will destroy and recreate the resource.", engine.StringPlanModifiers()[0].Description(ctx))
		})

		t.Run("name should be configured", func(t *testing.T) {
			assert.IsType(t, rSchema.StringAttribute{}, resourceSchema.Attributes["name"])

			name := resourceSchema.Attributes["name"].(rSchema.StringAttribute)
			assert.NotEmpty(t, name.Description)
			assert.True(t, name.IsRequired())
		})

		t.Run("features should be configured", func(t *testing.T) {
			assert.IsType(t, rSchema.ListAttribute{}, resourceSchema.Attributes["features"])

			features := resourceSchema.Attributes["features"].(rSchema.ListAttribute)
			assert.NotEmpty(t, features.Description)
			assert.True(t, features.IsComputed())
			assert.False(t, features.IsRequired())
			assert.False(t, features.IsOptional())
			assert.Contains(t, features.ListPlanModifiers(), listplanmodifier.UseStateForUnknown())
		})

		t.Run("details should be configured", func(t *testing.T) {
			assert.IsType(t, rSchema.StringAttribute{}, resourceSchema.Attributes["details"])

			details := resourceSchema.Attributes["details"].(rSchema.StringAttribute)
			assert.NotEmpty(t, details.Description)
			assert.True(t, details.IsOptional())
		})

		t.Run("details_secure should be configured", func(t *testing.T) {
			assert.IsType(t, rSchema.StringAttribute{}, resourceSchema.Attributes["details_secure"])

			detailsSecure := resourceSchema.Attributes["details_secure"].(rSchema.StringAttribute)
			assert.NotEmpty(t, detailsSecure.Description)
			assert.True(t, detailsSecure.IsOptional())
			assert.True(t, detailsSecure.IsSensitive())
		})

		t.Run("schedules should be configured", func(t *testing.T) {
			assert.IsType(t, rSchema.ObjectAttribute{}, resourceSchema.Attributes["schedules"])

			schedules := resourceSchema.Attributes["schedules"].(rSchema.ObjectAttribute)
			assert.NotEmpty(t, schedules.Description)
			assert.True(t, schedules.IsComputed())
			assert.False(t, schedules.IsRequired())
			assert.False(t, schedules.IsOptional())
		})
	})
}

func TestDatabaseDataSource(t *testing.T) {
	t.Parallel()

	t.Run("database rSchema should return expected fields", func(t *testing.T) {
		dataSourceSchema := DatabaseDataSource()

		assert.NotEmpty(t, dataSourceSchema.Description)
		assert.Equal(t, 6, len(dataSourceSchema.Attributes))

		t.Run("id should be configured", func(t *testing.T) {
			assert.IsType(t, dSchema.Int64Attribute{}, dataSourceSchema.Attributes["id"])

			id := dataSourceSchema.Attributes["id"].(dSchema.Int64Attribute)
			assert.NotEmpty(t, id.Description)
			assert.True(t, id.IsRequired())
		})

		t.Run("engine should be configured", func(t *testing.T) {
			assert.IsType(t, dSchema.StringAttribute{}, dataSourceSchema.Attributes["engine"])

			engine := dataSourceSchema.Attributes["engine"].(dSchema.StringAttribute)
			assert.NotEmpty(t, engine.Description)
			assert.True(t, engine.IsComputed())
		})

		t.Run("name should be configured", func(t *testing.T) {
			assert.IsType(t, dSchema.StringAttribute{}, dataSourceSchema.Attributes["name"])

			name := dataSourceSchema.Attributes["name"].(dSchema.StringAttribute)
			assert.NotEmpty(t, name.Description)
			assert.True(t, name.IsComputed())
		})

		t.Run("features should be configured", func(t *testing.T) {
			assert.IsType(t, dSchema.ListAttribute{}, dataSourceSchema.Attributes["features"])

			features := dataSourceSchema.Attributes["features"].(dSchema.ListAttribute)
			assert.NotEmpty(t, features.Description)
			assert.True(t, features.IsComputed())
		})

		t.Run("details should be configured", func(t *testing.T) {
			assert.IsType(t, dSchema.StringAttribute{}, dataSourceSchema.Attributes["details"])

			details := dataSourceSchema.Attributes["details"].(dSchema.StringAttribute)
			assert.NotEmpty(t, details.Description)
			assert.True(t, details.IsComputed())
		})

		t.Run("schedules should be configured", func(t *testing.T) {
			assert.IsType(t, dSchema.ObjectAttribute{}, dataSourceSchema.Attributes["schedules"])

			schedules := dataSourceSchema.Attributes["schedules"].(dSchema.ObjectAttribute)
			assert.NotEmpty(t, schedules.Description)
			assert.True(t, schedules.IsComputed())
		})
	})
}
