package budget

import "time"

var storage Storage = nil

type Storage interface {
    GetWalletForOwner(ownerId OwnerId, createIfAbsent bool) (*Wallet, error)
    CreateWalletOwner(ownerId OwnerId) (*Wallet, error)

    GetRegularTransactions(w WalletId) ([]*RegularTransaction, error)
    GetActualTransactions(w WalletId, tMin, tMax time.Time) ([]*ActualTransaction, error)

    AddActualTransaction(w WalletId, val ActualTransaction) error
    AddRegularTransaction(w WalletId, val RegularTransaction) error

    GetAllOwners() (map[OwnerId]OwnerData, error)
    SetWalletInfo(w WalletId, monthStart int) error
}
