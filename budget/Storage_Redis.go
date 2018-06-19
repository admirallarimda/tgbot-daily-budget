package budget

import "log"
import "fmt"
import "strconv"
import "strings"
import "errors"
import "time"
import "github.com/go-redis/redis"
import "github.com/satori/go.uuid"

var daysInMonth = map[time.Month]int {  time.January: 31,
                                        time.February: 28, // TODO: handle leap year
                                        time.March: 31,
                                        time.April: 30,
                                        time.May: 31,
                                        time.June: 30,
                                        time.July: 31,
                                        time.August: 31,
                                        time.September: 30,
                                        time.October: 31,
                                        time.November: 30,
                                        time.December: 31 }

const defaultMonthStart = 1

type RedisStorage struct {
    client *redis.Client
}

func uniqueStringSlice(s []string) []string {
    result := make([]string, 0, len(s))
    seen := make(map[string]bool, len(s))
    for _, elem := range s {
        if _, found := seen[elem]; found {
            continue
        }
        result = append(result, elem)
        seen[elem] = true
    }
    return result
}

func NewRedisStorage(server string, db int) Storage {
    s := &RedisStorage{}
    s.client = redis.NewClient(&redis.Options{
        Addr: server,
        DB: db})
    return s
}

func (s *RedisStorage) set(key, value string) error {
    log.Printf("Setting key: %s value: %s", key, value)

    status := s.client.Set(key, value, 0)
    err := status.Err()
    if err != nil {
        log.Printf("Unable to set value %s to key %s; error: %s", value, key, err)
        return err
    }

    return nil
}

func (s *RedisStorage) setHash(key string, fields map[string]interface{}) error {
    log.Printf("Setting hash at key %s for %d hash keys", key, len(fields))

    status := s.client.HMSet(key, fields)
    err := status.Err()
    if err != nil {
        log.Printf("Unable to set hash with key %s; error: %s", key, err)
        return err
    }

    return nil
}

func (s *RedisStorage) AddActualTransaction(w WalletId, val ActualTransaction) error {
    operation := "out"
    if val.Value >= 0 {
        operation = "in"
    }
    key := keyActualTransaction(w, operation, val.Time.Unix())
    fields := make(map[string]interface{}, 3)
    fields["value"] = val.Value
    fields["label"] = val.Label
    fields["raw"] = val.RawText

    return s.setHash(key, fields)
}

func (s *RedisStorage) AddRegularTransaction(w WalletId, t RegularTransaction) error {
    operation := "in"
    if t.Value < 0 {
        operation = "out"
    }
    now := time.Now().Unix() // necessary for distinguishing 2 records
    key := keyRegularTransaction(w, operation, t.Date, now)

    log.Printf("Setting regular monthly income/outcome with value '%d' to key '%s'", t.Value, key)

    fields := make(map[string]interface{}, 3)
    fields["value"] = t.Value
    fields["label"] = t.Label
    return s.setHash(key, fields)
}

func (s *RedisStorage) RemoveRegularTransaction(w WalletId, t RegularTransaction) error {
    scanner := scannerRegularTransactions(w)
    keys, err := s.getAllKeys(scanner)
    if err != nil {
        log.Printf("Could not scan for scanner '%s' due to error: %s", scanner, err)
    }
    targetKey := ""
    for _, key := range keys {
        keyParts := strings.Split(key, ":")
        dateStr := keyParts[4]
        if dateStr != strconv.Itoa(t.Date) {
            // it's not the key we're looking for
            continue
        }
        fields, err := s.client.HGetAll(key).Result()
        if err != nil {
            log.Printf("Could not get fields for key '%s' during regular transaction removal due to error: %s", w, err)
            return err
        }
        if t.Label == fields["label"] && strconv.Itoa(t.Value) == fields["value"] {
            targetKey = key
            break
        }
    }

    if targetKey == "" {
        log.Printf("No transaction in Redis found for wallet '%s' for transaction removal", w)
        return errors.New("Specified transaction has not been found in DB")
    }

    log.Printf("Removing transaction with key '%s'", targetKey)
    count, err := s.client.Del(targetKey).Result()
    if err != nil {
        log.Printf("Could not remove transaction with key '%s' due to error: %s", targetKey, err)
        return err
    }
    if count != 1 {
        log.Printf("Incorrectly removed transaction with key '%s' - removed %d records instead of 1", targetKey, count)
        return errors.New("Incorrect number of transactions have been removed")
    }
    return nil
}

