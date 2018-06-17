package budget

import "log"
import "fmt"
import "time"
import "errors"

type WalletId string

type Wallet struct {
    ID WalletId
    MonthStart int
    storage Storage
}

func NewWalletFromStorage(id string, monthStart int, storageconn Storage) *Wallet {
    wallet := &Wallet { ID: WalletId(id),
                        MonthStart: monthStart,
                        storage: storageconn}
    return wallet
}

func (w *Wallet) AddTransaction(t ActualTransaction) error {
    return w.storage.AddActualTransaction(w.ID, t)
}

func checkRegularTransactionLabelExist(transactions []*RegularTransaction, label string) bool {
    for _, t := range transactions {
        if t.Label == label {
            return true
        }
    }
    return false
}

func (w *Wallet) AddRegularTransaction(t RegularTransaction) error {
    date := t.Date
    if date < 1 || date > 28 {
        return errors.New("Only dates between 1 and 28 are allowed for regular income/outcome setting")
    }

    transactions, err := w.storage.GetRegularTransactions(w.ID)
    if err != nil {
        log.Printf("Could not add regular transactions - unable to get a list of all current regulars for wallet '%s'; error: %s", w.ID, err)
        return err
    }

    exists := checkRegularTransactionLabelExist(transactions, t.Label)
    if exists {
        log.Printf("Label '%s' already exists for wallet '%s', cannot add regular transaction", t.Label, w.ID)
        return errors.New(fmt.Sprintf("Label '%s' already exists", t.Label))
    }

    return w.storage.AddRegularTransaction(w.ID, t)
}

func checkRegularTransactionExactMatchExist(transactions []*RegularTransaction, t_checked RegularTransaction) bool {
    for _, t := range transactions {
        if *t == t_checked {
            return true
        }
    }
    return false
}

func (w *Wallet) RemoveRegularTransaction(t RegularTransaction) error {
    transactions, err := w.storage.GetRegularTransactions(w.ID)
    if err != nil {
        log.Printf("Could not remove regular transactions - unable to get a list of all current regulars for wallet '%s'; error: %s", w.ID, err)
        return err
    }

    exists := checkRegularTransactionExactMatchExist(transactions, t)
    if !exists {
        log.Printf("There are no exactly matched regular transaction for wallet '%s', cannot remove regular transaction", w.ID)
        return errors.New(fmt.Sprintf("Label '%s' already exists", t.Label))
    }

    return w.storage.RemoveRegularTransaction(w.ID, t)
}

func (w *Wallet) GetPlannedMonthlyIncome() (int, error) {
    log.Printf("Calculating planned monthly income for wallet '%s'", w.ID)
    transactions, err := w.storage.GetRegularTransactions(w.ID)
    if err != nil {
        log.Printf("Could not get monthly transactions for wallet '%s', error: %s", w.ID, err)
        return 0, err
    }

    totalIncome := 0
    for _, change := range transactions {
        totalIncome += change.Value
    }

    log.Printf("Total income for wallet '%s' is %d", w.ID, totalIncome)

    return totalIncome, nil
}


func calcCurMonthBorders(walletMonthStartDay int, now time.Time) (time.Time, time.Time) {
    if walletMonthStartDay < 1 || walletMonthStartDay > 28 {
        panic("Date must be between 1 and 28")
    }

    monthStart := time.Date(now.Year(), now.Month(), walletMonthStartDay, 0, 0, 0, 0, time.Local) // TODO: check whether UTC or Local is needed
    if now.Day() < walletMonthStartDay {
        // we've switched the month already, monthStart should be at the previous month
        monthStart = monthStart.AddDate(0, -1, 0)
    }
    monthEnd := monthStart.AddDate(0, 1, 0)
    log.Printf("Month borders are from %s to %s", monthStart, monthEnd)
    return monthStart, monthEnd
}

