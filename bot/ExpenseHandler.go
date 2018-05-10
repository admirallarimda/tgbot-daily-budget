package bot

import "gopkg.in/telegram-bot-api.v4"
import "regexp"

var re *regexp.Regexp = regexp.MustCompile("-?(\\d+)") // any number

type expenseHandler struct {
    baseHandler
}

func (h *expenseHandler) register(in_msg_chan chan<- tgbotapi.Message,
                                  out_msg_chan <-chan tgbotapi.MessageConfig,
                                  service_chan <-chan serviceMsg) handlerTrigger {
    h.in_msg_chan = in_msg_chan
    h.out_msg_chan = out_msg_chan

    return handlerTrigger{ re: re }
}
