package bot

import "regexp"
import "log"
import "fmt"
import "time"
import "strconv"
import "gopkg.in/telegram-bot-api.v4"

import "../budget"

var re *regexp.Regexp = regexp.MustCompile("^([+-]?)(\\d+) *(#([\\wa-zA-ZА-Яа-я]+))?$") // any number + label

type transactionHandler struct {
    baseHandler
}

func (h *transactionHandler) register(out_msg_chan chan<- tgbotapi.MessageConfig,
                                  service_chan chan<- serviceMsg) handlerTrigger {
    inCh := make(chan tgbotapi.Message, 0)
    h.in_msg_chan = inCh
    h.out_msg_chan = out_msg_chan

    h.storageconn = budget.CreateStorageConnection()

    return handlerTrigger{ re: re,
                           in_msg_chan: inCh }
}

func (h *transactionHandler) run() {
    for msg := range h.in_msg_chan {

        log.Printf("Message received from %s; text: %s", dumpMsgUserInfo(msg), msg.Text)

        matches := re.FindStringSubmatch(msg.Text)
        if matches == nil {
            // assert
            log.Printf("No match of regexp, wrong handler!")
            continue
        }
        log.Printf("Transaction regexp matched the following field: %s", matches)

        signStr := matches[1]
        sign := -1 // in most cases (no sign or '-' explicitly) we should pass negative number
        if signStr == "+" {
            sign = 1
        }
        amountStr := matches[2]
        amount, err := strconv.Atoi(amountStr)
        if err != nil {
            // assert
            log.Printf("Amount %s cannot be converted to Int; error: %s", amountStr, err)
            continue
        }

        label := ""
        if len(matches) > 4 {
            label = matches[4] // not 3 - using label without #
        }
        log.Printf("Message contains label '%s'", label)

        transaction := budget.NewActualTransaction(sign * amount, time.Now(), label, msg.Text)
        ownerId := budget.OwnerId(msg.Chat.ID)
        wallet, err := budget.GetWalletForOwner(ownerId, true, h.storageconn)
        if err != nil {
            log.Printf("Could not get wallet for %s with error: %s", dumpMsgUserInfo(msg), err)
            continue
        }

        matchesRegular, err := wallet.AddTransaction(*transaction)
        if err != nil {
            log.Printf("Could not add expence for %s with wallet %s due to error: %s", dumpMsgUserInfo(msg), wallet.ID, err)
            continue
        }

        log.Printf("Expense of %d has been successfully added to wallet %s for %s", transaction.Value, wallet.ID, dumpMsgUserInfo(msg))

        replyMsg := ""
        availMoney, err := wallet.GetBalance(time.Now())
        if err == nil {
            replyMsg = fmt.Sprintf("Currently available money: %d", availMoney)
        }

        if matchesRegular {
            replyMsg = fmt.Sprintf("%s\nYour recent transaction matches regular transaction, thus monthly income could be modified. Current values are: %s", replyMsg, constructIncomeMessage(wallet))
        }

        h.out_msg_chan<- tgbotapi.NewMessage(msg.Chat.ID, replyMsg)
    }
}
