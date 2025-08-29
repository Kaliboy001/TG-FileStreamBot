package commands

import (
	"EverythingSuckz/fsb/config"
	"EverythingSuckz/fsb/internal/userdb" // Import the new userdb package
	"EverythingSuckz/fsb/internal/utils"
	"context" // This import might still be needed for context.Background() in API calls.
	"go.uber.org/zap"

	"github.com/celestix/gotgproto/dispatcher"
	"github.com/celestix/gotgproto/dispatcher/handlers"
	"github.com/celestix/gotgproto/ext"
	"github.com/celestix/gotgproto/storage"
	"github.com/gotd/td/rpc" // Keep if rpc.As is used in channel membership check
	"github.com/gotd/td/tg"
)

// Assuming 'command' struct is defined in another file, e.g., internal/commands/commands.go
// Example (DO NOT put this in start.go, it's just for reference):
// type command struct {
// 	log *zap.Logger
// }
// func NewCommand(log *zap.Logger) *command {
// 	return &command{log: log}
// }


func (m *command) LoadStart(dispatcher dispatcher.Dispatcher) {
	log := m.log.Named("start")
	defer log.Sugar().Info("Loaded")
	dispatcher.AddHandler(handlers.NewCommand("start", m.start)) // Use m.start
	dispatcher.AddHandler(handlers.NewCallbackQuery(nil, m.handleCallbacks)) // Use m.handleCallbacks
}

func (m *command) start(ctx *ext.Context, u *ext.Update) error {
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
	if err := userdb.SaveUser(m.log, chatId); err != nil { // Use m.log
		m.log.Error("Failed to save user ID to database", zap.Error(err)) // Use m.log
	}
	// -----------------------------------------------------------

	// Show mandatory channel join message
	m.showChannelJoinMessage(ctx, u) // Use m.showChannelJoinMessage
	return dispatcher.EndGroups
}

func (m *command) showChannelJoinMessage(ctx *ext.Context, u *ext.Update) {
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

func (m *command) handleCallbacks(ctx *ext.Context, u *ext.Update) error {
	callbackQuery := u.CallbackQuery
	if callbackQuery == nil {
		return dispatcher.EndGroups
	}

	callbackData := string(callbackQuery.Data)
	chatID := callbackQuery.UserID

	// Use callbackQuery.MsgID directly for editing the message.
	messageID := callbackQuery.MsgID

	switch callbackData {
	case "check_membership":
		// Resolve the channel using its username
		channelPeer, err := ctx.ResolveUsername("KaIi_Bots")
		if err != nil {
			ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
				QueryID: callbackQuery.QueryID,
				Message: "Error: Could not resolve channel.",
				Alert:   true,
			})
			return dispatcher.EndGroups
		}

		// Check if the user is a member of the channel
		isMember := true
		_, err = ctx.API().ChannelsGetParticipant(context.Background(), &tg.ChannelsGetParticipantRequest{
			Channel:     channelPeer.GetInputChannel(),
			Participant: ctx.PeerStorage.GetInputPeerById(chatID), // Use chatID consistently for user
		})

		var rpcErr rpc.Error
		if err != nil && rpc.As(&rpcErr, err) { // Use rpc.As for error type assertion
			if rpcErr.Message == "PEER_ID_INVALID" {
				// The user is not a participant
				isMember = false
			} else {
				// Some other error occurred, you might want to log this for debugging
				m.log.Error("Error checking channel membership", zap.Error(err)) // Use m.log
				isMember = false
			}
		}

		if isMember {
			// Answer the callback query with an empty message (no popup)
			ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
				QueryID: callbackQuery.QueryID,
				Message: "",
			})

			// Edit the existing message instead of sending a new one
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

			// Corrected EditMessage call using callbackQuery.MsgID
			_, err := ctx.EditMessage(chatID, &tg.MessagesEditMessageRequest{
				ID:          messageID,
				Peer:        ctx.PeerStorage.GetInputPeerById(chatID),
				Message:     "Hi, send me any file to get a direct streamble link to that file.",
				ReplyMarkup: markup,
			})
			if err != nil {
				m.log.Error("Failed to edit message with welcome", zap.Error(err)) // Use m.log
				return err
			}
		} else {
			// User is not a member, show an alert
			ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
				QueryID: callbackQuery.QueryID,
				Message: "⚠️ You are not a member of the channel. Please join first and then click the '🔐 Joined' button again.",
				Alert:   true,
			})
		}

	case "dev_info":
		// Answer the callback query first (required)
		ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
			QueryID: callbackQuery.QueryID,
			Message: "",
		})

		// Edit the existing message with new content
		_, err := ctx.EditMessage(chatID, &tg.MessagesEditMessageRequest{
			ID:      messageID,
			Peer:    ctx.PeerStorage.GetInputPeerById(chatID),
			Message: "This bot developed by @Kaliboy002",
		})
		if err != nil {
			m.log.Error("Failed to edit message with dev info", zap.Error(err)) // Use m.log
			return err
		}
	}

	return dispatcher.EndGroups
}
