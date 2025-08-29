package commands

import (
	"fmt"
	"time"
	"os"

	"EverythingSuckz/fsb/config"
	"EverythingSuckz/fsb/internal/database"

	"github.com/celestix/gotgproto/dispatcher"
	"github.com/celestix/gotgproto/dispatcher/handlers"
	"github.com/celestix/gotgproto/ext"
)

func (m *command) LoadExport(dispatcher dispatcher.Dispatcher) {
	log := m.log.Named("export")
	defer log.Sugar().Info("Loaded")
	
	dispatcher.AddHandler(
		handlers.NewCommand("export", m.exportHandler),
	)
}

func (m *command) exportHandler(ctx *ext.Context, u *ext.Update) error {
	chatID := u.EffectiveChat().GetID()
	
	// Check if user is admin
	if chatID != config.ValueOf.AdminUserID {
		ctx.Reply(u, "❌ You are not authorized to use this command.", nil)
		return dispatcher.EndGroups
	}

	// Send processing message
	ctx.Reply(u, "🔄 Generating database export...", nil)

	// Export database data
	exportData, err := database.DB.ExportData()
	if err != nil {
		m.log.Sugar().Errorf("Failed to export database: %v", err)
		ctx.Reply(u, "❌ Failed to export database data.", nil)
		return err
	}

	// Create filename with timestamp
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("fsb_database_export_%s.json", timestamp)

	// Write to temporary file
	tempFile := "/tmp/" + filename
	err = os.WriteFile(tempFile, exportData, 0644)
	if err != nil {
		m.log.Sugar().Errorf("Failed to create export file: %v", err)
		ctx.Reply(u, "❌ Failed to create export file.", nil)
		return err
	}
	defer os.Remove(tempFile)

	// Get user count for summary
	totalUsers, _ := database.DB.GetTotalUserCount()

	// Send summary and file info
	summaryMsg := fmt.Sprintf("📊 Database Export Complete!\n\n📈 Total users: %d\n📅 Export time: %s\n💾 File: %s\n\nFile saved locally for download.", totalUsers, timestamp, filename)
	ctx.Reply(u, summaryMsg, nil)

	return dispatcher.EndGroups
}
