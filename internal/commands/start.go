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
        dispatcher.AddHandler(handlers.NewCallbackQuery(nil, handleDevCallback))
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
        
        // Create inline keyboard with Dev button
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
        
        ctx.Reply(u, "Hi, send me any file to get a direct streamble link to that file.", &ext.ReplyOpts{
                Markup: markup,
        })
        return dispatcher.EndGroups
}

func handleDevCallback(ctx *ext.Context, u *ext.Update) error {
        // Get the callback query data
        callbackQuery := u.CallbackQuery
        if callbackQuery == nil {
                return dispatcher.EndGroups
        }
        
        // Check if this is the "dev_info" callback
        if string(callbackQuery.Data) == "dev_info" {
                // Answer the callback query to remove the loading state
                ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
                        QueryID: callbackQuery.QueryID,
                        Message: "This bot developed by @Kaliboy002",
                })
        }
        
        return dispatcher.EndGroups
}
