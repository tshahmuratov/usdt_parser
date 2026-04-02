env "local" {
  src = "file://migrations/schema.sql"
  dev = "docker://postgres/16/dev?search_path=public"
  migration {
    dir = "file://migrations/versions"
  }
}
