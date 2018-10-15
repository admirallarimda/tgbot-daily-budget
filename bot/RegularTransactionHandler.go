package bot

import "regexp"
import "log"
import "fmt"
import "strconv"
import "sort"
import "strings"
import "gopkg.in/telegram-bot-api.v4"

import "github.com/admirallarimda/tgbotbase"
import "github.com/admirallarimda/tgbot-daily-budget/budget"

var incomeRe *regexp.Regexp = regexp.MustCompile("income (\\d+)")
var expenseRe *regexp.Regexp = regexp.MustCompile("expense (\\d+)")
var dateRe *regexp.Regexp = regexp.MustCompile("date (\\d{1,2})")
var labelRe *regexp.Regexp = regexp.MustCompile("#([\\wA-Za-zА-Яа-я]+)")
var removeRe *regexp.Regexp = regexp.MustCompile("(remove|delete)")

const regularCmd = "regular"

const example = "/" + regularCmd + " income 2000 date 20 #label"

type regularTransactionHandler struct {
	baseHandler
}

func NewRegularTransactionHandler(storage budget.Storage) tgbotbase.IncomingMessageHandler {
	h := &regularTransactionHandler{}
	h.storage = storage
	return h
}

func (h *regularTransactionHandler) HandleOne(msg tgbotapi.Message) {
	log.Printf("Parsing regular command for %s with text '%s'", dumpMsgUserInfo(msg), msg.Text)
	text := strings.Trim(msg.Text, " /")
	chatId := msg.Chat.ID
	ownerId := budget.OwnerId(chatId)
	w, err := h.storage.GetWalletForOwner(ownerId, true)
	if err != nil {
		log.Printf("Cannot get wallet for %s, error: %s", dumpMsgUserInfo(msg), err)
		h.OutMsgCh <- tgbotapi.NewMessage(chatId, "Cannot find your wallet. Have you entered /start ?")
		return
	}
	if text == regularCmd {
		h.showSummary(w, chatId)
	} else {
		h.parseTransaction(w, chatId, text)
	}
}

func (h *regularTransactionHandler) Init(outMsgCh chan<- tgbotapi.MessageConfig, srvCh chan<- tgbotbase.ServiceMsg) tgbotbase.HandlerTrigger {
	h.OutMsgCh = outMsgCh
	return tgbotbase.NewHandlerTrigger(nil, []string{"regular"})
}

func (h *regularTransactionHandler) Name() string {
	return "regular transaction"
}

func (h *regularTransactionHandler) showSummary(w *budget.Wallet, chatId int64) {
	transactions, err := h.storage.GetRegularTransactions(w.ID)
	if err != nil {
		log.Printf("Could not get list of regular transactions for wallet '%s'", w.ID)
		h.OutMsgCh <- tgbotapi.NewMessage(chatId, fmt.Sprintf("Could not load list of regular transactions"))
		return
	}
	incomes := make(map[int][]budget.RegularTransaction, len(transactions))
	expences := make(map[int][]budget.RegularTransaction, len(transactions))
	dates := make([]int, 0, len(transactions))
	for _, t := range transactions {
		date := t.Date
		dates = append(dates, date)
		if t.Value > 0 {
			incomes[date] = append(incomes[date], t)
		} else {
			expences[date] = append(expences[date], t)
		}
	}
	dates = uniqueInts(dates)
	sort.Ints(dates)
	incomeText := "Incomes:\n"
	expenseText := "Expenses:\n"
	for _, d := range dates {
		incomeList, found := incomes[d]
		if found {
			for _, income := range incomeList {
				incomeText += fmt.Sprintf("Day %d: +%d #%s\n", d, income.Value, income.Label)
			}
		}
		expenseList, found := expences[d]
		if found {
			for _, expense := range expenseList {
				expenseText += fmt.Sprintf("Day %d: %d #%s\n", d, expense.Value, expense.Label)
			}
		}
	}

	result := "Summary of regular transactions:\n" + incomeText + "\n" + expenseText + "\n\n" + constructIncomeMessage(w)
	h.OutMsgCh <- tgbotapi.NewMessage(chatId, result)
}

