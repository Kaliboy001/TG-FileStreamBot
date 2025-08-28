package commands

import (
	"EverythingSuckz/fsb/config"
	"EverythingSuckz/fsb/internal/utils"

	"github.com/celestix/gotgproto/dispatcher"
	"github.com/celestix/gotgproto/dispatcher/handlers"
	"github.com/celestix/gotgproto/ext"
	"github.com/celestix/gotgproto/storage"
)

func (m *command) LoadStart(dispatcher dispatcher.Dispatcher) {
	log := m.log.Named("start")
	defer log.Sugar().Info("Loaded")
	dispatcher.AddHandler(handlers.NewCommand("start", start))
	dispatcher.AddHandler(handlers.NewCallback(callbackDev, "dev_info"))
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

	// Use the correct inline keyboard types from your gotgproto version!
	keyboard := ext.NewInlineKeyboardMarkup(
		ext.NewInlineKeyboardRow(
			ext.NewInlineKeyboardButtonData("Dev", "dev_info"),
		),
	)

	ctx.Reply(u, "Hi, send me any file to get a direct streamble link to that file.", keyboard)
	return dispatcher.EndGroups
}

func callbackDev(ctx *ext.Context, u *ext.Update) error {
	cb := u.CallbackQuery()
	if cb == nil || cb.Data != "dev_info" {
		return nil
	}
	ctx.AnswerCallbackQuery(cb, nil)
	ctx.SendMessage(cb.Message.Peer, "This bot Developed by @Kaliboy002", nil)
	return nil
}
