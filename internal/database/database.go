package database

import (
	"time"

	"github.com/glebarez/sqlite"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// User represents a user record
type User struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    int64     `gorm:"uniqueIndex;not null"`
	Username  string    `gorm:"size:255"`
	FirstSeen time.Time `gorm:"not null"`
	LastSeen  time.Time `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Database holds the database connection
type Database struct {
	db  *gorm.DB
	log *zap.Logger
}

var DB *Database

// Connect initializes the database connection
func Connect(mongoURI string, log *zap.Logger) error {
	log = log.Named("Database")
	
	// Use SQLite for persistent storage (works on all cloud platforms)
	db, err := gorm.Open(sqlite.Open("users.db"), &gorm.Config{})
	if err != nil {
		log.Error("Failed to connect to database", zap.Error(err))
		return err
	}

	// Auto-migrate the schema
	if err := db.AutoMigrate(&User{}); err != nil {
		log.Error("Failed to migrate database", zap.Error(err))
		return err
	}

	DB = &Database{
		db:  db,
		log: log,
	}

	log.Info("Successfully connected to SQLite database")
	return nil
}

// Disconnect closes the database connection
func Disconnect() error {
	if DB != nil && DB.db != nil {
		sqlDB, err := DB.db.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// IsUserSeen checks if a user has been seen before
func (db *Database) IsUserSeen(userID int64) (bool, error) {
	var count int64
	err := db.db.Model(&User{}).Where("user_id = ?", userID).Count(&count).Error
	if err != nil {
		db.log.Error("Failed to check if user exists", zap.Error(err), zap.Int64("user_id", userID))
		return false, err
	}
	
	return count > 0, nil
}

// AddUser adds a new user to the database
func (db *Database) AddUser(userID int64, username string) error {
	now := time.Now()
	user := User{
		UserID:    userID,
		Username:  username,
		FirstSeen: now,
		LastSeen:  now,
	}

	err := db.db.Create(&user).Error
	if err != nil {
		db.log.Error("Failed to add user", zap.Error(err), zap.Int64("user_id", userID))
		return err
	}

	db.log.Info("New user added", zap.Int64("user_id", userID), zap.String("username", username))
	return nil
}

// UpdateUserLastSeen updates the last seen time for an existing user
func (db *Database) UpdateUserLastSeen(userID int64) error {
	err := db.db.Model(&User{}).Where("user_id = ?", userID).Update("last_seen", time.Now()).Error
	if err != nil {
		db.log.Error("Failed to update user last seen", zap.Error(err), zap.Int64("user_id", userID))
		return err
	}

	return nil
}

// GetTotalUserCount returns the total number of unique users
func (db *Database) GetTotalUserCount() (int64, error) {
	var count int64
	err := db.db.Model(&User{}).Count(&count).Error
	if err != nil {
		db.log.Error("Failed to get total user count", zap.Error(err))
		return 0, err
	}

	return count, nil
}

// GetAllUsers returns all users (for admin purposes)
func (db *Database) GetAllUsers() ([]User, error) {
	var users []User
	err := db.db.Find(&users).Error
	if err != nil {
		db.log.Error("Failed to get all users", zap.Error(err))
		return nil, err
	}

	return users, nil
}
