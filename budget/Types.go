package budget

import "time"

type ActualTransaction struct {
    Value int
    Time time.Time
    Label string
    RawText string // raw text - might be needed, but not necessary
}

func NewActualTransaction(value int, t time.Time, label, raw string) *ActualTransaction {
    amount := &ActualTransaction{ Value: value,
                             Time: t,
                             Label: label,
                             RawText: raw}
    return amount
}

type RegularTransaction struct {
    Value, Date int
    Label string
}

func NewRegularTransaction(value, date int, label string) *RegularTransaction {
    if date < 1 || date > 28 {
        panic("Date for monthly change is out of borders")
    }
    transaction := &RegularTransaction{ Value: value,
                                        Date: date,
                                        Label: label}
    return transaction
}

type OwnerId int64
type OwnerData struct {
    WalletId *string    `wallet`
    //Timezone *string  `tz`
    DailyReminderTime *time.Duration `dailyNotifTime` // from UTC midnight
    RegularTxs map[int][]RegularTransaction           // map 'dayOfMonth -> slice of RegularTransaction' used for reminding
}
