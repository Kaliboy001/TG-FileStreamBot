package commands

import (
	"EverythingSuckz/fsb/config"
	"EverythingSuckz/fsb/internal/utils"
	"context" // Import the context package

	"github.com/celestix/gotgproto/dispatcher"
	"github.com/celestix/gotgproto/dispatcher/handlers"
	"github.com/celestix/gotgproto/ext"
	"github.com/celestix/gotgproto/storage"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/rpc" // Import the rpc package for rpc.Error
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
			Participant: ctx.PeerStorage.GetInputPeerById(userId),
		})
		
		var rpcErr rpc.Error
		if err != nil && rpc.As(&rpcErr, err) { // Use rpc.As instead of ext.As
			if rpcErr.Message == "PEER_ID_INVALID" {
				// The user is not a participant
				isMember = false
			} else {
				// Some other error occurred, you might want to log this for debugging
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
		
		// Send new message with developer info
		ctx.SendMessage(userId, &tg.MessagesSendMessageRequest{
			Peer:    ctx.PeerStorage.GetInputPeerById(userId),
			Message: "This bot developed by @Kaliboy002",
		})
	}
	
	return dispatcher.EndGroups
}
