package bot

import "gopkg.in/telegram-bot-api.v4"
import "regexp"

type serviceMsg struct {
    stopBot bool
}

type handlerTrigger struct {
    re *regexp.Regexp
    cmd string
}

type msgHandler interface {
    register(in_msg_chan <-chan tgbotapi.Message,
             out_msg_chan chan<- tgbotapi.MessageConfig,
             service_chan chan<- serviceMsg) handlerTrigger
    run() // to be called with 'go' statement
}

type baseHandler struct {
    in_msg_chan <-chan tgbotapi.Message
    out_msg_chan chan<- tgbotapi.MessageConfig
    service_chan chan<- serviceMsg
}
