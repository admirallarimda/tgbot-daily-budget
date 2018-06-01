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

func calcCurMonthBorders(walletMonthStartDay int, now time.Time) (time.Time, time.Time) {
    if walletMonthStartDay < 1 || walletMonthStartDay > 28 {
        panic("Date must be between 1 and 28")
    }

    monthStart := time.Date(now.Year(), now.Month(), walletMonthStartDay, 0, 0, 0, 0, time.Local) // TODO: check whether UTC or Local is needed
    if now.Day() < walletMonthStartDay {
        // we've switched the month already, monthStart should be at the previous month
        monthStart = monthStart.AddDate(0, -1, 0)
    }
    monthEnd := monthStart.AddDate(0, 1, 0)
    log.Printf("Month borders are from %s to %s", monthStart, monthEnd)
    return monthStart, monthEnd
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

func (s *RedisStorage) AddActualTransaction(w Wallet, val ActualTransaction) error {
    operation := "out"
    if val.Value >= 0 {
        operation = "in"
    }
    key := fmt.Sprintf("wallet:%s:%s:%d", w.ID, operation, val.Time.Unix())
    fields := make(map[string]interface{}, 3)
    fields["value"] = val.Value
    fields["label"] = val.Label
    fields["raw"] = val.RawText

    return s.setHash(key, fields)
}

func (s *RedisStorage) checkRegularChangeLabelExist(w Wallet, label string) (bool, error) {
    transactions, err := s.getRegularTransactions(w)
    if err != nil {
        log.Printf("Cannot check if label '%s' exists for wallet '%s' regular transactions, error: %s", label, w.ID, err)
        return true, err
    }

    for _, change := range transactions {
        if change.Label == label {
            return true, nil
        }
    }
    return false, nil
}

func (s *RedisStorage) AddRegularChange(w Wallet, change RegularTransaction) error {
    date := change.Date
    if date < 1 || date > 28 {
        return errors.New("Only dates between 1 and 28 are allowed for regular income/outcome setting")
    }

    exists, err := s.checkRegularChangeLabelExist(w, change.Label)
    if err != nil {
        return err
    }
    if exists {
        log.Printf("Label '%s' already exists for wallet '%s', cannot add regular change", change.Label, w.ID)
        return errors.New(fmt.Sprintf("Label '%s' already exists", change.Label))
    }

    operation := "in"
    if change.Value < 0 {
        operation = "out"
    }
    now := time.Now().Unix() // necessary for distinguishing 2 records
    key := fmt.Sprintf("wallet:%s:monthly:%s:%d:%d", w.ID, operation, date, now)

    log.Printf("Setting regular monthly income/outcome with value %d to key %s", change.Value, key)

    fields := make(map[string]interface{}, 3)
    fields["value"] = change.Value
    fields["label"] = change.Label
    return s.setHash(key, fields)
}

func (s *RedisStorage) getRegularTransactions(w Wallet) ([]*RegularTransaction, error) {
    log.Printf("Getting regular wallet transactions for wallet '%s'", w.ID)

    result := make([]*RegularTransaction, 0, 10)

    repeatedKeysGuard := make(map[string]bool, 0)
    scanMatch := fmt.Sprintf("wallet:%s:monthly:*", w.ID)
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

            log.Printf("Getting income values for key %s", k)
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

    log.Printf("Wallet '%s' has %d regular transactions", w.ID, len(result))

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
    return result, nil
}

func (s *RedisStorage) getTransactionsForTimeWindow(w Wallet, t1, t2 time.Time) ([]ActualTransaction, error) {
    if t2.Before(t1) {
        panic("Time borders misaligned")
    }

    matchIn := fmt.Sprintf("wallet:%s:in:*", w.ID)
    matchOut := fmt.Sprintf("wallet:%s:out:*", w.ID)

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

    result := make([]ActualTransaction, 0, len(allKeys))
    for _, k := range allKeys {
        tUnix, err := strconv.ParseInt(strings.Split(k, ":")[3], 10, 64)
        if err != nil {
            log.Printf("Cannot convert time from key '%s' to integer, error: %s", k, err)
        }
        t := time.Unix(tUnix, 0)
        if t.After(t1) && t.Before(t2) {
            log.Printf("Key '%s' is in our time window, getting data from it", t)
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
            result = append(result, *NewActualTransaction(value, t, fields["label"], ""))
        }
    }
    return result, nil
}

func (s *RedisStorage) GetMonthlyIncome(w Wallet) (int, error) {
    log.Printf("Getting monthly income")
    transactions, err := s.getRegularTransactions(w)
    if err != nil {
        log.Print("Could not get monthly transactions for wallet '%s', error: %s", w.ID, err)
        return 0, err
    }

    totalIncome := 0
    for _, change := range transactions {
        totalIncome += change.Value
    }

    log.Printf("Total income for wallet %s is %d", w.ID, totalIncome)

    return totalIncome, nil
}

func (s *RedisStorage) getMonthStart(w Wallet) (int, error) {
    log.Printf("Looking for month start for wallet %s", w.ID)
    key := fmt.Sprintf("wallet:%s", w.ID)

    monthStartField := "monthStart"
    exists := s.client.HExists(key, monthStartField)
    if exists.Err() != nil {
        log.Printf("Could not check existance of month start field '%s' for wallet key %s", monthStartField, key)
        return 0, nil
    }
    if exists.Val() == false {
        log.Printf("Month start value for wallet key %s is not set, using default", key)
        return 1, nil
    }

    log.Printf("Month start field '%s' exists for wallet key %s, retrieving it", monthStartField, key)
    res := s.client.HGet(key, monthStartField)

    if res.Err() != nil {
        log.Printf("Could not get month start for wallet with key %s, error: %s", key, res.Err())
        return 0, res.Err()
    }

    val, err := res.Int64()
    if err != nil {
        log.Printf("Could not convert value %s of month start for wallet with key %s, error: %s", res.Val(), key, err)
        return 0, err
    }

    if val < 1 || val > 28 {
        log.Printf("Month start for wallet key %s is out of expected ranges", key)
        return 0, errors.New("Month start value out of range")
    }

    return int(val), nil
}

func (s *RedisStorage) GetMonthlyIncomeTillDate(w Wallet, t time.Time) (int, error) {
    log.Printf("Calculating monthly income for wallet %s till time %s", w.ID, t)

    regularTransactions, err := s.getRegularTransactions(w)
    if err != nil {
        log.Print("Could not get monthly transactions for wallet '%s', error: %s", w.ID, err)
        return 0, err
    }
    regularTransactionsLabeled := make(map[string]int, len(regularTransactions))
    for _, regularElem := range regularTransactions {
        if regularElem.Label == "" {
            panic("Label for regular transaction is empty")
        }
        regularTransactionsLabeled[regularElem.Label] = regularElem.Value
    }

    monthStart, err := s.getMonthStart(w)
    if err != nil {
        log.Printf("Month start for wallet %s could not be retrieved due to error: %s", w.ID, err)
        return 0, err
    }
    log.Printf("Month start for wallet %s is %d", w.ID, monthStart)

    t1, t2 := calcCurMonthBorders(monthStart, time.Now())
    transactions, err := s.getTransactionsForTimeWindow(w, t1, t2)
    if err != nil {
        log.Printf("Could not receive transactions due to error: %s", err)
        return 0, err
    }

    monthly := 0
    regularReplacedWithActual := make(map[string]bool, 0)
    for _, transaction := range transactions {
        label := transaction.Label
        value := transaction.Value
        if regularValue, regularFound := regularTransactionsLabeled[label]; regularFound {
            log.Printf("Found a transaction labeled '%s' which replaces value %d -> %d", label, regularValue, value)
            monthly += value
            regularReplacedWithActual[label] = true
        } else if value > 0 {
            log.Printf("Found a transaction labeled '%s' with positive income %d, using it for monthly income calculation", label, value)
            monthly += value
        } // else value <= 0 { log.Printf("Skipping transaction labeled '%s' with negative value %d", label, value) }
    }

    // adding those regular transactions which were not yet matches with an actial one
    for label, value := range regularTransactionsLabeled {
        if _, found := regularReplacedWithActual[label]; found {
            log.Printf("Regular transaction labeled '%s' with value %d has already been matched by an actual transaction, skipping", label, value)
            continue
        }
        log.Printf("Regular transaction labeled '%s' with value %d will be used for monthly income calculation", label, value)
        monthly += value
    }
    log.Printf("Final montly income is: %d", monthly)

    var result float32 = 0
    // calculating result based on hoe many days have passed considering whether we've reached the end of prev month
    curDay := t.Day()
    if curDay >= monthStart {
        daysInCurMonth := daysInMonth[t.Month()]
        result = float32(monthly) / float32(daysInCurMonth) * float32((curDay - monthStart + 1)) // +1 as we assume that daily portion is granted at the beginning of the day
    } else {
        // tricky code to calc how many days have passed if we've reached the end of the previous month
        prevMonth := time.December
        if t.Month() != time.January {
            prevMonth = t.Month() - 1
        }
        result = float32(monthly) / float32(31 - (monthStart - curDay) - (31 - daysInMonth[prevMonth]))
    }

    log.Printf("Calculated montly income till date %s: it equals to %f", t, result)
    return int(result), nil
}

func (s *RedisStorage) GetMonthlyExpenseTillDate(w Wallet, t time.Time) (int, error) {
    log.Printf("Getting monthly expenses for wallet %s till date %s", w.ID, t)

    monthStartDay, err := s.getMonthStart(w)
    if err != nil {
        log.Printf("Month start for wallet %s could not be retrieved due to error: %s", w.ID, err)
        return 0, err
    }
    monthStart := time.Date(t.Year(), t.Month(), monthStartDay, 0, 0, 0, 0, time.Local) // TODO: check whether UTC or Local is needed
    if t.Day() < monthStartDay {
        // we've switched the month already, monthStart should be at the previous month
        monthStart = monthStart.AddDate(0, -1, 0)
    }
    log.Printf("Month start is calculated to be %s", monthStart)

    var totalExpense int64 = 0
    scanMatch := fmt.Sprintf("wallet:%s:out:*", w.ID)
    var cursor uint64 = 0
    for {
        keys, newcursor, err := s.client.Scan(cursor, scanMatch, 10).Result()
        log.Printf("Monthly expense scan by match %s has returned %d keys with cursor %d", scanMatch, len(keys), newcursor)
        cursor = newcursor
        if err != nil {
            log.Printf("Error happened during scanning with match: %s; error: %s", scanMatch, err)
            return 0, err
        }

        for _, k := range keys {
            keyParts := strings.Split(k, ":")
            timeStr := keyParts[3]
            timeUnix, err := strconv.ParseInt(timeStr, 10, 64)
            if err != nil {
                log.Printf("Could not convert %s to int due to error: %s", timeStr, err)
                continue
            }
            expenseTime := time.Unix(timeUnix, 0)
            if expenseTime.Before(monthStart) || expenseTime.After(t) {
                log.Printf("Expense time %s is either before month start %s or right border %s", expenseTime, monthStart, t)
                continue
            }

            expenseResult := s.client.HGet(k, "value")
            if expenseResult.Err() != nil {
                log.Printf("Could not get value for key %s due to error: %s", k, expenseResult.Err())
                continue
            }

            expenseAmount, err := expenseResult.Int64()
            if err != nil {
                log.Printf("Could not convert expense amount %s to int, error: %s", expenseResult.Val(), err)
                continue
            }
            totalExpense += expenseAmount
        }

        if cursor == 0 {
            log.Printf("Scanning finished")
            break
        }
    }

    log.Printf("Calculated total expense %d for wallet %s from %s till %s", totalExpense, w.ID, monthStart, t)
    return -int(totalExpense), nil // returning positive value

}

func (s *RedisStorage) GetWalletForOwner(ownerId OwnerId) (*Wallet, error) {
    key := fmt.Sprintf("owner:%d", ownerId)
    log.Printf("Getting wallet for owner via key %s", key)
    result := s.client.HGetAll(key)
    if result == nil {
        log.Printf("Could not get user info for owner with key %s", key)
        return nil, errors.New("No owner info")
    }

    log.Printf("Got info about owner key %s. Info: %v", key, result.Val())
    // TODO: add teams
    walletIdStr, found := result.Val()["wallet"]
    if !found {
        log.Printf("No wallet found for owner key %s", key)
        return nil, errors.New("No wallet for owner")
        //TODO: request new wallet?
    }

    walletId, err := uuid.FromString(walletIdStr)
    if err != nil {
        log.Printf("Could not convert wallet ID %s to uuid, error: %s", walletIdStr, err)
        return nil, err
    }
    return &Wallet{ID: walletId}, nil
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

func (s *RedisStorage) CreateWalletOwner(ownerId OwnerId) error {
    log.Printf("Starting creation of owner %d", ownerId)

    key := fmt.Sprintf("owner:%d", ownerId)
    owner := s.client.HGetAll(key)
    if owner != nil && len(owner.Val()) > 0 {
        log.Printf("Owner %d has been already created", ownerId)
        return errors.New("Owner exists")
    }

    walletId, err := s.createWallet()
    if err != nil {
        log.Printf("Could not create wallet for owner %d with error: %s", ownerId, err)
        return err
    }
    log.Printf("Wallet %s has been created for owner %d", walletId, ownerId)

    s.attachWalletToUser(key, walletId)

    return nil
}

func (s *RedisStorage) createWallet() (string, error) {
    final_id := ""
    for final_id == "" {
        id, err := uuid.NewV4()
        if err != nil {
            log.Printf("Could get new wallet UUID due to error: %s", err)
            return "", err
        }

        key := fmt.Sprintf("wallet:%s", id.String())
        log.Printf("Checking if wallet with key %s exists", key)
        result := s.client.HGetAll(key)
        if result != nil && len(result.Val()) > 0 {
            log.Printf("Wallet with key %s exists, trying another one", key)
            continue
        }

        log.Printf("Wallet with key %s doesn't exist, using it", key)
        s.client.HSet(key, "created", time.Now().Unix())
        final_id = id.String()
    }

    return final_id, nil
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
