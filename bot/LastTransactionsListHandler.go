package bot

import "regexp"
import "log"
import "fmt"
import "time"
import "strconv"
import "sort"
import "gopkg.in/telegram-bot-api.v4"

import "github.com/admirallarimda/tgbot-daily-budget/budget"
import "github.com/admirallarimda/tgbotbase"

var lastNRe *regexp.Regexp = regexp.MustCompile("(\\d+)")

type lastTransactionsListHandler struct {
	baseHandler
}

func NewLastTransactionsHandler(storage budget.Storage) tgbotbase.IncomingMessageHandler {
	h := &lastTransactionsListHandler{}
	h.storage = storage
	return h
}

func (h *lastTransactionsListHandler) Init(outMsgCh chan<- tgbotapi.Chattable, srvCh chan<- tgbotbase.ServiceMsg) tgbotbase.HandlerTrigger {
	h.OutMsgCh = outMsgCh
	return tgbotbase.NewHandlerTrigger(nil, []string{"last"})
}

func (h *lastTransactionsListHandler) Name() string {
	return "list of last transactions"
}

func (h *lastTransactionsListHandler) HandleOne(msg tgbotapi.Message) {
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

	wallet, err := budget.GetWalletForOwner(ownerId, true, h.storage)
	if err != nil {
		log.Printf("Could not get wallet for %s with error: %s", dumpMsgUserInfo(msg), err)
		h.OutMsgCh <- tgbotapi.NewMessage(chatId, "Could not obtain wallet for you:( Try to contact bot owner")
		return
	}

	// TODO: retrieve only necessary transactions for arbitrary amount of time
	now := time.Now()
	transactions, err := h.storage.GetActualTransactions(wallet.ID, now.Add(-time.Hour*24*30), now)
	if err != nil {
		log.Printf("Could not get transactions for wallet '%s' for %s with error: %s", wallet.ID, dumpMsgUserInfo(msg), err)
		h.OutMsgCh <- tgbotapi.NewMessage(chatId, "Could not obtain transactions in your wallet :( Try to contact bot owner")
		return
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
	h.OutMsgCh <- tgbotapi.NewMessage(chatId, result)
}
