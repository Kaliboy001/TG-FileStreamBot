package commands

import (
	"EverythingSuckz/fsb/config"
	"EverythingSuckz/fsb/internal/utils"
	"fmt"
	"sync/atomic"

	"github.com/celestix/gotgproto/dispatcher"
	"github.com/celestix/gotgproto/dispatcher/handlers"
	"github.com/celestix/gotgproto/ext"
	"github.com/celestix/gotgproto/storage"
	"github.com/gotd/td/tg"
)

// Global counter for total users. Using an atomic counter for thread safety.
// Note: This counter will reset if the bot application restarts.
var totalUsers int64 = 0

// The Telegram User ID of the admin who will receive notifications.
const adminID int64 = 6070733162

func (m *command) LoadStart(dispatcher dispatcher.Dispatcher) {
	log := m.log.Named("start")
	defer log.Sugar().Info("Loaded")

	// We use a closure here to correctly pass the 'm' instance
	// into the handler function, which allows us to use 'm.log'.
	dispatcher.AddHandler(handlers.NewCommand("start", func(ctx *ext.Context, u *ext.Update) error {
		chatId := u.EffectiveChat().GetID()
		
		// This line from your original code is the correct way to get the peer type.
		peerChatId := ctx.PeerStorage.GetPeerById(chatId)
		if peerChatId.Type != int(storage.TypeUser) {
			return dispatcher.EndGroups
		}
		if len(config.ValueOf.AllowedUsers) != 0 && !utils.Contains(config.ValueOf.AllowedUsers, chatId) {
			ctx.Reply(u, "You are not allowed to use this bot.", nil)
			return dispatcher.EndGroups
		}
		
		// --- New Feature: Admin Notification for New Users ---
		
		// Atomically increment the total users count.
		newTotalUsers := atomic.AddInt64(&totalUsers, 1)

		// Correctly get the new user's username. Provide a fallback if it's not set.
		userUsername := "N/A"
		if u.EffectiveUser() != nil && u.EffectiveUser().Username != "" {
			userUsername = "@" + u.EffectiveUser().Username
		}

		// Format the notification message.
		notificationMessage := fmt.Sprintf(
			"➕ New User Notification ➕\n👤 User: %s\n🆔 User ID: %d\n📊 Total Users of Bot: %d",
			userUsername,
			chatId,
			newTotalUsers,
		)
		
		// Construct the tg.MessagesSendMessageRequest as required by ctx.SendMessage
		sendMessageRequest := &tg.MessagesSendMessageRequest{
			Peer:    ctx.PeerStorage.GetInputPeerById(adminID),
			Message: notificationMessage,
		}
		
		// Send the notification to the defined adminID.
		_, err := ctx.SendMessage(adminID, sendMessageRequest)
		if err != nil {
			// Corrected: Use m.log.Sugar().Errorf for printf-style error logging
			m.log.Sugar().Errorf("Failed to send new user notification to admin (%d): %v", adminID, err)
		}
		
		// --- End of New Feature ---
		
		// Show mandatory channel join message
		showChannelJoinMessage(ctx, u)
		return dispatcher.EndGroups
	}))

	dispatcher.AddHandler(handlers.NewCallbackQuery(nil, handleCallbacks))
}

func showChannelJoinMessage(ctx *ext.Context, u *ext.Update) {
	// Create inline keyboard with Join Channel and Joined buttons
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

		// Edit the same message
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

		// Edit again with developer info
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
