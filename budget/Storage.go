package budget

import "time"

var storage Storage = nil

type Storage interface {
    AddAmountChange(w Wallet, val AmountChange) error
    AddRegularChange(w Wallet, val MonthlyChange) error
    // GetAmountChanges(w Wallet, t1, t2 time.Date) ([]AmountChange, error)
    GetMonthlyIncome(w Wallet) (int, error)
    GetMonthlyIncomeTillDate(w Wallet, t time.Time) (int, error)

    GetWalletForUser(userId int) (*Wallet, error)

    CreateUser(userId int) error


}

func GetStorage() Storage {
    if storage == nil {
        panic("storage is not yet set")
    }
    return storage
}

func SetStorage(s Storage) {
    if storage != nil {
        panic("storage has already been set")
    }
    storage = s
}
