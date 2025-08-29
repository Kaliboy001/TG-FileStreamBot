package commands

import (
	"EverythingSuckz/fsb/config"
	"EverythingSuckz/fsb/internal/userdb" // Import the new userdb package
	"EverythingSuckz/fsb/internal/utils"
	"context" // This import is needed for context.Background() in API calls.
	"go.uber.org/zap"

	"github.com/celestix/gotgproto/dispatcher"
	"github.com/celestix/gotgproto/dispatcher/handlers"
	"github.com/celestix/gotgproto/ext"
	"github.com/celestix/gotgproto/storage"
	"github.com/gotd/td/rpc" // Import for rpc.Error and rpc.As
	"github.com/gotd/td/tg"
)

// The 'command' struct is assumed to be defined in another file,
// e.g., internal/commands/commands.go, and holds the logger instance.
type command struct {
	log *zap.Logger
}

// NewCommand is a constructor for the command struct.
func NewCommand(log *zap.Logger) *command {
	return &command{log: log}
}

// LoadStart registers the /start command and callback query handler.
func (m *command) LoadStart(dispatcher dispatcher.Dispatcher) {
	log := m.log.Named("start")
	defer log.Sugar().Info("Loaded")
	dispatcher.AddHandler(handlers.NewCommand("start", m.start))
	dispatcher.AddHandler(handlers.NewCallbackQuery(nil, m.handleCallbacks))
}

// start handles the /start command.
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

	// --- Save the user's ID to the SQLite database ---
	if err := userdb.SaveUser(m.log, chatId); err != nil {
		m.log.Error("Failed to save user ID to database", zap.Error(err))
	}
	// ---------------------------------------------------

	// Show mandatory channel join message
	m.showChannelJoinMessage(ctx, u)
	return dispatcher.EndGroups
}

// showChannelJoinMessage sends the initial message with channel join buttons.
func (m *command) showChannelJoinMessage(ctx *ext.Context, u *ext.Update) {
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

// handleCallbacks processes inline keyboard button presses.
func (m *command) handleCallbacks(ctx *ext.Context, u *ext.Update) error {
	callbackQuery := u.CallbackQuery
	if callbackQuery == nil {
		return dispatcher.EndGroups
	}

	callbackData := string(callbackQuery.Data)
	chatID := callbackQuery.UserID
	
	// Use the message ID from the callback query as it refers to the message that generated the callback.
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
		_, err = ctx.Client.API().ChannelsGetParticipant(context.Background(), &tg.ChannelsGetParticipantRequest{ // Corrected: ctx.Client.API()
			Channel:     channelPeer.GetInputChannel(),
			Participant: ctx.PeerStorage.GetInputPeerById(chatID),
		})

		var rpcErr rpc.Error
		if err != nil && rpc.As(&rpcErr, err) { // Corrected: rpc.As for error type assertion
			if rpcErr.Message == "PEER_ID_INVALID" {
				// The user is not a participant
				isMember = false
			} else {
				// Some other error occurred, log it for debugging
				m.log.Error("Error checking channel membership", zap.Error(err))
				isMember = false
			}
		}

		if isMember {
			// Answer the callback query with an empty message (no popup)
			ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
				QueryID: callbackQuery.QueryID,
				Message: "",
			})

			// Edit the existing message to show the welcome message and Dev button
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
				ID:          messageID,
				Peer:        ctx.PeerStorage.GetInputPeerById(chatID),
				Message:     "Hi, send me any file to get a direct streamble link to that file.",
				ReplyMarkup: markup,
			})
			if err != nil {
				m.log.Error("Failed to edit message with welcome", zap.Error(err))
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

		// Edit the existing message with developer info
		_, err := ctx.EditMessage(chatID, &tg.MessagesEditMessageRequest{
			ID:      messageID,
			Peer:    ctx.PeerStorage.GetInputPeerById(chatID),
			Message: "This bot developed by @Kaliboy002",
		})
		if err != nil {
			m.log.Error("Failed to edit message with dev info", zap.Error(err))
			return err
		}
	}

	return dispatcher.EndGroups
}
