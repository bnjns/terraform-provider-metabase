package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccUserDataSource_Basic(t *testing.T) {
	userEmail := testAccRandEmail()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "metabase_user" "test" {
	email      = "%s"
	first_name = "%s"
	last_name  = "%s"
}

data "metabase_user" "test" {
	id = metabase_user.test.id
}
`, userEmail, testAccUserFirstName, testAccUserLastName),
				Check: testAccCheckUserConf("data.metabase_user.test", userEmail, testAccUserFirstName, testAccUserLastName, false),
			},
		},
	})
}

func TestAccUserDataSource_Superuser(t *testing.T) {
	userEmail := testAccRandEmail()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "metabase_user" "test" {
	email        = "%s"
	first_name   = "%s"
	last_name    = "%s"
	is_superuser = true
}

data "metabase_user" "test" {
	id = metabase_user.test.id
}
`, userEmail, testAccUserFirstName, testAccUserLastName),
				Check: testAccCheckUserConf("data.metabase_user.test", userEmail, testAccUserFirstName, testAccUserLastName, true),
			},
		},
	})
}
