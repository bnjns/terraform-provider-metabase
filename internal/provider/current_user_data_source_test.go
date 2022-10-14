package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccCurrentUserDataSource_Read(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `data "metabase_current_user" "test" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.metabase_current_user.test", "id", "1"),
					resource.TestCheckResourceAttr("data.metabase_current_user.test", "email", "example@example.com"),
					resource.TestCheckResourceAttr("data.metabase_current_user.test", "first_name", "Example"),
					resource.TestCheckResourceAttr("data.metabase_current_user.test", "last_name", "User"),
					resource.TestCheckResourceAttr("data.metabase_current_user.test", "common_name", "Example User"),
					resource.TestCheckNoResourceAttr("data.metabase_current_user.test", "locale"),
					resource.TestCheckResourceAttr("data.metabase_current_user.test", "group_ids.#", "0"),

					resource.TestCheckResourceAttrSet("data.metabase_current_user.test", "google_auth"),
					resource.TestCheckResourceAttrSet("data.metabase_current_user.test", "ldap_auth"),

					resource.TestCheckResourceAttr("data.metabase_current_user.test", "is_active", "true"),
					resource.TestCheckResourceAttr("data.metabase_current_user.test", "is_installer", "true"),
					resource.TestCheckResourceAttr("data.metabase_current_user.test", "is_qbnewb", "true"),
					resource.TestCheckResourceAttr("data.metabase_current_user.test", "is_superuser", "true"),
					resource.TestCheckResourceAttr("data.metabase_current_user.test", "has_invited_second_user", "false"),
					resource.TestCheckResourceAttr("data.metabase_current_user.test", "has_question_and_dashboard", "false"),

					resource.TestCheckResourceAttrSet("data.metabase_current_user.test", "date_joined"),
					resource.TestCheckResourceAttrSet("data.metabase_current_user.test", "first_login"),
					resource.TestCheckResourceAttrSet("data.metabase_current_user.test", "last_login"),
					resource.TestCheckResourceAttrSet("data.metabase_current_user.test", "updated_at"),
				),
			},
		},
	})
}
