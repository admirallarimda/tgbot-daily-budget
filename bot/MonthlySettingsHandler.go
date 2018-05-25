package bot

import "regexp"
import "log"
import "fmt"
import "strconv"
import "gopkg.in/telegram-bot-api.v4"

import "../budget"

var incomeRe *regexp.Regexp = regexp.MustCompile("income (\\d+)")
var expenseRe *regexp.Regexp = regexp.MustCompile("expense (\\d+)")
var dateRe *regexp.Regexp = regexp.MustCompile("date (\\d{1,2})")
var labelRe *regexp.Regexp = regexp.MustCompile("#([\\wA-Za-zА-Яа-я]+)")

const example = "/monthly income 2000 date 20 #label"

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
        labelMatches := labelRe.FindStringSubmatch(text)

        ownerId := budget.OwnerId(msg.Chat.ID)

        if len(dateMatches) == 0 {
            log.Printf("No date in message, cannot proceed")
            h.out_msg_chan<- tgbotapi.NewMessage(int64(ownerId), fmt.Sprintf("Date (from 1 to 28) is mandatory for monthly settings (example: %s)", example))
            continue
        }
        dateStr := dateMatches[1]
        date, err := strconv.Atoi(dateStr)
        if err != nil {
            log.Printf("Could not convert date %s to int", dateStr)
            h.out_msg_chan<- tgbotapi.NewMessage(int64(ownerId), fmt.Sprintf("Your date '%s' is not an integer number (but it should be)", dateStr))
            continue
        }
        if date < 1 || date > 28 {
            log.Printf("Incorrect date %d", date)
            h.out_msg_chan<- tgbotapi.NewMessage(int64(ownerId), fmt.Sprintf("Your date '%d' should be between 1 and 28. If your planned income/expense is at dates 29, 30 or 31 please use 1 or 28 (which is closer)", date))
            continue
        }
        log.Printf("Parsed date: %d", date)

        if len(labelMatches) == 0 {
            log.Printf("Labels are empty in text '%s'", text)
            h.out_msg_chan<- tgbotapi.NewMessage(int64(ownerId), fmt.Sprintf("Currently labels are mandatory for regular income/expenses, please follow the example: %s", example))
            continue
        }
        label := labelMatches[1]
        log.Printf("Parsed label: %s", label)

        // TODO: think of possibility to add multiple values (/monthly income 500 #salary1 date 5 income 200 #salary2 date 20 expense 10 expense 40 date 3)
        changes := make([]*budget.MonthlyChange, 0, len(incomeMatches) + len(expenseMatches))
        if len(incomeMatches) > 0 {
            valStr := incomeMatches[1]
            incomeVal, err := strconv.Atoi(valStr)
            if err != nil {
                log.Printf("Could not convert income value %s to int", valStr)
                h.out_msg_chan<- tgbotapi.NewMessage(int64(ownerId), fmt.Sprintf("Your income '%s' is not an integer number (but it should be)", valStr))
            } else {
                changes = append(changes, budget.NewMonthlyChange(incomeVal, date, label))
            }
        }

        // TODO: almost duplicate, can be merged to 1 function
        if len(expenseMatches) > 0 {
            valStr := expenseMatches[1]
            expenseVal, err := strconv.Atoi(valStr)
            if err != nil {
                log.Printf("Could not convert expense value %s to int", valStr)
                h.out_msg_chan<- tgbotapi.NewMessage(int64(ownerId), fmt.Sprintf("Your expense '%s' is not an integer number (but it should be)", valStr))
            } else {
                changes = append(changes, budget.NewMonthlyChange(-expenseVal, date, label))
            }
        }

        if len(changes) == 0 {
            log.Printf("No changes are going to be written after user command")
            h.out_msg_chan<- tgbotapi.NewMessage(int64(ownerId), fmt.Sprintf("No income or expense were found in the message (example: %s)", example))
            continue
        }

        w, err := budget.GetStorage().GetWalletForOwner(ownerId)
        if err != nil {
            log.Printf("Cannot get wallet for %s, error: %s", dumpMsgUserInfo(msg), err)
            h.out_msg_chan<- tgbotapi.NewMessage(int64(ownerId), "Cannot find your wallet. Have you entered /start ?")
            continue
        }

        for _, c := range(changes) {
            err = budget.GetStorage().AddRegularChange(*w, *c)
            if err != nil {
                log.Printf("Cannot add regular change for wallet %s of %s with error: %s", w.ID, dumpMsgUserInfo(msg), err)
                h.out_msg_chan<- tgbotapi.NewMessage(int64(ownerId), fmt.Sprintf("Something went wrong - cannot add planned income/expense. Please contact owner. Error: %s", err))
                // TODO: automessage to owner?
                continue
            }
        }

        monthlyIncome, err := budget.GetStorage().GetMonthlyIncome(*w)
        if err != nil {
            log.Printf("Could not receive monthly income for wallet %s of %s, error: %s", w.ID, dumpMsgUserInfo(msg), err)
            h.out_msg_chan<- tgbotapi.NewMessage(int64(ownerId), fmt.Sprintf("Something went wrong - cannot get your planned monthly income. Please contact owner. Error: %s", err))
            // TODO: automessage to owner?
            continue
        }

        log.Printf("Total monthly income for %s is %d", dumpMsgUserInfo(msg), monthlyIncome)
        h.out_msg_chan<- tgbotapi.NewMessage(int64(ownerId), fmt.Sprintf("Your planned monthly income is: %d", monthlyIncome))
        // TODO: current monthly settings
    }
}
