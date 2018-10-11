package bot

/*
import "log"
import "fmt"
import "gopkg.in/telegram-bot-api.v4"

import "github.com/admirallarimda/tgbot-daily-budget/budget"

type startHandler struct {
	baseHandler
}

func (h *startHandler) register(out_msg_chan chan<- tgbotapi.MessageConfig,
	service_chan chan<- serviceMsg) handlerTrigger {
	inCh := make(chan tgbotapi.Message, 0)
	h.in_msg_chan = inCh
	h.out_msg_chan = out_msg_chan

	h.storageconn = budget.CreateStorageConnection()

	return handlerTrigger{cmd: "start",
		in_msg_chan: inCh}
}

func (h *startHandler) run() {
	for msg := range h.in_msg_chan {

		ownerId := budget.OwnerId(msg.Chat.ID)

		wallet, err := budget.GetWalletForOwner(ownerId, true, h.storageconn)
		if err != nil {
			log.Printf("Could not create wallet owner for %s due to error: %s", dumpMsgUserInfo(msg), err)
			h.out_msg_chan <- tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("Something went wrong, could not initialize the wallet"))
			continue
		}

		log.Printf("Wallet ID '%s' owner for %s has been successfully added", wallet.ID, dumpMsgUserInfo(msg))

		h.out_msg_chan <- tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("Wallet is ready for use"))
	}
}
*/
