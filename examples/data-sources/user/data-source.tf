data "metabase_user" "example" {
  id = 2
}

locals {
  user_email = data.metabase_user.example.email
}
