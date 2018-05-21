package bot

import "regexp"
import "log"
import "strconv"
import "gopkg.in/telegram-bot-api.v4"

import "../budget"

var incomeRe *regexp.Regexp = regexp.MustCompile("income (\\d+)")
var expenseRe *regexp.Regexp = regexp.MustCompile("expense (\\d+)")
var dateRe *regexp.Regexp = regexp.MustCompile("date (\\d{1,2})")

type monthlyHandler struct {
    baseHandler
}

func (h *monthlyHandler) register(out_msg_chan chan<- tgbotapi.MessageConfig,
                                  service_chan chan<- serviceMsg) handlerTrigger {
    inCh := make(chan tgbotapi.Message, 0)
    h.in_msg_chan = inCh
    h.out_msg_chan = out_msg_chan

    return handlerTrigger{ cmd: "monthly",
                           in_msg_chan: inCh }
}

func (h *monthlyHandler) run() {
    for msg := range h.in_msg_chan {
        text := msg.Text

        incomeMatches := incomeRe.FindStringSubmatch(text) // TODO: FindAll?
        expenseMatches := expenseRe.FindStringSubmatch(text) // TODO: FindAll?
        dateMatches := dateRe.FindStringSubmatch(text)

        if len(dateMatches) == 0 {
            log.Printf("No date in message, cannot proceed")
            // TODO: reply to user
            continue
        }
        dateStr := dateMatches[1]
        date, err := strconv.Atoi(dateStr)
        if err != nil {
            log.Printf("Could not convert date %s to int", dateStr)
            // TODO: message to the user?
            continue
        }
        if date < 1 || date > 28 {
            log.Print("Incorrect date %d", date)
            // TODO: reply to user
            continue
        }

        changes := make([]*budget.MonthlyChange, 0, len(incomeMatches) + len(expenseMatches))
        if len(incomeMatches) > 0 {
            valStr := incomeMatches[1]
            incomeVal, err := strconv.Atoi(valStr)
            if err != nil {
                log.Printf("Could not convert income value %s to int", valStr)
                // TODO: indicate it to the user
            } else {
                changes = append(changes, budget.NewMonthlyChange(incomeVal, date, ""))
            }
        }

        // TODO: almost duplicate, can be merged to 1 function
        if len(expenseMatches) > 0 {
            valStr := expenseMatches[1]
            expenseVal, err := strconv.Atoi(valStr)
            if err != nil {
                log.Printf("Could not convert expense value %s to int", valStr)
                // TODO: indicate it to the user
            } else {
                changes = append(changes, budget.NewMonthlyChange(-expenseVal, date, ""))
            }
        }

        if len(changes) == 0 {
            log.Printf("No changes are going to be written after user command")
            // TODO: reply to user?
            continue
        }

        ownerId := msg.Chat.ID
        w, err := budget.GetStorage().GetWalletForOwner(ownerId)
        if err != nil {
            log.Printf("Cannot get wallet for %s, error: %s", dumpMsgUserInfo(msg), err)
            // TODO: reply to user? create automatically?
            continue
        }

        for _, c := range(changes) {
            err = budget.GetStorage().AddRegularChange(*w, *c)
            if err != nil {
                log.Printf("Cannot add regular change for wallet %s of %s with error: %s", w.ID, dumpMsgUserInfo(msg), err)
                continue
            }
        }

        monthlyIncome, err := budget.GetStorage().GetMonthlyIncome(*w)
        if err != nil {
            log.Printf("Could not receive monthly income for wallet %s of %s, error: %s", w.ID, dumpMsgUserInfo(msg), err)
        }

        log.Printf("Total monthly income for %s is %d", dumpMsgUserInfo(msg), monthlyIncome)

        // TODO: reply + current monthly settings
    }
}
