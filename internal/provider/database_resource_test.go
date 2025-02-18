package provider

import (
	"github.com/bnjns/metabase-sdk-go/service/database"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/assert"
	"testing"
)

var ignoredDatabaseImportAttributes = []string{"details", "details_secure"}

func TestIsSensitiveDatabaseDetail(t *testing.T) {
	t.Parallel()

	t.Run("detail with sensitive key should be sensitive", func(t *testing.T) {
		result := isSensitiveDatabaseDetail("password", "example")

		assert.True(t, result)
	})

	t.Run("detail with a non-string value should not be sensitive", func(t *testing.T) {
		result := isSensitiveDatabaseDetail("port", 5432)

		assert.False(t, result)
	})

	t.Run("detail with a redacted string value should be sensitive", func(t *testing.T) {
		result := isSensitiveDatabaseDetail("field", "**MetabasePass**")

		assert.True(t, result)
	})

	t.Run("detail with a non-redacted string value should not be sensitive", func(t *testing.T) {
		result := isSensitiveDatabaseDetail("host", "localhost")

		assert.False(t, result)
	})
}

func TestCheckDatabaseDetails(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		engine         database.Engine
		expectedErrors []error
	}{
		{
			engine:         database.EnginePostgres,
			expectedErrors: []error{errMissingDbName, errMissingHost, errMissingUser, errMissingPassword},
		},
	}

	for _, testCase := range testCases {
		t.Run(string(testCase.engine), func(t *testing.T) {
			errs := checkDatabaseDetails(testCase.engine, map[string]interface{}{})

			assert.NotZero(t, len(errs)) // TODO: remove when all engines are tested
			assert.Len(t, errs, len(testCase.expectedErrors))
			assert.ElementsMatch(t, errs, testCase.expectedErrors)
		})
	}
}

func TestBuildSchedules(t *testing.T) {
	t.Parallel()

	t.Run("a database with no schedules should return a null object", func(t *testing.T) {
		db := database.Database{
			Schedules: nil,
		}

		schedules, diags := buildSchedules(&db)
		assert.Zero(t, len(diags))
		assert.True(t, schedules.IsNull())
	})

	t.Run("a database with valid schedules should return a non-empty map", func(t *testing.T) {
		hour := int64(1)
		minuteOnHour := int64(0)
		minuteEvery := int64(1)
		db := database.Database{
			Schedules: &database.Schedules{
				MetadataSync: &database.ScheduleSettings{
					Type:   database.ScheduleTypeDaily,
					Day:    nil,
					Frame:  nil,
					Hour:   &hour,
					Minute: &minuteOnHour,
				},
				CacheFieldValues: &database.ScheduleSettings{
					Type:   database.ScheduleTypeHourly,
					Day:    nil,
					Frame:  nil,
					Hour:   nil,
					Minute: &minuteEvery,
				},
			},
		}

		schedules, diags := buildSchedules(&db)
		assert.Zero(t, len(diags))
		assert.False(t, schedules.IsNull())
		assert.False(t, schedules.IsUnknown())

		assert.Equal(t, 2, len(schedules.Attributes()))
	})

	t.Run("a database with nil schedules should return a map with null values", func(t *testing.T) {
		db := database.Database{
			Schedules: &database.Schedules{
				MetadataSync:     nil,
				CacheFieldValues: nil,
			},
		}

		schedules, diags := buildSchedules(&db)
		assert.Zero(t, len(diags))
		assert.False(t, schedules.IsNull())
		assert.False(t, schedules.IsUnknown())

		assert.Equal(t, 2, len(schedules.Attributes()))
		for _, schedule := range schedules.Attributes() {
			assert.True(t, schedule.IsNull())
		}
	})
}

func TestBuildDatabaseDetails(t *testing.T) {
	t.Parallel()

	t.Run("a database with no details should return null", func(t *testing.T) {
		db := database.Database{
			Details: nil,
		}

		details, detailsSecure, diags := buildDatabaseDetails(&db)
		assert.Zero(t, len(diags))
		assert.True(t, details.IsNull())
		assert.True(t, detailsSecure.IsNull())
	})

	t.Run("a database with an empty details map should be parsed as empty json objects", func(t *testing.T) {
		db := database.Database{
			Details: &database.Details{},
		}

		details, detailsSecure, diags := buildDatabaseDetails(&db)
		assert.Zero(t, len(diags))
		assert.Equal(t, `{}`, details.ValueString())
		assert.Equal(t, `{}`, detailsSecure.ValueString())
	})

	t.Run("a database with details should have the sensitive details separated", func(t *testing.T) {
		db := database.Database{
			Details: &database.Details{
				"password": "password",
				"name":     "database name",
				"redacted": "**MetabasePass**",
			},
		}

		details, detailsSecure, diags := buildDatabaseDetails(&db)
		assert.Zero(t, len(diags))
		assert.Equal(t, `{"name":"database name"}`, details.ValueString())
		assert.Equal(t, `{"password":"password","redacted":"**MetabasePass**"}`, detailsSecure.ValueString())
	})
}

func TestAccDatabaseResource_PostgreSQL(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "metabase_database" "test" {
	engine = "postgres"
	name   = "Test PostgreSQL"

	details = jsonencode({
		host   = "postgres"
		port   = 5432
		dbname = "postgres"
		user   = "postgres"
	})
	details_secure = jsonencode({
		password = "postgres"
	})
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("metabase_database.test", "id"),
					resource.TestCheckResourceAttr("metabase_database.test", "engine", "postgres"),
					resource.TestCheckResourceAttr("metabase_database.test", "name", "Test PostgreSQL"),
					resource.TestCheckResourceAttrSet("metabase_database.test", "details"),
					resource.TestCheckResourceAttrSet("metabase_database.test", "details_secure"),
				),
			},
			{
				ResourceName:            "metabase_database.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: ignoredDatabaseImportAttributes,
			},
			{
				Config: providerConfig + `
resource "metabase_database" "test" {
	engine = "postgres"
	name   = "Test PostgreSQL (Updated)"

	details = jsonencode({
		host   = "postgres"
		port   = 5432
		dbname = "postgres"
		user   = "postgres"
	})
	details_secure = jsonencode({
		password = "postgres"
	})
}
			`,
				Check: resource.TestCheckResourceAttr("metabase_database.test", "name", "Test PostgreSQL (Updated)"),
			},
		},
	})
}
