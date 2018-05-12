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
