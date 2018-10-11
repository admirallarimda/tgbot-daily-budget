package bot

import "fmt"
import "gopkg.in/telegram-bot-api.v4"

func dumpMsgUserInfo(msg tgbotapi.Message) string {
	return fmt.Sprintf("chat ID: %d (type '%s'), message issued by user ID: %d (username: '%s')", msg.Chat.ID,
		msg.Chat.Type,
		msg.From.ID,
		msg.From.UserName)
}

func uniqueInts(list []int) []int {
	result := make([]int, 0, len(list))
	seen := make(map[int]bool, len(list))
	for _, i := range list {
		if _, found := seen[i]; found {
			continue
		}
		result = append(result, i)
		seen[i] = true
	}
	return result
}
