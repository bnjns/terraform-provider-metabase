package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestShouldIncludeDatabaseDetails(t *testing.T) {
	t.Parallel()

	t.Run("detail with sensitive key should not be included", func(t *testing.T) {
		result := shouldIncludeDatabaseDetail("password", "example")

		assert.False(t, result)
	})

	t.Run("detail with a non-string value should be included", func(t *testing.T) {
		result := shouldIncludeDatabaseDetail("port", 5432)

		assert.True(t, result)
	})

	t.Run("detail with a redacted string value should not be included", func(t *testing.T) {
		result := shouldIncludeDatabaseDetail("field", "**MetabasePass**")

		assert.False(t, result)
	})

	t.Run("detail with a non-redacted string value should be included", func(t *testing.T) {
		result := shouldIncludeDatabaseDetail("host", "localhost")

		assert.True(t, result)
	})
}

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
