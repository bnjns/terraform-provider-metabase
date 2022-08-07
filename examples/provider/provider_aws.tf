provider "aws" {
  # Your AWS provider config goes here
}
data "aws_secretsmanager_secret" "metabase_credentials" {
  name = "metabase-credentials"
}
data "aws_secretsmanager_secret_version" "metabase_credentials" {
  secret_id = data.aws_secretsmanager_secret.metabase_credentials.id
}
locals {
  metabase_credentials = jsondecode(data.aws_secretsmanager_secret_version.metabase_credentials.secret_string)
}

provider "metabase" {
  host     = local.metabase_credentials["host"]
  username = local.metabase_credentials["username"]
  password = local.metabase_credentials["password"]
}
