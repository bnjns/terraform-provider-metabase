package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

var (
	resourceFirstName = "Test"
	resourceLastName  = "User"
)

func TestAccUserResource_Basic(t *testing.T) {
	resourceEmail := testAccRandEmail()

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
`, resourceEmail, resourceFirstName, resourceLastName),
				Check: testAccCheckUserConf("metabase_user.test", resourceEmail, resourceFirstName, resourceLastName, false),
			},
			{
				ResourceName:      "metabase_user.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccUserResource_NoName(t *testing.T) {
	resourceEmail := testAccRandEmail()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "metabase_user" "test" {
	email      = "%s"
}
`, resourceEmail),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("metabase_user.test", "first_name"),
					resource.TestCheckNoResourceAttr("metabase_user.test", "last_name"),
					resource.TestCheckResourceAttr("metabase_user.test", "common_name", resourceEmail),
				),
			},
			{
				ResourceName:      "metabase_user.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccUserResource_Superuser(t *testing.T) {
	resourceEmail := testAccRandEmail()

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
`, resourceEmail, resourceFirstName, resourceLastName),
				Check: testAccCheckUserConf("metabase_user.test", resourceEmail, resourceFirstName, resourceLastName, true),
			},
			{
				ResourceName:      "metabase_user.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccUserResource_Groups(t *testing.T) {
	t.Skip("To be added when groups are supported")
}

func testAccRandEmail() string {
	return fmt.Sprintf("%s@example.com", acctest.RandString(8))
}

func testAccCheckUserConf(resourceName string, email string, firstName string, lastName string, isSuperuser bool) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "email", email),
		resource.TestCheckResourceAttr(resourceName, "first_name", firstName),
		resource.TestCheckResourceAttr(resourceName, "last_name", lastName),
		resource.TestCheckResourceAttr(resourceName, "common_name", firstName+" "+lastName),
		resource.TestCheckNoResourceAttr(resourceName, "locale"),
		resource.TestCheckResourceAttr(resourceName, "group_ids.#", "0"),

		resource.TestCheckResourceAttrSet(resourceName, "google_auth"),
		resource.TestCheckResourceAttrSet(resourceName, "ldap_auth"),

		resource.TestCheckResourceAttr(resourceName, "is_active", "true"),
		resource.TestCheckResourceAttrSet(resourceName, "is_installer"),
		resource.TestCheckResourceAttrSet(resourceName, "is_qbnewb"),
		resource.TestCheckResourceAttr(resourceName, "is_superuser", fmt.Sprintf("%t", isSuperuser)),
		resource.TestCheckResourceAttrSet(resourceName, "has_invited_second_user"),
		resource.TestCheckResourceAttrSet(resourceName, "has_question_and_dashboard"),

		resource.TestCheckResourceAttrSet(resourceName, "date_joined"),
		resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
	)
}
