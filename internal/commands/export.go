package commands

import (
	"fmt"
	"time"
	"os"
	"io"
	"crypto/rand"
	"math/big"

	"EverythingSuckz/fsb/config"
	"EverythingSuckz/fsb/internal/database"

	"github.com/celestix/gotgproto/dispatcher"
	"github.com/celestix/gotgproto/dispatcher/handlers"
	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/tg"
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

	// Get user count for summary
	totalUsers, _ := database.DB.GetTotalUserCount()

	// Send the file via Telegram
	err = m.sendFileToTelegram(ctx, chatID, exportData, filename, totalUsers, timestamp)
	if err != nil {
		m.log.Sugar().Errorf("Failed to send file via Telegram: %v", err)
		// Fallback: save locally and send message
		os.WriteFile(filename, exportData, 0644)
		ctx.Reply(u, fmt.Sprintf("📊 Database Export Complete!\n\n📈 Total users: %d\n📅 Export time: %s\n💾 File: %s\n\n⚠️ File saved locally (Telegram upload failed)", totalUsers, timestamp, filename), nil)
		return err
	}

	m.log.Sugar().Infof("Database exported successfully for admin %d, sent via Telegram", chatID)
	return dispatcher.EndGroups
}

// sendFileToTelegram sends the export file directly through Telegram
func (m *command) sendFileToTelegram(ctx *ext.Context, chatID int64, fileData []byte, filename string, totalUsers int64, timestamp string) error {
	// Create a temporary file for upload
	tempFile := "/tmp/" + filename
	err := os.WriteFile(tempFile, fileData, 0644)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile)

	// Open the file
	file, err := os.Open(tempFile)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// Get file size
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %v", err)
	}
	fileSize := fileInfo.Size()

	// Generate random ID for upload
	randomID, _ := rand.Int(rand.Reader, big.NewInt(9223372036854775807))

	// Read file in chunks and upload
	const chunkSize = 512 * 1024 // 512KB chunks
	var part int32 = 0
	
	for {
		chunk := make([]byte, chunkSize)
		n, err := file.Read(chunk)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read file chunk: %v", err)
		}

		// Upload this chunk
		uploaded, err := ctx.Raw.UploadSaveFilePart(ctx, &tg.UploadSaveFilePartRequest{
			FileID:   randomID.Int64(),
			FilePart: part,
			Bytes:    chunk[:n],
		})
		if err != nil {
			return fmt.Errorf("failed to upload file part %d: %v", part, err)
		}
		if !uploaded {
			return fmt.Errorf("upload failed for part %d", part)
		}
		part++
	}

	// Create InputFile for the uploaded data
	inputFile := &tg.InputFile{
		ID:    randomID.Int64(),
		Parts: part,
		Name:  filename,
	}

	// Create document media
	media := &tg.InputMediaUploadedDocument{
		File:     inputFile,
		MimeType: "application/json",
		Attributes: []tg.DocumentAttributeClass{
			&tg.DocumentAttributeFilename{
				FileName: filename,
			},
		},
	}

	// Create caption with summary
	caption := fmt.Sprintf("📊 **Database Export Complete!**\n\n📈 Total users: %d\n📅 Export time: %s\n💾 File size: %.2f KB\n\n✅ Complete user database backup!", totalUsers, timestamp, float64(fileSize)/1024)

	// Send the document
	_, err = ctx.Raw.MessagesSendMedia(ctx, &tg.MessagesSendMediaRequest{
		Peer:     ctx.PeerStorage.GetInputPeerById(chatID),
		Media:    media,
		Message:  caption,
		RandomID: randomID.Int64(),
	})

	if err != nil {
		return fmt.Errorf("failed to send document: %v", err)
	}

	return nil
}
