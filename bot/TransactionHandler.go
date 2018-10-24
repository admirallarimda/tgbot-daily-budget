package bot

import "regexp"
import "log"
import "fmt"
import "time"
import "strconv"
import "gopkg.in/telegram-bot-api.v4"
import "github.com/admirallarimda/tgbotbase"

import "github.com/admirallarimda/tgbot-daily-budget/budget"

var re *regexp.Regexp = regexp.MustCompile("^([+-]?)(\\d+) *(#([\\wa-zA-ZА-Яа-я]+))?$") // any number + label

type transactionHandler struct {
	baseHandler
}

func NewTransactionHandler(storage budget.Storage) tgbotbase.IncomingMessageHandler {
	s := &transactionHandler{}
	s.storage = storage
	return s
}

func (h *transactionHandler) HandleOne(msg tgbotapi.Message) {
	log.Printf("Transaction: message received from %s; text: %s", dumpMsgUserInfo(msg), msg.Text)

	matches := re.FindStringSubmatch(msg.Text)
	if matches == nil {
		panic("Transaction: no match of regexp, wrong handler!")
	}
	log.Printf("Transaction: regexp matched the following field: %s", matches)

	signStr := matches[1]
	sign := -1 // in most cases (no sign or '-' explicitly) we should pass negative number
	if signStr == "+" {
		sign = 1
	}
	amountStr := matches[2]
	amount, err := strconv.Atoi(amountStr)
	if err != nil {
		fmt.Printf("Transaction: amount %s cannot be converted to Int; error: %s", amountStr, err)
		return
	}

	label := ""
	if len(matches) > 4 {
		label = matches[4] // not 3 - using label without #
	}
	log.Printf("Message contains label '%s'", label)

	transaction := budget.NewActualTransaction(sign*amount, time.Now(), label, msg.Text)
	ownerId := budget.OwnerId(msg.Chat.ID)
	wallet, err := budget.GetWalletForOwner(ownerId, true, h.storage)
	if err != nil {
		fmt.Printf("Could not get wallet for %s with error: %s", dumpMsgUserInfo(msg), err)
		return
	}

	matchesRegular, err := wallet.AddTransaction(*transaction)
	if err != nil {
		log.Printf("Could not add expence for %s with wallet %s due to error: %s", dumpMsgUserInfo(msg), wallet.ID, err)
		return
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

	h.OutMsgCh <- tgbotapi.NewMessage(msg.Chat.ID, replyMsg)
}

func (h *transactionHandler) Init(outMsgCh chan<- tgbotapi.Chattable, srvCh chan<- tgbotbase.ServiceMsg) tgbotbase.HandlerTrigger {
	h.OutMsgCh = outMsgCh
	return tgbotbase.NewHandlerTrigger(re, nil)
}

func (h *transactionHandler) Name() string {
	return "transaction"
}
