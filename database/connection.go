package database

import (
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	dsn := os.Getenv("DB_URL")
	if dsn == "" {
		log.Fatal("DB_URL environment variable not set")
	}
	var err error

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	err = DB.AutoMigrate(
		&Company{},
		&Role{},
		&QuestionBank{},
		&Post{},
	) 

	if err != nil {
		log.Fatalf("Failed to migrate: %v", err)
	}
	
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	log.Println("Database connection established")
}