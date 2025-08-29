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
        
        // Get the message ID of the message to be edited from the callback query
        messageID := callbackQuery.Message.GetID()
        
        switch callbackData {
        case "check_membership":
                // Answer the callback query with empty message (no popup)
                ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
                        QueryID: callbackQuery.QueryID,
                        Message: "",
                })
                
                // Edit the existing message instead of sending a new one
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
                
                ctx.EditMessage(userId, messageID, &tg.MessagesEditMessageRequest{
                        Peer:         ctx.PeerStorage.GetInputPeerById(userId),
                        Message:      "Hi, send me any file to get a direct streamble link to that file.",
                        ReplyMarkup:  markup,
                })
                
        case "dev_info":
                // Answer the callback query first (required)
                ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
                        QueryID: callbackQuery.QueryID,
                        Message: "",
                })
                
                // Edit the existing message with new content
                ctx.EditMessage(userId, messageID, &tg.MessagesEditMessageRequest{
                        Peer:    ctx.PeerStorage.GetInputPeerById(userId),
                        Message: "This bot developed by @Kaliboy002",
                })
        }
        
        return dispatcher.EndGroups
}
