resource "metabase_database" "example" {
  engine = "postgres"
  name   = "PostgreSQL database"

  details = jsonencode({
    host   = "localhost"
    port   = 5432
    dbname = "database"
    user   = "username"
  })
  details_secure = jsonencode({
    password = "password"
  })
}
