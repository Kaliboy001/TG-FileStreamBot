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

// Register the start command and callback query handler
func (m *command) LoadStart(dispatcher dispatcher.Dispatcher) {
	log := m.log.Named("start")
	defer log.Sugar().Info("Loaded")
	dispatcher.AddHandler(handlers.NewCommand("start", start))
	dispatcher.AddHandler(handlers.NewMessage(callbackDevFilter, devCallback))
}

// Filter for callback query with data "dev_info"
func callbackDevFilter(u *ext.Update) bool {
	return u.CallbackQuery != nil && string(u.CallbackQuery.Data) == "dev_info"
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

	row := tg.KeyboardButtonRow{
		Buttons: []tg.KeyboardButtonClass{
			&tg.KeyboardButtonCallback{
				Text: "Dev",
				Data: []byte("dev_info"),
			},
		},
	}
	markup := &tg.ReplyInlineMarkup{
		Rows: []tg.KeyboardButtonRow{row},
	}

	ctx.Reply(u, "Hi, send me any file to get a direct streamble link to that file.", &ext.ReplyOpts{
		Markup: markup,
	})

	return dispatcher.EndGroups
}

// Handle the callback query for "Dev" button
func devCallback(ctx *ext.Context, u *ext.Update) error {
	cb := u.CallbackQuery
	if cb == nil {
		return dispatcher.EndGroups
	}
	// Answer callback to remove the "loading..." animation (with no popup)
	_ = ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
		QueryID: cb.QueryID,
	})

	// Send a new message in the chat with the developer info
	peer := cb.Peer
	_, err := ctx.SendMessage(peer, &tg.MessagesSendMessageRequest{
		Peer:    peer,
		Message: "This bot Developed by @Kaliboy002",
	})
	return err
}
