package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccDatabaseDataSource_Basic(t *testing.T) {
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
data "metabase_database" "test" {
	id = metabase_database.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.metabase_database.test", "id"),
					resource.TestCheckResourceAttr("data.metabase_database.test", "name", "Test PostgreSQL"),
					resource.TestCheckResourceAttrSet("data.metabase_database.test", "details"),
				),
			},
		},
	})
}
