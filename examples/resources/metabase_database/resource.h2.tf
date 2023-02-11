resource "metabase_database" "example" {
  engine = "h2"
  name   = "H2 database"

  details = jsonencode({
    db = "/path/to/database/file.db;USER=username;PASSWORD=password"
  })
}
