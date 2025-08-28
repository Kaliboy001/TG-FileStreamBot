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
	dispatcher.AddHandler(handlers.NewCommand("start", startHandler))
	dispatcher.AddHandler(handlers.NewMessage(devButtonCallbackFilter, devButtonCallback))
}

// The filter must match func(u *ext.Update) bool
func devButtonCallbackFilter(u *ext.Update) bool {
	return u.CallbackQuery != nil && string(u.CallbackQuery.Data) == "dev_info"
}

// /start command handler: sends greeting and inline button
func startHandler(ctx *ext.Context, u *ext.Update) error {
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

// Handler for the callback query ("Dev" button) - sends a new message
func devButtonCallback(ctx *ext.Context, u *ext.Update) error {
	cb := u.CallbackQuery
	if cb == nil {
		return dispatcher.EndGroups
	}

	// Remove loading animation with correct request type
	req := &tg.MessagesSetBotCallbackAnswerRequest{
		QueryID: cb.QueryID,
	}
	_, _ = ctx.AnswerCallback(req)

	chatId := u.EffectiveChat().GetID()
	inputPeer := ctx.PeerStorage.GetInputPeerById(chatId)

	_, err := ctx.SendMessage(chatId, &tg.MessagesSendMessageRequest{
		Peer:    inputPeer,
		Message: "This bot Developed by @Kaliboy002",
	})
	return err
}
