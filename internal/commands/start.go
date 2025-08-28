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

func showWelcomeMessage(ctx *ext.Context, u *ext.Update) {
        // Create inline keyboard with Dev button (same as before)
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
}

func checkChannelMembership(ctx *ext.Context, userId int64, channelUsername string) bool {
        // Get channel info
        channelPeer := &tg.InputPeerUsername{Username: channelUsername}
        
        // Check if user is member of the channel
        req := &tg.ChannelsGetParticipantRequest{
                Channel: channelPeer,
                UserID:  &tg.InputUserID{UserID: userId},
        }
        
        _, err := ctx.Raw().ChannelsGetParticipant(ctx, req)
        if err != nil {
                // User is not a member
                return false
        }
        
        // User is a member
        return true
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
                // Check if user is member of the channel
                isMember := checkChannelMembership(ctx, userId, "KaIi_Bots")
                
                if !isMember {
                        // User is not a member - show alert
                        ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
                                QueryID: callbackQuery.QueryID,
                                Message: "❌ You are not joined to our channel. Please join first!",
                                Alert:   true,
                        })
                } else {
                        // User is a member - answer callback and show welcome message
                        ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
                                QueryID: callbackQuery.QueryID,
                                Message: "",
                        })
                        
                        // Show welcome message with Dev button
                        showWelcomeMessage(ctx, u)
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