func (s *RedisStorage) GetRegularTransactions(w WalletId) ([]*RegularTransaction, error) {
    log.Printf("Getting regular wallet transactions for wallet '%s'", w)

    result := make([]*RegularTransaction, 0, 10)

    repeatedKeysGuard := make(map[string]bool, 0)
    scanMatch := scannerRegularTransactions(w)
    var cursor uint64 = 0
    for {
        keys, newcursor, err := s.client.Scan(cursor, scanMatch, 10).Result()
        log.Printf("Monthly income scan by match %s has returned %d keys with cursor %d", scanMatch, len(keys), newcursor)
        cursor = newcursor
        if err != nil {
            log.Printf("Error happened during scanning with match: %s; error: %s", scanMatch, err)
            return nil, err
        }

        for _, k := range keys {
            _, found := repeatedKeysGuard[k]
            if found {
                log.Printf("Key %s has already been already added, skipping it", k)
                continue
            }

            log.Printf("Getting income values for key '%s'", k)
            fields, err := s.client.HGetAll(k).Result()
            if err != nil {
                log.Printf("Cannot get wallet info for key '%s', error: %s", k, err)
                continue // let's just skip it
            }

            valueStr := fields["value"]
            value, err := strconv.Atoi(valueStr)
            if err != nil {
                log.Printf("Could not convert value %s to integer, error: %s", valueStr, err)
                return nil, err
            }

            keyParts := strings.Split(k, ":")
            dateStr := keyParts[4]
            date, err := strconv.Atoi(dateStr)
            if err != nil {
                log.Printf("Could not convert time %s to integer, error: %s", dateStr, err)
                return nil, err
            }

            result = append(result, NewRegularTransaction(value, date, fields["label"]))

            repeatedKeysGuard[k] = true
        }

        if cursor == 0 {
            log.Printf("Scanning finished")
            break
        }
    }

    log.Printf("Wallet '%s' has %d regular transactions", w, len(result))

    return result, nil
}

func (s *RedisStorage) getAllKeys(matchPattern string) ([]string, error) {
    log.Printf("Starting scanning for match '%s'", matchPattern)
    result := make([]string, 0, 10)
    var cursor uint64 = 0
    for {
        keys, newcursor, err := s.client.Scan(cursor, matchPattern, 10).Result()
        if err != nil {
            log.Printf("Error happened while scanning with match pattern '%s', error: %s", matchPattern, err)
            return nil, err
        }
        cursor = newcursor
        result = append(result, keys...)
        if cursor == 0 {
            log.Printf("Scanning for '%s' has finished, result contains %d elements", matchPattern, len(result))
            break
        }
    }
    log.Printf("Scanner '%s' returned %d keys", matchPattern, len(result))
    return result, nil
}

func (s *RedisStorage) GetActualTransactions(w WalletId, t1, t2 time.Time) ([]*ActualTransaction, error) {
    if t2.Before(t1) {
        panic("Time borders misaligned")
    }

    matchIn := fmt.Sprintf("wallet:%s:in:*", w)
    matchOut := fmt.Sprintf("wallet:%s:out:*", w)

    keysIn, err := s.getAllKeys(matchIn)
    if err != nil {
        log.Printf("Error happened during getting all IN transactions via match '%s', error: %s", matchIn, err)
        return nil, err
    }
    log.Printf("Found %d keys matching scanner '%s'", len(keysIn), matchIn)
    keysOut, err := s.getAllKeys(matchOut)
    if err != nil {
        log.Printf("Error happened during getting all OUT transactions via match '%s', error: %s", matchOut, err)
        return nil, err
    }
    log.Printf("Found %d keys matching scanner '%s'", len(keysOut), matchOut)
    allKeys := append(keysIn, keysOut...)
    allKeys = uniqueStringSlice(allKeys)

    result := make([]*ActualTransaction, 0, len(allKeys))
    for _, k := range allKeys {
        tUnix, err := strconv.ParseInt(strings.Split(k, ":")[3], 10, 64)
        if err != nil {
            log.Printf("Cannot convert time from key '%s' to integer, error: %s", k, err)
        }
        t := time.Unix(tUnix, 0)
        if t.After(t1) && (t.Equal(t2) || t.Before(t2)) {
            log.Printf("Key '%s' corresponding to date %s is in our time window, getting data from it", k, t)
            fields, err := s.client.HGetAll(k).Result()
            if err != nil {
                log.Printf("Cannot get transaction for key '%s', error: %s", k, err)
                continue // let's just skip it
            }

            valueStr := fields["value"]
            value, err := strconv.Atoi(valueStr)
            if err != nil {
                log.Printf("Could not convert value %s to integer, error: %s", valueStr, err)
                return nil, err
            }
            result = append(result, NewActualTransaction(value, t, fields["label"], ""))
        }
    }
    return result, nil
}


