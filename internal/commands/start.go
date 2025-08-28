package commands

import (
	"EverythingSuckz/fsb/config"
	"EverythingSuckz/fsb/internal/utils"

	"github.com/celestix/gotgproto/dispatcher"
	"github.com/celestix/gotgproto/dispatcher/handlers"
	"github.com/celestix/gotgproto/ext"
	"github.com/celestix/gotgproto/storage"
	"github.com/gotd/td/tg"
)

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

	// Show mandatory channel join message
	showChannelJoinMessage(ctx, u)
	return dispatcher.EndGroups
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
	// Get the callback query data
	callbackQuery := u.CallbackQuery
	if callbackQuery == nil {
		return dispatcher.EndGroups
	}

	callbackData := string(callbackQuery.Data)
	userId := callbackQuery.UserID

	switch callbackData {
	case "check_membership":
		
		// Create a slice of InputPeer for the channel. You'll need the channel ID from your config.
		// Replace 'YOUR_CHANNEL_ID_HERE' with the actual channel ID.
		// For example, if your channel ID is stored in config.ValueOf.LogChannelID, use that.
		// Note: Channel IDs from the API are usually negative.
		channelPeer, err := ctx.ResolveUsername("KaIi_Bots") // Or use the channel ID from config.ValueOf.LogChannelID
		if err != nil {
			ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
				QueryID: callbackQuery.QueryID,
				Message: "Error: Could not resolve channel.",
				Alert:   true,
			})
			return dispatcher.EndGroups
		}

		// Check if the user is a member of the channel.
		// We'll use the GetParticipant method for this.
		// A common way to check this is to call the API method and handle the error case where the user is not found.
		isMember := false
		_, err = ctx.API().ChannelsGetParticipant(ctx.Context, &tg.ChannelsGetParticipantRequest{
			Channel: channelPeer.GetInputChannel(),
			UserID:  ctx.PeerStorage.GetInputPeerByID(userId),
		})

		// A more reliable check would be to see if the user is a member.
		// This can be done by checking the type of participant returned or checking for specific errors.
		// If err is nil, it means the user is a member.
		if err == nil {
			isMember = true
		} else {
			// Specific error for not a member
			var rpcErr tg.RPCError
			if ext.As(&rpcErr, err) && rpcErr.Message == "PEER_ID_INVALID" {
				// The user is not a participant.
				isMember = false
			} else {
				// Some other error occurred. You might want to log this.
				isMember = false
			}
		}

		if isMember {
			// Answer the callback query with an empty message (no popup)
			ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
				QueryID: callbackQuery.QueryID,
				Message: "",
			})

			// Send the welcome message with the Dev button
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

			ctx.SendMessage(userId, &tg.MessagesSendMessageRequest{
				Peer:        ctx.PeerStorage.GetInputPeerById(userId),
				Message:     "Hi, send me any file to get a direct streamable link to that file.",
				ReplyMarkup: markup,
			})
		} else {
			// User is not a member, show an alert.
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

		// Send new message with developer info
		ctx.SendMessage(userId, &tg.MessagesSendMessageRequest{
			Peer:    ctx.PeerStorage.GetInputPeerById(userId),
			Message: "This bot developed by @Kaliboy002",
		})
	}

	return dispatcher.EndGroups
}
