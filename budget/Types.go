package budget

import "time"
import "log"
import "github.com/satori/go.uuid"

type AmountChange struct {
    Value int
    Time time.Time
    Label string
    RawText string // raw text - might be needed, but not necessary
}

func NewAmountChange(value int, t time.Time, label, raw string) *AmountChange {
    amount := &AmountChange{ Value: value,
                             Time: t,
                             Label: label,
                             RawText: raw}
    return amount
}

type MonthlyChange struct {
    Value, Date int
    Label string
}

func NewMonthlyChange(value, date int, label string) *MonthlyChange {
    if date < 1 || date > 28 {
        panic("Date for monthly change is out of borders")
    }
    change := &MonthlyChange{ Value: value,
                              Date: date,
                              Label: label}
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

type OwnerId int64
type OwnerData struct {
    WalletId *string    `wallet`
    //Timezone *string  `tz`
    DailyReminderTime *time.Duration `dailyNotifTime` // from UTC midnight
}
