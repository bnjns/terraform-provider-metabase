package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/assert"
	"terraform-provider-metabase/internal/client"
	"testing"
)

func TestCheckDatabaseDetails(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		engine         client.DatabaseEngine
		expectedErrors []error
	}{
		{
			engine:         client.EngineAmazonRedshift,
			expectedErrors: []error{errMissingDbName, errMissingHost, errMissingPort, errMissingUsername, errMissingPassword},
		},
		{
			engine:         client.EngineBigQuery,
			expectedErrors: []error{errMissingGcpCredentials},
		},
		{
			engine:         client.EngineDruid,
			expectedErrors: []error{errMissingHost, errMissingPort},
		},
		{
			engine:         client.EngineGoogleAnalytics,
			expectedErrors: []error{errMissingGcpCredentials},
		},
		{
			engine:         client.EngineH2,
			expectedErrors: []error{errMissingConnString},
		},
		{
			engine:         client.EngineMongoDB,
			expectedErrors: []error{errMissingDbName, errMissingHost, errMissingPort, errMissingUsername, errMissingPassword},
		},
		{
			engine:         client.EngineMySQL,
			expectedErrors: []error{errMissingDbName, errMissingHost, errMissingPort, errMissingUsername, errMissingPassword},
		},
		{
			engine:         client.EngineOracle,
			expectedErrors: []error{},
		},
		{
			engine:         client.EnginePostgres,
			expectedErrors: []error{errMissingDbName, errMissingHost, errMissingPort, errMissingUsername, errMissingPassword},
		},
		{
			engine:         client.EnginePresto,
			expectedErrors: []error{},
		},
		{
			engine:         client.EnginePrestoDeprecated,
			expectedErrors: []error{},
		},
		{
			engine:         client.EngineSnowflake,
			expectedErrors: []error{},
		},
		{
			engine:         client.EngineSparkSQL,
			expectedErrors: []error{},
		},
		{
			engine:         client.EngineSQLServer,
			expectedErrors: []error{},
		},
		{
			engine:         client.EngineSQLite,
			expectedErrors: []error{},
		},
	}

	for _, testCase := range testCases {
		t.Run(string(testCase.engine), func(t *testing.T) {
			errs := checkDatabaseDetails(testCase.engine, types.Map{})

			assert.NotZero(t, len(errs)) // TODO: remove when all engines are tested
			assert.Len(t, errs, len(testCase.expectedErrors))
			assert.ElementsMatch(t, errs, testCase.expectedErrors)
		})
	}
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

	details = {
		db = "zip:/app/metabase.jar!/sample-database.db;USER=GUEST;PASSWORD=guest"
	}
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("metabase_database.test", "id"),
					resource.TestCheckResourceAttr("metabase_database.test", "name", "Test H2"),
					resource.TestCheckResourceAttrSet("metabase_database.test", "details.db"),
				),
			},
			{
				ResourceName:      "metabase_database.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: providerConfig + `
resource "metabase_database" "test" {
	engine = "h2"
	name   = "Test H2 (Updated)"

	details = {
		db = "zip:/app/metabase.jar!/sample-database.db;USER=GUEST;PASSWORD=guest"
	}
}
`,
				Check: resource.TestCheckResourceAttr("metabase_database.test", "name", "Test H2 (Updated)"),
			},
		},
	})
}
