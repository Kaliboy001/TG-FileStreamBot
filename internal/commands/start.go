package commands

import (
	"EverythingSuckz/fsb/config"
	"EverythingSuckz/fsb/internal/userdb" // Import the new userdb package
	"EverythingSuckz/fsb/internal/utils"
	"context"

	"github.com/celestix/gotgproto/dispatcher"
	"github.com/celestix/gotgproto/dispatcher/handlers"
	"github.com/celestix/gotgproto/ext"
	"github.com/celestix/gotgproto/storage"
	"github.com/gotd/td/rpc"
	"github.com/gotd/td/tg"
	"go.uber.org/zap" // Make sure zap is imported for logging
)

type command struct {
	log *zap.Logger
}

func NewCommand(log *zap.Logger) *command {
	return &command{log: log}
}

func (m *command) LoadStart(dispatcher dispatcher.Dispatcher) {
	log := m.log.Named("start")
	defer log.Sugar().Info("Loaded")
	dispatcher.AddHandler(handlers.NewCommand("start", start))
	dispatcher.AddHandler(handlers.NewCallbackQuery(nil, handleCallbacks))
}

func start(ctx *ext.Context, u *ext.Update) error {
	chatId := u.EffectiveChat().GetID()
	peerChatId := ctx.PeerStorage.GetPeerById(chatId)
	if peerChatId.Type != int(storage.TypeUser) {
		return dispatcher.EndGroups
	}
	if len(config.ValueOf.AllowedUsers) != 0 && !utils.Contains(config.ValueOf.AllowedUsers, chatId) {
		ctx.Reply(u, "You are not allowed to use this bot.", nil)
		return dispatcher.EndGroups
	}
	
	// --- ADD THIS LINE TO SAVE THE USER TO THE SQLITE DATABASE ---
	// This will save the user's ID to the separate SQLite database
	if err := userdb.SaveUser(ctx.Logger.Named("start_command"), chatId); err != nil {
		ctx.Logger.Error("Failed to save user ID to database", zap.Error(err))
		// Decide if you want to stop processing or just log the error
		// For now, we'll just log and continue to avoid disrupting the bot's core functionality
	}
	// -----------------------------------------------------------

	// Show mandatory channel join message
	showChannelJoinMessage(ctx, u)
	return dispatcher.EndGroups
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
	messageID := u.EffectiveMessage.GetID()

	switch callbackData {
	case "check_membership":
		// Answer the callback query with empty message (no popup)
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
			ID:          callbackQuery.MsgID, // Use MsgID from callbackQuery for editing the original message
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
			ID:      callbackQuery.MsgID, // Use MsgID from callbackQuery for editing the original message
			Message: "This bot developed by @Kaliboy002",
		})
		if err != nil {
			return err
		}
	}

	return dispatcher.EndGroups
}
