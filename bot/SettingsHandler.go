package bot

import "regexp"
import "log"
import "fmt"
import "errors"
import "strconv"
import "gopkg.in/telegram-bot-api.v4"

import "../budget"

var monthStartRe *regexp.Regexp = regexp.MustCompile("monthStart (\\d{1,2})")

type settingsHandler struct {
    baseHandler
}

func (h *settingsHandler) register(out_msg_chan chan<- tgbotapi.MessageConfig,
                                   service_chan chan<- serviceMsg) handlerTrigger {
    inCh := make(chan tgbotapi.Message, 0)
    h.in_msg_chan = inCh
    h.out_msg_chan = out_msg_chan

    h.storageconn = budget.CreateStorageConnection()

    return handlerTrigger{ cmd: "settings",
                           in_msg_chan: inCh }
}

func (h *settingsHandler) setMonthStart(ownerId budget.OwnerId, date int) error {
    if date < 1 || date > 28 {
        return errors.New("Date must be between 1 and 28")
    }

    wallet, err := budget.GetWalletForOwner(ownerId, true, h.storageconn)
    if err != nil {
        return err
    }

    err = wallet.SetMonthStart(date)
    if err != nil {
        return err
    }

    log.Printf("Month start for wallet '%s' has been successfully modified to %d", wallet.ID, date)
    return nil
}

func (h *settingsHandler) run() {
    for msg := range h.in_msg_chan {
        text := msg.Text
        chatId := msg.Chat.ID
        ownerId := budget.OwnerId(chatId)
        if monthStartRe.MatchString(text) {
            matches := monthStartRe.FindStringSubmatch(text)
            dateStr := matches[1]
            date, err := strconv.Atoi(dateStr)
            if err != nil {
                log.Printf("Could not convert date '%s' for month start setting due to error: %s", dateStr, err)
                h.out_msg_chan<- tgbotapi.NewMessage(chatId, fmt.Sprintf("Incorrect format of date for month start modification"))
                continue
            }
            err = h.setMonthStart(ownerId, date)
            if err != nil {
                log.Printf("Could not set month start %d for owner %d due to error: %s", date, ownerId, err)
                h.out_msg_chan<- tgbotapi.NewMessage(chatId, fmt.Sprintf("Could not set month start due to the following reason: %s", err))
                continue
            }
            h.out_msg_chan<- tgbotapi.NewMessage(chatId, fmt.Sprintf("Month start has been successfully modified"))
        }
    }
}
