package commands

import (
	"EverythingSuckz/fsb/config"
	"EverythingSuckz/fsb/internal/utils"

	"github.com/celestix/gotgproto/dispatcher"
	"github.com/celestix/gotgproto/dispatcher/handlers"
	"github.com/celestix/gotgproto/ext"
	"github.com/celestix/gotgproto/storage"
	"github.com/gotd/td/telegram/message/keyboard"
	"github.com/gotd/td/telegram/message/styling"
)

// The callback data for the "Help" button.
const helpCallbackData = "help_button_callback"

// LoadStart loads the command and callback handlers for the /start command.
func (m *command) LoadStart(dispatcher dispatcher.Dispatcher) {
	log := m.log.Named("start")
	defer log.Sugar().Info("Loaded")

	// Handler for the /start command.
	dispatcher.AddHandler(handlers.NewCommand("start", start))

	// Handler for the callback query from the inline button.
	dispatcher.AddHandler(handlers.NewCallback(helpCallbackData, helpCallback))
}

// start handles the /start command, sending the welcome message and inline button.
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

	// Create the inline keyboard button.
	// The button's text is "Help", and it sends a callback with the data "help_button_callback".
	markup := keyboard.Inline(
		keyboard.Row(
			keyboard.Callback("Help", []byte(helpCallbackData)),
		),
	)

	// Send the message with the inline button.
	ctx.Reply(
		u,
		"Hi, send me any file to get a direct streamble link to that file.",
		&ext.ReplyOpts{
			ReplyMarkup: markup,
		},
	)

	return dispatcher.EndGroups
}

// helpCallback handles the callback from the "Help" button, sending the detailed info.
func helpCallback(ctx *ext.Context, u *ext.Update) error {
	// Answer the callback query to show a new message.
	_, err := ctx.Reply(
		u,
		"This bot is used to change the files or direct download streaming link with no limitations.",
		&ext.ReplyOpts{
			// You can use a styling option like Bold to make the text stand out.
			ParseMode: styling.Plain,
		},
	)
	
	// This will show a small notification pop-up to the user confirming the action.
	ctx.AnswerCallbackQuery(u.Update.GetCallbackQuery().ID)

	return err
}