func (w *Wallet) GetActualMonthlyIncome() (int, error) {
    log.Printf("Calculating monthly income for wallet '%s'", w.ID)

    regularTransactions, err := w.storage.GetRegularTransactions(w.ID)
    if err != nil {
        log.Printf("Could not get regular transactions for wallet '%s', error: %s", w.ID, err)
        return 0, err
    }
    regularTransactionsLabeled := make(map[string]int, len(regularTransactions))
    for _, regularElem := range regularTransactions {
        if regularElem.Label == "" {
            panic("Label for regular transaction is empty")
        }
        regularTransactionsLabeled[regularElem.Label] = regularElem.Value
    }

    log.Printf("Month start for wallet '%s' is %d", w.ID, w.MonthStart)

    t1, t2 := calcCurMonthBorders(w.MonthStart, time.Now())
    transactions, err := w.storage.GetActualTransactions(w.ID, t1, t2)
    if err != nil {
        log.Printf("Could not receive transactions due to error: %s", err)
        return 0, err
    }

    monthly := 0
    regularReplacedWithActual := make(map[string]bool, 0)
    for _, transaction := range transactions {
        label := transaction.Label
        value := transaction.Value
        if regularValue, regularFound := regularTransactionsLabeled[label]; regularFound {
            log.Printf("Found a transaction labeled '%s' which replaces value %d -> %d (in addition to other same labeled transactions)", label, regularValue, value)
            monthly += value
            regularReplacedWithActual[label] = true
        } else if value > 0 {
            log.Printf("Found a transaction labeled '%s' with positive income %d, using it for monthly income calculation", label, value)
            monthly += value
        } // else value <= 0 { log.Printf("Skipping transaction labeled '%s' with negative value %d", label, value) }
    }

    // adding those regular transactions which were not yet matches with an actial one
    for label, value := range regularTransactionsLabeled {
        if _, found := regularReplacedWithActual[label]; found {
            log.Printf("Regular transaction labeled '%s' with value %d has already been matched by an actual transaction, skipping", label, value)
            continue
        }
        log.Printf("Regular transaction labeled '%s' with value %d will be used for monthly income calculation", label, value)
        monthly += value
    }
    log.Printf("Final montly income is: %d", monthly)

    return monthly, nil
}

func (w *Wallet) GetActualMonthlyIncomeTillDate(t time.Time) (int, error) {
    log.Printf("Calculating monthly income for wallet '%s' till time %s", w.ID, t)

    monthly, err := w.GetActualMonthlyIncome()
    if err != nil {
        log.Printf("Could not calculate monthly income for wallet '%s' due to error: %s", w.ID, err)
        return 0, err
    }

    var result float32 = 0
    // calculating result based on hoe many days have passed considering whether we've reached the end of prev month
    curDay := t.Day()
    if curDay >= w.MonthStart {
        daysInCurMonth := daysInMonth[t.Month()]
        result = float32(monthly) / float32(daysInCurMonth) * float32((curDay - w.MonthStart + 1)) // +1 as we assume that daily portion is granted at the beginning of the day
    } else {
        // tricky code to calc how many days have passed if we've reached the end of the previous month
        prevMonth := time.December
        if t.Month() != time.January {
            prevMonth = t.Month() - 1
        }
        result = float32(monthly) / float32(31 - (w.MonthStart - curDay) - (31 - daysInMonth[prevMonth]))
    }

    log.Printf("Calculated montly income till date %s: it equals to %f", t, result)
    return int(result), nil
}


func (w *Wallet) GetMonthlyExpenseTillDate(t time.Time) (int, error) {
    log.Printf("Getting monthly expenses for wallet %s till date %s", w.ID, t)

    // TODO: cache results of actual transactions so we don't need to call it again
    t1, t2 := calcCurMonthBorders(w.MonthStart, t)
    transactions, err := w.storage.GetActualTransactions(w.ID, t1, t)
    if err != nil {
        log.Printf("Could not receive transactions due to error: %s", err)
        return 0, err
    }

    var totalExpense int = 0
    for _, t := range transactions {
        if t.Value < 0 {
            totalExpense += t.Value
        }
    }

    log.Printf("Calculated total expense %d for wallet '%s' from %s till %s", totalExpense, w.ID, t1, t2)
    return -int(totalExpense), nil // returning positive value
}

func (w * Wallet) GetBalance(t time.Time) (int, error) {
    log.Printf("Starting to calculate available amount for wallet '%s' for time %s", w.ID, t)

    // getting current available money
    curAvailIncome, err := w.GetActualMonthlyIncomeTillDate(t)
    if err != nil {
        log.Printf("Unable to get current available amount due to error: %s", err)
        return 0, err
    }

    curExpenses, err := w.GetMonthlyExpenseTillDate(t)
    if err != nil {
        log.Printf("Unable to get current expenses due to error: %s", err)
        return 0, err
    }
    availMoney := curAvailIncome - curExpenses
    log.Printf("Currently available money for wallet '%s': %d (income: %d; expenses: %d)", w.ID, availMoney, curAvailIncome, curExpenses)
    return availMoney, nil
}

func (w *Wallet) SetMonthStart(date int) error {
    if date < 1 || date > 28 {
        panic("Date is out of range")
    }
    oldDate := w.MonthStart
    w.MonthStart = date
    err := w.storage.SetWalletInfo(w.ID, w.MonthStart)
    if err != nil {
        log.Printf("Could not update wallet '%s' date from %d to %d. Reverting to oiginal value", w.ID, oldDate, date)
        w.MonthStart = oldDate
    }
    return err
}
