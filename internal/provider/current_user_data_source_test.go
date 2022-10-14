package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccCurrentUserDataSource_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `data "metabase_current_user" "test" {}`,
				Check:  testAccCheckUserConf("data.metabase_current_user.test", "example@example.com", "Example", "User", true),
			},
		},
	})
}
