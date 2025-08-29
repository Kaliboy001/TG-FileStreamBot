package commands

import (
	"EverythingSuckz/fsb/config"
	"EverythingSuckz/fsb/internal/database" // Import the new database package
	"EverythingSuckz/fsb/internal/models"    // Import the new models package
	"EverythingSuckz/fsb/internal/utils"
	"errors" // Import the errors package
	"fmt"
	"sync/atomic"

	"github.com/celestix/gotgproto/dispatcher"
	"github.com/celestix/gotgproto/dispatcher/handlers"
	"github.com/celestix/gotgproto/ext"
	"github.com/celestix/gotgproto/storage"
	"github.com/gotd/td/tg"
	"gorm.io/gorm" // Import gorm for error checking
)

// The Telegram User ID of the admin who will receive notifications.
const adminID int64 = 6070733162

func (m *command) LoadStart(disp dispatcher.Dispatcher) {
	log := m.log.Named("start")
	defer log.Sugar().Info("Loaded")

	disp.AddHandler(handlers.NewCommand("start", func(ctx *ext.Context, u *ext.Update) error {
		chatId := u.EffectiveChat().GetID()
		
		peerChatId := ctx.PeerStorage.GetPeerById(chatId)
		if peerChatId.Type != int(storage.TypeUser) {
			return dispatcher.EndGroups
		}
		if len(config.ValueOf.AllowedUsers) != 0 && !utils.Contains(config.ValueOf.AllowedUsers, chatId) {
			ctx.Reply(u, "You are not allowed to use this bot.", nil)
			return dispatcher.EndGroups
		}
		
		// --- New Feature: Admin Notification for New Users ---
		
		var existingUser models.User
		// Check if the user already exists in the database
		result := database.DB.Where("user_id = ?", chatId).First(&existingUser)

		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// This is a new user!
			newUser := models.User{UserID: chatId}
			database.DB.Create(&newUser)

			// Get the updated total user count from the database
			var totalUsers int64
			database.DB.Model(&models.User{}).Count(&totalUsers)

			userUsername := "N/A"
			if u.EffectiveUser() != nil && u.EffectiveUser().Username != "" {
				userUsername = "@" + u.EffectiveUser().Username
			}

			notificationMessage := fmt.Sprintf(
				"➕ New User Notification ➕\n👤 User: %s\n🆔 User ID: %d\n📊 Total Users of Bot: %d",
				userUsername,
				chatId,
				totalUsers,
			)
			
			sendMessageRequest := &tg.MessagesSendMessageRequest{
				Peer:    ctx.PeerStorage.GetInputPeerById(adminID),
				Message: notificationMessage,
			}
			
			_, err := ctx.SendMessage(adminID, sendMessageRequest)
			if err != nil {
				m.log.Sugar().Errorf("Failed to send new user notification to admin (%d): %v", adminID, err)
			}
		}

		// --- End of New Feature ---
		
		// The original logic for all users proceeds here.
		showChannelJoinMessage(ctx, u)
		return dispatcher.EndGroups
	}))

	disp.AddHandler(handlers.NewCallbackQuery(nil, handleCallbacks))
}

func showChannelJoinMessage(ctx *ext.Context, u *ext.Update) {
	markup := &tg.ReplyInlineMarkup{
		Rows: []tg.KeyboardButtonRow{
			{
				Buttons: []tg.KeyboardButtonClass{
					&tg.KeyboardButtonURL{
						Text: "Join Channel",
						URL:  "https://t.me/KaIi_Bots",
					},
					&tg.KeyboardButtonCallback{
						Text: "🔐 Joined",
						Data: []byte("check_membership"),
					},
				},
			},
		},
	}
	
	ctx.Reply(u, "⚠️ To use this bot, you must first join our Telegram channels\n\nAfter successfully joining, click the 🔐 Joined button to confirm your bot membership and to continue.", &ext.ReplyOpts{
		Markup: markup,
	})
}

func handleCallbacks(ctx *ext.Context, u *ext.Update) error {
	callbackQuery := u.CallbackQuery
	if callbackQuery == nil {
		return dispatcher.EndGroups
	}

	callbackData := string(callbackQuery.Data)
	chatID := callbackQuery.UserID

	switch callbackData {
	case "check_membership":
		ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
			QueryID: callbackQuery.QueryID,
			Message: "",
		})
		
		markup := &tg.ReplyInlineMarkup{
			Rows: []tg.KeyboardButtonRow{
				{
					Buttons: []tg.KeyboardButtonClass{
						&tg.KeyboardButtonCallback{
							Text: "Dev",
							Data: []byte("dev_info"),
						},
					},
				},
			},
		}

		_, err := ctx.EditMessage(chatID, &tg.MessagesEditMessageRequest{
			Peer:        ctx.PeerStorage.GetInputPeerById(chatID),
			ID:          callbackQuery.MsgID,
			Message:     "Hi, send me any file to get a direct streamble link to that file.",
			ReplyMarkup: markup,
		})
		if err != nil {
			return err
		}

	case "dev_info":
		ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
			QueryID: callbackQuery.QueryID,
			Message: "",
		})

		_, err := ctx.EditMessage(chatID, &tg.MessagesEditMessageRequest{
			Peer:    ctx.PeerStorage.GetInputPeerById(chatID),
			ID:      callbackQuery.MsgID,
			Message: "This bot developed by @Kaliboy002",
		})
		if err != nil {
			return err
		}
	}

	return dispatcher.EndGroups
}
