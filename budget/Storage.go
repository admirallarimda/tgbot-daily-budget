package budget

import "time"

var storage Storage = nil

type Storage interface {
	GetWalletForOwner(ownerId OwnerId, createIfAbsent bool) (*Wallet, error)
	CreateWalletOwner(ownerId OwnerId) (*Wallet, error)
	GetAllOwners() (map[OwnerId]OwnerData, error)

	GetOwnerDailyNotificationTime(id OwnerId) (*time.Duration, error)
	SetOwnerDailyNotificationTime(id OwnerId, notifTime *time.Duration) error

	SetWalletInfo(w WalletId, monthStart int) error

	AddActualTransaction(w WalletId, val ActualTransaction) error
	GetActualTransactions(w WalletId, tMin, tMax time.Time) ([]ActualTransaction, error)

	AddRegularTransaction(w WalletId, val RegularTransaction) error
	GetRegularTransactions(w WalletId) ([]RegularTransaction, error)
	RemoveRegularTransaction(w WalletId, t RegularTransaction) error
}
