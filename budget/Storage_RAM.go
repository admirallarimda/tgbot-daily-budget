package budget

import "time"
import "errors"
import "fmt"

type walletDetails struct {
    monthStart int
}

type ramStorage struct {
    walletTransactions map[WalletId][]*ActualTransaction
    walletRegularTransactions map[WalletId][]*RegularTransaction
    walletInfo map[WalletId] walletDetails

    ownerDataMap map[OwnerId]OwnerData

    nextWalletId int
}

func NewRamStorage() Storage {
    storage := &ramStorage {
        walletTransactions: make(map[WalletId][]*ActualTransaction, 0),
        walletRegularTransactions: make(map[WalletId][]*RegularTransaction, 0),
        walletInfo: make(map[WalletId]walletDetails, 0),
        ownerDataMap: make(map[OwnerId]OwnerData, 0),
        nextWalletId: 1}
    return storage
}

func (s *ramStorage) AddActualTransaction(w WalletId, val ActualTransaction) error {
    _, found := s.walletTransactions[w]
    if !found {
        s.walletTransactions[w] = make([]*ActualTransaction, 0, 1)
    }
    s.walletTransactions[w] = append(s.walletTransactions[w], &val)
    return nil
}

func (s *ramStorage) AddRegularTransaction(w WalletId, val RegularTransaction) error {
    _, found := s.walletRegularTransactions[w]
    if !found {
        s.walletRegularTransactions[w] = make([]*RegularTransaction, 0, 1)
    }
    s.walletRegularTransactions[w] = append(s.walletRegularTransactions[w], &val)
    return nil
}

func (s *ramStorage) GetRegularTransactions(w WalletId) ([]*RegularTransaction, error) {
    records := s.walletRegularTransactions[w]
    return records, nil // OK if there are no such transactions
}

func (s *ramStorage) GetActualTransactions(w WalletId, tMin, tMax time.Time) ([]*ActualTransaction, error) {
    allRecords := s.walletTransactions[w]
    if allRecords == nil || len(allRecords) == 0 {
        return nil, nil
    }
    records := make([]*ActualTransaction, 0, len(allRecords))
    for _, r := range allRecords {
        if r.Time.After(tMin) && r.Time.Before(tMax) {
            records = append(records, r)
        }
    }
    return records, nil // OK if no such transactions
}

func (s *ramStorage) GetWalletForOwner(ownerId OwnerId, createIfAbsent bool) (*Wallet, error) {
    return nil, nil
}

func (s *ramStorage) CreateWalletOwner(ownerId OwnerId) (*Wallet, error) {
    _, found := s.ownerDataMap[ownerId]
    if found {
        return nil, errors.New("Owner exists")
    }
    wId := fmt.Sprintf("%d", s.nextWalletId)
    s.nextWalletId++
    ownerData := OwnerData {WalletId: &wId}
    s.ownerDataMap[ownerId] = ownerData
    wallet := NewWalletFromStorage(wId, 1, s)
    return wallet, nil
}

func (s *ramStorage) GetAllOwners() (map[OwnerId]OwnerData, error) {
    // TODO: implement
    return nil, nil
}

func (s *ramStorage) SetWalletInfo(w WalletId, monthStart int) error {
    // TODO: implement
    return nil
}
