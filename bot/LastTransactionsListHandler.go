package bot

import "regexp"
import "log"
import "fmt"
import "time"
import "strconv"
import "sort"
import "gopkg.in/telegram-bot-api.v4"

import "github.com/admirallarimda/tgbot-daily-budget/budget"

var lastNRe *regexp.Regexp = regexp.MustCompile("(\\d+)")

type lastTransactionsListHandler struct {
	baseHandler
}

func (h *lastTransactionsListHandler) register(out_msg_chan chan<- tgbotapi.MessageConfig,
	service_chan chan<- serviceMsg) handlerTrigger {
	inCh := make(chan tgbotapi.Message, 0)
	h.in_msg_chan = inCh
	h.out_msg_chan = out_msg_chan

	h.storageconn = budget.CreateStorageConnection()

	return handlerTrigger{cmd: "last",
		in_msg_chan: inCh}
}

func (h *lastTransactionsListHandler) run() {
	for msg := range h.in_msg_chan {
		text := msg.Text
		chatId := msg.Chat.ID
		ownerId := budget.OwnerId(chatId)
		log.Printf("Last transactions request received from from %s; text: %s", dumpMsgUserInfo(msg), text)

		numberOfShownTransactions := 10
		matches := lastNRe.FindStringSubmatch(text)
		if matches != nil {
			var err error = nil
			numberOfShownTransactions, err = strconv.Atoi(matches[1])
			if err != nil {
				panic("Could not convert number of transactions, though regexp should handle it correctly")
			}
		}
		log.Printf("Going to show %d last transactions", numberOfShownTransactions)

		wallet, err := budget.GetWalletForOwner(ownerId, true, h.storageconn)
		if err != nil {
			log.Printf("Could not get wallet for %s with error: %s", dumpMsgUserInfo(msg), err)
			h.out_msg_chan <- tgbotapi.NewMessage(chatId, "Could not obtain wallet for you:( Try to contact bot owner")
			continue
		}

		// TODO: retrieve only necessary transactions for arbitrary amount of time
		now := time.Now()
		transactions, err := h.storageconn.GetActualTransactions(wallet.ID, now.Add(-time.Hour*24*30), now)
		if err != nil {
			log.Printf("Could not get transactions for wallet '%s' for %s with error: %s", wallet.ID, dumpMsgUserInfo(msg), err)
			h.out_msg_chan <- tgbotapi.NewMessage(chatId, "Could not obtain transactions in your wallet :( Try to contact bot owner")
			continue
		}

		sort.Slice(transactions, func(i, j int) bool { return transactions[i].Time.Before(transactions[j].Time) })
		if numberOfShownTransactions < len(transactions) {
			transactions = transactions[len(transactions)-numberOfShownTransactions:]
		}

		result := fmt.Sprintf("List of last %d transactions:\n", len(transactions))
		for _, t := range transactions {
			label := "--None--"
			if t.Label != "" {
				label = fmt.Sprintf("#%s", t.Label)
			}
			result += fmt.Sprintf("Date: %s; amount: %d; label: %s\n", t.Time.Format(time.RFC3339), t.Value, label)
		}
		h.out_msg_chan <- tgbotapi.NewMessage(chatId, result)
	}
}
