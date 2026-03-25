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
  
  # Change this line:
  dev = "postgres://postgres:mysecretpassword@127.0.0.1:5499/postgres?sslmode=disable"

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