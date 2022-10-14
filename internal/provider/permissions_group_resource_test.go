package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccPermissionsGroupResource_Basic(t *testing.T) {
	name := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "metabase_permissions_group" "test" {
	name = "%s"
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("metabase_permissions_group.test", "id"),
					resource.TestCheckResourceAttr("metabase_permissions_group.test", "name", name),
				),
			},
			{
				ResourceName:      "metabase_permissions_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccPermissionsGroupResource_Update(t *testing.T) {
	originalName := acctest.RandString(10)
	updatedName := acctest.RandString(11)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "metabase_permissions_group" "test" {
	name = "%s"
}
`, originalName),
			},
			{
				Config: providerConfig + fmt.Sprintf(`
resource "metabase_permissions_group" "test" {
	name = "%s"
}
`, updatedName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("metabase_permissions_group.test", "name", updatedName),
				),
			},
		},
	})
}
