resource "metabase_user" "example" {
  email      = "email@example.com"
  first_name = "Example"
  last_name  = "User"
  group_ids  = [3, 4]
}
