package budget

import "time"
import "log"
import "github.com/satori/go.uuid"

type AmountChange struct {
    Value int
    Time time.Time
}

func NewAmountChange(value int, t time.Time) *AmountChange {
    amount := &AmountChange{ Value: value,
                             Time: t}
    return amount
}

type MonthlyChange struct {
    Value, Date int
    Description string
}

func NewMonthlyChange(value, date int, desc string) *MonthlyChange {
    if date < 1 || date > 28 {
        panic("Date for monthly change is out of borders")
    }
    change := &MonthlyChange{ Value: value,
                              Date: date,
                              Description: desc}
    return change
}


type Wallet struct {
    ID uuid.UUID
}

func NewWallet() *Wallet {
    id, err := uuid.NewV4()
    if err != nil {
        log.Printf("Could not create new wallet due to error: %s", err)
        return nil
    }
    wallet := &Wallet{ ID: id }
    return wallet
}