func (s *RedisStorage) GetWalletForOwner(ownerId OwnerId, createIfAbsent bool) (*Wallet, error) {
    key := keyOwner(ownerId)
    log.Printf("Getting wallet for owner via key '%s'", key)
    result := s.client.HGetAll(key)
    if result == nil {
        if !createIfAbsent {
            log.Printf("Could not get user info for owner with key %s", key)
            return nil, errors.New("No owner info")
        }
        return s.CreateWalletOwner(ownerId)
    }

    log.Printf("Got info about owner key '%s'. Info: %v", key, result.Val())
    walletId, found := result.Val()["wallet"]
    if !found {
        log.Printf("No wallet found for owner key '%s'", key)
        if !createIfAbsent {
            return nil, errors.New("No wallet for owner")
        }
        return s.CreateWalletOwner(ownerId)
    }

    walletKey := keyWallet(WalletId(walletId))
    fields, err := s.client.HGetAll(walletKey).Result()
    if err != nil {
        log.Printf("Could not get wallet fields via key '%s'", walletKey)
        return nil, err
    }
    monthStart := defaultMonthStart
    monthStartStr, found := fields["monthStart"]
    if found {
        monthStart, err = strconv.Atoi(monthStartStr)
        if err != nil {
            log.Printf("Could not convert month start %s for wallet '%s' due to error: %s", monthStartStr, walletKey, err)
            return nil, err
        }
    }

    return NewWalletFromStorage(walletId, monthStart, s), nil
}

func (s *RedisStorage) attachWalletToUser(ownerKey string, walletId string) error {
    res := s.client.HSet(ownerKey, "wallet", walletId)

    if res != nil && res.Val() == false {
        log.Printf("Could not attach owner '%s' and wallet '%s'", ownerKey, walletId)
        return errors.New("Could not attach wallet to owner")
    }

    log.Printf("Attached owner with key '%s' and wallet '%s'", ownerKey, walletId)
    return nil
}

func (s *RedisStorage) CreateWalletOwner(ownerId OwnerId) (*Wallet, error) {
    log.Printf("Starting creation of owner %d", ownerId)

    key := keyOwner(ownerId)
    owner := s.client.HGetAll(key)
    if owner != nil && len(owner.Val()) > 0 {
        log.Printf("Owner %d has been already created", ownerId)
        return nil, errors.New("Owner exists")
    }

    wallet, err := s.createWallet()
    if err != nil {
        log.Printf("Could not create wallet for owner %d with error: %s", ownerId, err)
        return nil, err
    }
    log.Printf("Wallet %s has been created for owner %d", wallet.ID, ownerId)

    s.attachWalletToUser(key, string(wallet.ID))

    return wallet, nil
}

func (s *RedisStorage) createWallet() (*Wallet, error) {
    var wallet *Wallet = nil
    for wallet == nil {
        id, err := uuid.NewV4()
        if err != nil {
            log.Printf("Could get new wallet UUID due to error: %s", err)
            return nil, err
        }

        key := keyWallet(WalletId(id.String()))
        log.Printf("Checking if wallet with key %s exists", key)
        result := s.client.HGetAll(key)
        if result != nil && len(result.Val()) > 0 {
            log.Printf("Wallet with key %s exists, trying another one", key)
            continue
        }

        log.Printf("Wallet with key %s doesn't exist, using it", key)
        s.client.HSet(key, "created", time.Now().Unix())
        s.client.HSet(key, "monthStart", defaultMonthStart)
        wallet = NewWalletFromStorage(id.String(), defaultMonthStart, s)
    }

    return wallet, nil
}

func parseOwnerData(data map[string]string) OwnerData {
    ownerData := OwnerData{}

    if walletId, found := data["wallet"]; found {
        ownerData.WalletId = &walletId
    }
    if reminderTime, found := data["dailyNotifTime"]; found {
        dur, err := time.ParseDuration(reminderTime)
        if err == nil {
            ownerData.DailyReminderTime = &dur
        }
    } else {
        dur := time.Duration(9) * time.Hour
        ownerData.DailyReminderTime = &dur
    }

    return ownerData
}

func (s *RedisStorage) GetAllOwners() (map[OwnerId]OwnerData, error) {
    matcher := "owner:*"
    var cursor uint64 = 0
    resultMap := make(map[OwnerId]OwnerData, 0)
    for {
        keys, newcursor, err := s.client.Scan(cursor, matcher, 10).Result()
        if err != nil {
            log.Printf("Could not get owners via match %s due to error: %s", matcher, err)
            return nil, err
        }
        cursor = newcursor
        log.Printf("Received new batch of %d keys (cursor: %d)", len(keys), cursor)
        for _, k := range keys {
            log.Printf("Getting data for key %s", k)
            rawData, err := s.client.HGetAll(k).Result()
            if err != nil {
                log.Printf("Owner data for key %s cannot be retrieved due to error: %s", k, err)
                continue
            }
            ownerData := parseOwnerData(rawData)
            log.Printf("Owner data has been parsed into: %+v", ownerData)
            keyParts := strings.Split(k, ":")
            ownerId, err := strconv.ParseInt(keyParts[1], 10, 64)
            if err != nil {
                log.Printf("Could not get owner ID from key %s; error: %s", k, err)
                continue
            }
            resultMap[OwnerId(ownerId)] = ownerData
        }

        if cursor == 0 {
            log.Printf("Scanning for all owners finished")
            break
        }
    }
    return resultMap, nil
}

func (s *RedisStorage) SetWalletInfo(w WalletId, monthStart int) error {
    key := keyWallet(w)
    fields := make(map[string]interface{}, 3)
    fields["monthStart"] = monthStart
    return s.setHash(key, fields)
}
