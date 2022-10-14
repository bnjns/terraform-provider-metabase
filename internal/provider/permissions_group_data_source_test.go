package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccPermissionsGroupDataSource_Basic(t *testing.T) {
	groupName := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "metabase_permissions_group" "test" {
	name = "%s"
}
data "metabase_permissions_group" "test" {
	id = metabase_permissions_group.test.id
}
`, groupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.metabase_permissions_group.test", "id"),
					resource.TestCheckResourceAttr("data.metabase_permissions_group.test", "name", groupName),
				),
			},
		},
	})
}
