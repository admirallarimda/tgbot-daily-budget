package bot

import "log"
import "fmt"
import "gopkg.in/telegram-bot-api.v4"

import "github.com/admirallarimda/tgbot-daily-budget/budget"
import "github.com/admirallarimda/tgbotbase"

type startHandler struct {
	baseHandler
}

func NewStartHandler(storage budget.Storage) tgbotbase.IncomingMessageHandler {
	h := &startHandler{}
	h.storage = storage
	return h
}

func (h *startHandler) Init(outMsgCh chan<- tgbotapi.MessageConfig, srvCh chan<- tgbotbase.ServiceMsg) tgbotbase.HandlerTrigger {
	h.OutMsgCh = outMsgCh
	return tgbotbase.NewHandlerTrigger(nil, []string{"start"})
}

func (h *startHandler) Name() string {
	return "start"
}

func (h *startHandler) HandleOne(msg tgbotapi.Message) {
	ownerId := budget.OwnerId(msg.Chat.ID)

	wallet, err := budget.GetWalletForOwner(ownerId, true, h.storage)
	if err != nil {
		log.Printf("Could not create wallet owner for %s due to error: %s", dumpMsgUserInfo(msg), err)
		h.OutMsgCh <- tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("Something went wrong, could not initialize the wallet"))
		return
	}

	log.Printf("Wallet ID '%s' owner for %s has been successfully added", wallet.ID, dumpMsgUserInfo(msg))

	h.OutMsgCh <- tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("Wallet is ready for use"))
}
