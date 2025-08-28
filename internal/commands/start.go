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
	// Register callback handler for the "Dev" button
	dispatcher.AddHandler(handlers.NewCallback(devCallback, "dev_info"))
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

	// Define the inline keyboard using gotd/td/tg types
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

	// Send the message with the inline button
	ctx.Reply(u, "Hi, send me any file to get a direct streamble link to that file.", &ext.ReplyOpts{
		Markup: markup,
	})

	return dispatcher.EndGroups
}

func devCallback(ctx *ext.Context, u *ext.Update) error {
	cb := u.CallbackQuery
	if cb == nil || string(cb.Data) != "dev_info" {
		return dispatcher.EndGroups
	}
	// Always answer the callback to avoid "loading" spinner in Telegram
	ctx.AnswerCallback(cb, "")
	// Send the developer info as a reply to the callback
	ctx.SendMessage(cb.Peer, "This bot Developed by @Kaliboy002", nil)
	return dispatcher.EndGroups
}
