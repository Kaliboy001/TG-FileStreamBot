package models

import (
	"gorm.io/gorm"
)

// User represents a user who has started the bot.
// The GORM model handles fields like ID, CreatedAt, etc. automatically.
type User struct {
	gorm.Model
	UserID int64 `gorm:"uniqueIndex"` // The Telegram user's ID. uniqueIndex ensures no duplicate IDs.
}
