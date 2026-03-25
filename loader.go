package main

import (
	"fmt"
	"os"

	"ariga.io/atlas-provider-gorm/gormschema"
	"example/downhill-api/database" 
)

func main() {

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