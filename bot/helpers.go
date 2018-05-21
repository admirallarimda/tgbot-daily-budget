package bot

import "fmt"
import "gopkg.in/telegram-bot-api.v4"

func dumpMsgUserInfo(msg tgbotapi.Message) string {
    return fmt.Sprintf("chat ID: %d (type '%s'), message issued by user ID: %d (username: '%s')", msg.Chat.ID,
                                                                                                  msg.Chat.Type,
                                                                                                  msg.From.ID,
                                                                                                  msg.From.UserName)
}
