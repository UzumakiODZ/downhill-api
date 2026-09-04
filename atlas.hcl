data "external_schema" "gorm" {
  program = [
    "go",
    "run",
    "-mod=mod",
    "./loader.go"
  ]
}

env "gorm" {
  src = data.external_schema.gorm.url
  
  dev = try(getenv("DEV_DB_URL"), "docker://postgres/15/dev")
  url = getenv("DB_URL")

  schemas = ["public"]
  
  migration {
    dir = "file://migrations"
  }
  
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}