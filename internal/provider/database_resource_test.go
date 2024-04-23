package provider

import (
	"github.com/bnjns/metabase-sdk-go/service/database"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/assert"
	"testing"
)

var ignoredDatabaseImportAttributes = []string{"details", "details_secure"}

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
