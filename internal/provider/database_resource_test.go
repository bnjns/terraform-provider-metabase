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

	t.Run("h2 engine", func(t *testing.T) {
		errs := checkDatabaseDetails(client.EngineH2, types.Map{})

		assert.Len(t, errs, 1)
		assert.ErrorIs(t, errs[0], errH2MissingConnString)
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
