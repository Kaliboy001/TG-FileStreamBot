package commands

import (
	"EverythingSuckz/fsb/config"
	"EverythingSuckz/fsb/internal/utils"

	"github.com/celestix/gotgproto/dispatcher"
	"github.com/celestix/gotgproto/dispatcher/handlers"
	"github.com/celestix/gotgproto/ext"
	"github.com/celestix/gotgproto/storage"
	"github.com/celestix/gotgproto/mtproto"
	"github.com/celestix/gotgproto/mtproto/types"
)

// Register both command and callback handlers
func (m *command) LoadStart(dispatcher dispatcher.Dispatcher) {
	log := m.log.Named("start")
	defer log.Sugar().Info("Loaded")
	dispatcher.AddHandler(handlers.NewCommand("start", start))
	dispatcher.AddHandler(handlers.NewCallback(callbackDev, "dev_info"))
}

// /start command handler with inline button
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

	// Inline button markup
	keyboard := &types.ReplyMarkupInlineKeyboard{
		Rows: [][]*types.KeyboardButtonInline{
			{
				{
					// Text on the button and callback data
					Text:         "Dev",
					CallbackData: "dev_info",
				},
			},
		},
	}

	ctx.Reply(u, "Hi, send me any file to get a direct streamble link to that file.", &mtproto.ReplyMarkup{
		InlineKeyboard: keyboard,
	})
	return dispatcher.EndGroups
}

// Callback handler for "Dev" button
func callbackDev(ctx *ext.Context, u *ext.Update) error {
	cb := u.CallbackQuery()
	if cb == nil || cb.Data != "dev_info" {
		return nil
	}
	ctx.AnswerCallbackQuery(cb, nil)
	ctx.SendMessage(cb.Message.Peer, "This bot Developed by @Kaliboy002", nil)
	return nil
}
