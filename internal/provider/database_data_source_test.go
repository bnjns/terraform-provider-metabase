package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccDatabaseDataSource_H2(t *testing.T) {
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
data "metabase_database" "test" {
	id = metabase_database.test.id 
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.metabase_database.test", "id"),
					resource.TestCheckResourceAttr("data.metabase_database.test", "engine", "h2"),
					resource.TestCheckResourceAttr("data.metabase_database.test", "name", "Test H2"),
					resource.TestCheckResourceAttrSet("data.metabase_database.test", "details.db"),
				),
			},
		},
	})
}
