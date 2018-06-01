package budget

import "time"

var storage Storage = nil

type Storage interface {
    AddActualTransaction(w Wallet, val ActualTransaction) error
    AddRegularChange(w Wallet, val RegularTransaction) error
    // GetActualTransactions(w Wallet, t1, t2 time.Date) ([]ActualTransaction, error)
    GetMonthlyIncome(w Wallet) (int, error)
    GetMonthlyIncomeTillDate(w Wallet, t time.Time) (int, error)
    GetMonthlyExpenseTillDate(w Wallet, t time.Time) (int, error)

    GetWalletForOwner(ownerId OwnerId) (*Wallet, error)

    CreateWalletOwner(ownerId OwnerId) error

    GetAllOwners() (map[OwnerId]OwnerData, error)
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
