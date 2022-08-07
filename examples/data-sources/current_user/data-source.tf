data "metabase_current_user" "this" {}

locals {
  current_user_id = data.metabase_current_user.this.id
}
