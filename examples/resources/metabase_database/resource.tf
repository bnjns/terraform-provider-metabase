resource "metabase_database" "example" {
  engine = "h2"
  name   = "H2 database"

  details = jsonencode({
    db = "zip:/app/metabase.jar!/sample-database.db;USER=GUEST;PASSWORD=guest"
  })

  schedules = {
    metadata_sync = {
      type   = "hourly"
      minute = 50
    }
  }
}
