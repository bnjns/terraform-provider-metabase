package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/assert"
	"terraform-provider-metabase/internal/client"
	"testing"
)

var ignoredDatabaseImportAttributes = []string{"details", "details_secure"}

func TestCheckDatabaseDetails(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		engine         client.DatabaseEngine
		expectedErrors []error
	}{
		{
			engine:         client.EngineH2,
			expectedErrors: []error{errMissingConnString},
		},
		{
			engine:         client.EnginePostgres,
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

	t.Run("a database with no schedules should return a non-null empty map", func(t *testing.T) {
		database := client.Database{
			Schedules: nil,
		}

		schedules, diags := buildSchedules(&database)
		assert.Zero(t, len(diags))
		assert.False(t, schedules.IsNull())
		assert.False(t, schedules.IsUnknown())
		assert.Zero(t, len(schedules.Elements()))
	})

	t.Run("a database with valid schedules should return a non-empty map", func(t *testing.T) {
		hour := int64(1)
		minuteOnHour := int64(0)
		minuteEvery := int64(1)
		database := client.Database{
			Schedules: &map[string]client.DatabaseSchedule{
				"example": {
					Day:    nil,
					Frame:  nil,
					Hour:   &hour,
					Minute: &minuteOnHour,
					Type:   "daily",
				},
				"another": {
					Day:    nil,
					Frame:  nil,
					Hour:   nil,
					Minute: &minuteEvery,
					Type:   "hourly",
				},
			},
		}

		schedules, diags := buildSchedules(&database)
		assert.Zero(t, len(diags))
		assert.False(t, schedules.IsNull())
		assert.False(t, schedules.IsUnknown())

		assert.Equal(t, 2, len(schedules.Elements()))
	})
}

func TestAccDatabaseResource_H2(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "metabase_database" "test" {
	engine = "h2"
	name   = "Test H2"

	details_secure = jsonencode({
		db = "zip:/app/metabase.jar!/sample-database.db;USER=GUEST;PASSWORD=guest"
	})
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("metabase_database.test", "id"),
					resource.TestCheckResourceAttr("metabase_database.test", "engine", "h2"),
					resource.TestCheckResourceAttr("metabase_database.test", "name", "Test H2"),
					resource.TestCheckNoResourceAttr("metabase_database.test", "details"),
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
	engine = "h2"
	name   = "Test H2 (Updated)"

	details_secure = jsonencode({
		db = "zip:/app/metabase.jar!/sample-database.db;USER=GUEST;PASSWORD=guest"
	})
}
			`,
				Check: resource.TestCheckResourceAttr("metabase_database.test", "name", "Test H2 (Updated)"),
			},
		},
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
