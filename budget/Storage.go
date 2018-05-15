package budget

var storage Storage = nil

type Storage interface {
    AddAmountChange(w Wallet, val AmountChange) error
    AddRegularChange(w Wallet, val, date int, description string) error
    // GetAmountChanges(w Wallet, t1, t2 time.Date) ([]AmountChange, error)
    GetMonthlyIncome(w Wallet) (int, error)

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
