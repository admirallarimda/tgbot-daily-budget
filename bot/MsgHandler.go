package bot

import "gopkg.in/telegram-bot-api.v4"
import "regexp"
import "log"
import "github.com/admirallarimda/tgbot-daily-budget/budget"

type serviceMsg struct {
	stopBot bool
}

type handlerTrigger struct {
	re  *regexp.Regexp
	cmd string

	in_msg_chan chan<- tgbotapi.Message
}

func (h *handlerTrigger) Handle(msg tgbotapi.Message) bool {
	if h.re != nil && h.re.MatchString(msg.Text) {
		log.Printf("Message text '%s' matched regexp '%s', message will be sent to handler", msg.Text, h.re)
		h.in_msg_chan <- msg
		return true
	}
	if msg.IsCommand() && h.cmd == msg.Command() {
		log.Printf("Message text '%s' matched command '%s', message will be sent to handler", msg.Text, h.cmd)
		h.in_msg_chan <- msg
		return true
	}
	return false
}

type msgHandler interface {
	register(out_msg_chan chan<- tgbotapi.MessageConfig,
		service_chan chan<- serviceMsg) handlerTrigger
	run() // to be called with 'go' statement
}

type baseHandler struct {
	in_msg_chan  <-chan tgbotapi.Message
	out_msg_chan chan<- tgbotapi.MessageConfig
	service_chan chan<- serviceMsg

	storageconn budget.Storage
}
