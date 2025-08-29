package userdb

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/glebarez/sqlite" // Import the SQLite driver
	"go.uber.org/zap"
)

var db *sql.DB

// InitUserDB initializes the SQLite database for user storage
func InitUserDB(log *zap.Logger) error {
	log = log.Named("UserDB")
	var err error
	// Open the SQLite database. We'll use a separate file for user data.
	db, err = sql.Open("sqlite", "users.db")
	if err != nil {
		log.Error("Failed to open user database", zap.Error(err))
		return fmt.Errorf("failed to open user database: %w", err)
	}

	// Create the users table if it doesn't exist
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY,
		joined_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Error("Failed to create users table", zap.Error(err))
		return fmt.Errorf("failed to create users table: %w", err)
	}

	log.Info("Successfully initialized user database and table.")
	return nil
}

// CloseUserDB closes the database connection
func CloseUserDB(log *zap.Logger) {
	log = log.Named("UserDB")
	if db != nil {
		err := db.Close()
		if err != nil {
			log.Error("Failed to close user database", zap.Error(err))
		} else {
			log.Info("User database closed.")
		}
	}
}

// SaveUser saves a user's ID to the database if it doesn't already exist
func SaveUser(log *zap.Logger, userID int64) error {
	log = log.Named("UserDB")

	// Check if the user already exists
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE id = ?", userID).Scan(&count)
	if err != nil {
		log.Error("Error checking for existing user", zap.Error(err))
		return fmt.Errorf("error checking for existing user: %w", err)
	}

	if count > 0 {
		log.Sugar().Infof("User %d already exists in the database. Skipping save.", userID)
		return nil
	}

	// If the user doesn't exist, insert the new user
	insertUserSQL := `INSERT INTO users(id, joined_at) VALUES(?, ?)`
	_, err = db.Exec(insertUserSQL, userID, time.Now())
	if err != nil {
		log.Error("Failed to save user to database", zap.Error(err))
		return fmt.Errorf("failed to save user %d to database: %w", userID, err)
	}

	log.Sugar().Infof("Successfully saved user %d to database", userID)
	return nil
}
