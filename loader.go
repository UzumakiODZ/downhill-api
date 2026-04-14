package main

import (
	"fmt"
	"os"

	"downhill-api/database"

	"ariga.io/atlas-provider-gorm/gormschema"
)

func printSchemaStatements() {

	stmts, err := gormschema.New("postgres").Load(
		&database.User{},
		&database.Company{},
		&database.Role{},
		&database.QuestionBank{},
		&database.Post{},
	)

	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load gorm schema: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(stmts)
}
