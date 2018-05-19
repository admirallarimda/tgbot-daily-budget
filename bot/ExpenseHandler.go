package bot

import "regexp"
import "log"
import "fmt"
import "time"
import "strconv"
import "gopkg.in/telegram-bot-api.v4"

import "../budget"

var re *regexp.Regexp = regexp.MustCompile("^-?(\\d+)$") // any number

type expenseHandler struct {
    baseHandler
}

func (h *expenseHandler) register(out_msg_chan chan<- tgbotapi.MessageConfig,
                                  service_chan chan<- serviceMsg) handlerTrigger {
    inCh := make(chan tgbotapi.Message, 0)
    h.in_msg_chan = inCh
    h.out_msg_chan = out_msg_chan

    return handlerTrigger{ re: re,
                           in_msg_chan: inCh }
}

func (h *expenseHandler) run() {
    for msg := range h.in_msg_chan {
        matches := re.FindStringSubmatch(msg.Text)
        if matches == nil {
            // assert
            log.Printf("No match of regexp, wrong handler!")
            continue
        }

        amountStr := matches[1]
        amount, err := strconv.Atoi(amountStr)
        if err != nil {
            // assert
            log.Printf("Amount %s cannot be converted to Int; error: %s", amountStr, err)
            continue
        }
        change := budget.NewAmountChange(-amount, time.Now()) // TODO: add correct handling for in/out (pos/neg) values here? or separate handler?
        userId := msg.From.ID
        wallet, err := budget.GetStorage().GetWalletForUser(userId)
        if err != nil {
            log.Printf("Could not get wallet for user %d with error: %s", userId, err)
            continue
        }

        err = budget.GetStorage().AddAmountChange(*wallet, *change)
        if err != nil {
            log.Printf("Could not add expence for user %d with wallet %s due to error: %s", userId, wallet.ID, err)
            continue
        }

        log.Printf("Expense of %d has been successfully added to wallet %s of user %d", change.Value, wallet.ID, userId)

        // getting current available money
        curAvailIncome, err := budget.GetStorage().GetMonthlyIncomeTillDate(*wallet, time.Now())
        if err != nil {
            log.Printf("Unable to get current available amount due to error: %s", err)
            h.out_msg_chan<- tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("Unable to get current available amount :("))
            continue
        }

        curExpenses, err := budget.GetStorage().GetMonthlyExpenseTillDate(*wallet, time.Now())
        if err != nil {
            log.Printf("Unable to get current expenses due to error: %s", err)
            h.out_msg_chan<- tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("Unable to get current expenses :("))
            continue
        }

        availMoney := curAvailIncome - curExpenses
        log.Printf("Currently available money: %d (income: %d; expenses: %d)", availMoney, curAvailIncome, curExpenses)
        h.out_msg_chan<- tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("Currently available money: %d", availMoney))
    }
}