func (h *regularTransactionHandler) parseTransaction(w *budget.Wallet, chatId int64, text string) {
	incomeMatches := incomeRe.FindStringSubmatch(text)   // TODO: FindAll?
	expenseMatches := expenseRe.FindStringSubmatch(text) // TODO: FindAll?
	dateMatches := dateRe.FindStringSubmatch(text)
	labelMatches := labelRe.FindStringSubmatch(text)
	toBeRemoved := removeRe.MatchString(text)

	if len(dateMatches) == 0 {
		log.Printf("No date in message, cannot proceed")
		h.OutMsgCh <- tgbotapi.NewMessage(chatId, fmt.Sprintf("Date (from 1 to 28) is mandatory for monthly settings (example: %s)", example))
		return
	}
	dateStr := dateMatches[1]
	date, err := strconv.Atoi(dateStr)
	if err != nil {
		log.Printf("Could not convert date %s to int", dateStr)
		h.OutMsgCh <- tgbotapi.NewMessage(chatId, fmt.Sprintf("Your date '%s' is not an integer number (but it should be)", dateStr))
		return
	}
	if date < 1 || date > 28 {
		log.Printf("Incorrect date %d", date)
		h.OutMsgCh <- tgbotapi.NewMessage(chatId, fmt.Sprintf("Your date '%d' should be between 1 and 28. If your planned income/expense is at dates 29, 30 or 31 please use 1 or 28 (which is closer)", date))
		return
	}
	log.Printf("Parsed date: %d", date)

	if len(labelMatches) == 0 {
		log.Printf("Labels are empty in text '%s'", text)
		h.OutMsgCh <- tgbotapi.NewMessage(chatId, fmt.Sprintf("Currently labels are mandatory for regular income/expenses, please follow the example: %s", example))
		return
	}
	label := labelMatches[1]
	log.Printf("Parsed label: %s", label)

	// TODO: think of possibility to add multiple values (/monthly income 500 #salary1 date 5 income 200 #salary2 date 20 expense 10 expense 40 date 3)
	transactions := make([]*budget.RegularTransaction, 0, len(incomeMatches)+len(expenseMatches))
	if len(incomeMatches) > 0 {
		valStr := incomeMatches[1]
		incomeVal, err := strconv.Atoi(valStr)
		if err != nil {
			log.Printf("Could not convert income value %s to int", valStr)
			h.OutMsgCh <- tgbotapi.NewMessage(chatId, fmt.Sprintf("Your income '%s' is not an integer number (but it should be)", valStr))
		} else {
			transactions = append(transactions, budget.NewRegularTransaction(incomeVal, date, label))
		}
	}

	// TODO: almost duplicate, can be merged to 1 function
	if len(expenseMatches) > 0 {
		valStr := expenseMatches[1]
		expenseVal, err := strconv.Atoi(valStr)
		if err != nil {
			log.Printf("Could not convert expense value %s to int", valStr)
			h.OutMsgCh <- tgbotapi.NewMessage(chatId, fmt.Sprintf("Your expense '%s' is not an integer number (but it should be)", valStr))
		} else {
			transactions = append(transactions, budget.NewRegularTransaction(-expenseVal, date, label))
		}
	}

	if len(transactions) == 0 {
		log.Printf("No transactions are going to be written after user command")
		h.OutMsgCh <- tgbotapi.NewMessage(chatId, fmt.Sprintf("No income or expense were found in the message (example: %s)", example))
		return
	}

	for _, c := range transactions {
		var err error = nil
		if toBeRemoved {
			err = w.RemoveRegularTransaction(*c)
		} else {
			err = w.AddRegularTransaction(*c)
		}

		if err != nil {
			log.Printf("Cannot process regular change for wallet %s of %d with error: %s", w.ID, chatId, err)
			h.OutMsgCh <- tgbotapi.NewMessage(chatId, fmt.Sprintf("Something went wrong - cannot process planned income/expense. Please contact owner. Error: %s", err))
			// TODO: automessage to owner?
			return
		}
	}

	h.OutMsgCh <- tgbotapi.NewMessage(chatId, constructIncomeMessage(w))
}
