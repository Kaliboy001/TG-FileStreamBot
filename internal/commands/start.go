package commands

import (
        "EverythingSuckz/fsb/config"
        "EverythingSuckz/fsb/internal/database"
        "EverythingSuckz/fsb/internal/utils"
        "fmt"

        "github.com/celestix/gotgproto/dispatcher" // This is the package
        "github.com/celestix/gotgproto/dispatcher/handlers"
        "github.com/celestix/gotgproto/ext"
        "github.com/celestix/gotgproto/storage"
        "github.com/gotd/td/tg"
)

// Note: User data is now stored in MongoDB instead of memory

// The Telegram User ID of the admin who will receive notifications.
const adminID int64 = 6070733162

// LoadStart initializes the handlers for the /start command and callbacks.
// The 'disp' parameter is the dispatcher instance.
func (m *command) LoadStart(disp dispatcher.Dispatcher) { // Renamed parameter to 'disp'
        log := m.log.Named("start")
        defer log.Sugar().Info("Loaded")

        // Add a handler for the /start command.
        // We use a closure here to correctly pass the 'm' instance (containing the logger)
        // into the handler function, which allows us to use 'm.log'.
        disp.AddHandler(handlers.NewCommand("start", func(ctx *ext.Context, u *ext.Update) error {
                chatId := u.EffectiveChat().GetID()
                
                // This line from your original code is the correct way to get the peer type.
                peerChatId := ctx.PeerStorage.GetPeerById(chatId)
                if peerChatId.Type != int(storage.TypeUser) {
                        // Now 'dispatcher.EndGroups' correctly refers to the package constant.
                        return dispatcher.EndGroups
                }
                if len(config.ValueOf.AllowedUsers) != 0 && !utils.Contains(config.ValueOf.AllowedUsers, chatId) {
                        ctx.Reply(u, "You are not allowed to use this bot.", nil)
                        return dispatcher.EndGroups
                }
                
                // --- New Feature: Admin Notification for New Users (MongoDB) ---
                
                // Check if this is a first-time user using MongoDB
                isUserSeen, err := database.DB.IsUserSeen(chatId)
                if err != nil {
                        m.log.Sugar().Errorf("Failed to check if user is seen: %v", err)
                        // Continue with bot flow even if database check fails
                } else if !isUserSeen {
                        // Get user's username with fallback
                        userUsername := "N/A"
                        if u.EffectiveUser() != nil && u.EffectiveUser().Username != "" {
                                userUsername = "@" + u.EffectiveUser().Username
                        }
                        
                        // Add new user to database
                        if err := database.DB.AddUser(chatId, userUsername); err != nil {
                                m.log.Sugar().Errorf("Failed to add user to database: %v", err)
                        }

                        // Get total user count from database
                        totalUsers, err := database.DB.GetTotalUserCount()
                        if err != nil {
                                m.log.Sugar().Errorf("Failed to get total user count: %v", err)
                                totalUsers = 0
                        }

                        // Format the notification message
                        notificationMessage := fmt.Sprintf(
                                "➕ New User Notification ➕\n👤 User: %s\n🆔 User ID: %d\n📊 Total Users of Bot: %d",
                                userUsername,
                                chatId,
                                totalUsers,
                        )
                        
                        // Construct the tg.MessagesSendMessageRequest as required by ctx.SendMessage
                        sendMessageRequest := &tg.MessagesSendMessageRequest{
                                Peer:    ctx.PeerStorage.GetInputPeerById(adminID),
                                Message: notificationMessage,
                        }
                        
                        // Send the notification to the defined adminID
                        _, err = ctx.SendMessage(adminID, sendMessageRequest)
                        if err != nil {
                                m.log.Sugar().Errorf("Failed to send new user notification to admin (%d): %v", adminID, err)
                        }
                } else {
                        // Update last seen for existing users
                        if err := database.DB.UpdateUserLastSeen(chatId); err != nil {
                                m.log.Sugar().Errorf("Failed to update user last seen: %v", err)
                        }
                }
                
                // --- End of New Feature ---
                
                // Show mandatory channel join message
                showChannelJoinMessage(ctx, u)
                return dispatcher.EndGroups
        }))

        // Add a handler for callback queries.
        disp.AddHandler(handlers.NewCallbackQuery(nil, handleCallbacks))
}

// showChannelJoinMessage sends a message prompting the user to join a channel.
func showChannelJoinMessage(ctx *ext.Context, u *ext.Update) {
        // Create inline keyboard with Join Channel and Joined buttons.
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

// handleCallbacks processes incoming callback queries.
func handleCallbacks(ctx *ext.Context, u *ext.Update) error {
        callbackQuery := u.CallbackQuery
        if callbackQuery == nil {
                // Now 'dispatcher.EndGroups' correctly refers to the package constant.
                return dispatcher.EndGroups
        }

        callbackData := string(callbackQuery.Data)
        chatID := callbackQuery.UserID

        switch callbackData {
        case "check_membership":
                ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
                        QueryID: callbackQuery.QueryID,
                        Message: "",
                })
                
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

                // Edit the same message to greet the user.
                _, err := ctx.EditMessage(chatID, &tg.MessagesEditMessageRequest{
                        Peer:        ctx.PeerStorage.GetInputPeerById(chatID),
                        ID:          callbackQuery.MsgID,
                        Message:     "Hi, send me any file to get a direct streamble link to that file.",
                        ReplyMarkup: markup,
                })
                if err != nil {
                        return err
                }

        case "dev_info":
                ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
                        QueryID: callbackQuery.QueryID,
                        Message: "",
                })

                // Edit again with developer information.
                _, err := ctx.EditMessage(chatID, &tg.MessagesEditMessageRequest{
                        Peer:    ctx.PeerStorage.GetInputPeerById(chatID),
                        ID:      callbackQuery.MsgID,
                        Message: "This bot developed by @Kaliboy002",
                })
                if err != nil {
                        return err
                }
        }

        return dispatcher.EndGroups
}
