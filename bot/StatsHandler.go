package bot

import (
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/admirallarimda/tgbot-daily-budget/budget"
	"github.com/admirallarimda/tgbotbase"
	"gopkg.in/telegram-bot-api.v4"
)

type statsHandler struct {
	baseHandler
}

func NewStatsHandler(storage budget.Storage) tgbotbase.IncomingMessageHandler {
	h := &statsHandler{}
	h.storage = storage
	return h
}

func (h *statsHandler) Init(outMsgCh chan<- tgbotapi.Chattable, srvCh chan<- tgbotbase.ServiceMsg) tgbotbase.HandlerTrigger {
	h.OutMsgCh = outMsgCh
	return tgbotbase.NewHandlerTrigger(nil, []string{"stats"})
}

func (h *statsHandler) Name() string {
	return "statistics"
}

func (h *statsHandler) HandleOne(msg tgbotapi.Message) {
	owner := budget.OwnerId(msg.Chat.ID)
	wallet, err := budget.GetWalletForOwner(owner, false, h.storage)
	if err != nil {
		log.Printf("Wallet is absent during stats preparation for %s due to error: %s", dumpMsgUserInfo(msg), err)
		h.OutMsgCh <- tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("There is no wallet - stats cannot be obtained"))
		return
	}
	reply, err := prepareMonthlySummary(owner, wallet, time.Now())
	if err != nil {
		log.Printf("Could not prepare monthly stats for %s due to error: %s", dumpMsgUserInfo(msg), err)
		h.OutMsgCh <- tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("Thre is a problem with stats preparation"))
		return
	}
	h.OutMsgCh <- tgbotapi.NewMessage(msg.Chat.ID, reply)
}

func prepareMonthlySummary(owner budget.OwnerId, wallet *budget.Wallet, t time.Time) (string, error) {
	log.Printf("Preparing monthly stats to owner %d with wallet '%s'", owner, wallet.ID)
	summary, err := wallet.GetMonthlySummary(t)
	if err != nil {
		return "", err
	}

	type keyValue struct {
		key   string
		value int
	}
	var sortedExpenses []keyValue
	for k, v := range summary.ExpenseSummary {
		sortedExpenses = append(sortedExpenses, keyValue{key: k, value: v})
	}
	sort.Slice(sortedExpenses, func(i, j int) bool {
		return sortedExpenses[i].value < sortedExpenses[j].value // lowest value will be the first
	})

	msg := fmt.Sprintf("Last month summary (for dates from %s to %s):", summary.TimeStart, summary.TimeEnd)
	for _, kv := range sortedExpenses {
		label_txt := "unlabeled category"
		if kv.key != "" {
			label_txt = fmt.Sprintf("category labeled '%s'", kv.key)
		}
		msg = fmt.Sprintf("%s\nSpent %d for %s", msg, -(kv.value), label_txt)
	}

	return msg, nil
}
