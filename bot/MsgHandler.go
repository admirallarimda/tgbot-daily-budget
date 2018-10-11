package bot

import "github.com/admirallarimda/tgbotbase"
import "github.com/admirallarimda/tgbot-daily-budget/budget"

type baseHandler struct {
	tgbotbase.BaseHandler
	storage budget.Storage
}
