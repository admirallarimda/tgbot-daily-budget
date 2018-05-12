package bot

import "log"
import "gopkg.in/telegram-bot-api.v4"

import "../budget"

type startHandler struct {
    baseHandler
}

func (h *startHandler) register(in_msg_chan <-chan tgbotapi.Message,
                                out_msg_chan chan<- tgbotapi.MessageConfig,
                                service_chan chan<- serviceMsg) handlerTrigger {
    h.in_msg_chan = in_msg_chan
    h.out_msg_chan = out_msg_chan

    return handlerTrigger{ cmd: "start" }
}

func (h *startHandler) run() {
    for msg := range h.in_msg_chan {
        userId := msg.From.ID
        err:= budget.GetStorage().CreateUser(userId)
        if err != nil {
            log.Printf("Could not create user %d due to error: %s", userId, err)
            continue
        }

        log.Printf("User %d has been successfully added", userId)

        // TODO: reply + current available amount
    }
}
