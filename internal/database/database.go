package database

import (
	"EverythingSuckz/fsb/internal/models"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// DB is the global database connection instance.
var DB *gorm.DB

// InitDB connects to the database and runs auto-migration.
func InitDB() {
	var err error
	DB, err = gorm.Open(sqlite.Open("fsb.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto-migrate the schema to create the 'users' table
	err = DB.AutoMigrate(&models.User{})
	if err != nil {
		log.Fatalf("Failed to auto-migrate database: %v", err)
	}
}
