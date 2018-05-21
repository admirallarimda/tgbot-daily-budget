package budget

import "log"
import "fmt"
import "strconv"
import "strings"
import "errors"
import "time"
import "math"
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

func NewRedisStorage(server string) Storage {
    s := &RedisStorage{}
    s.client = redis.NewClient(&redis.Options{
        Addr: server})
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

func (s *RedisStorage) AddAmountChange(w Wallet, val AmountChange) error {
    operation := "out"
    if val.Value >= 0 {
        operation = "in"
    }
    key := fmt.Sprintf("wallet:%s:%s:%d", w.ID, operation, val.Time.Unix())
    value := strconv.Itoa(val.Value)

    return s.set(key, value)
}

func (s *RedisStorage) AddRegularChange(w Wallet, change MonthlyChange) error {
    date := change.Date
    if date < 1 || date > 28 {
        return errors.New("Only dates between 1 and 28 are allowed for regular income/outcome setting")
    }

    operation := "out"
    if change.Value >= 0 {
        operation = "in"
    }
    key := fmt.Sprintf("wallet:%s:monthly:%s:%d", w.ID, operation, date)

    log.Printf("Setting regular monthly income/outcome with value %d to key %s", change.Value, key)
    return s.client.LPush(key, change.Value).Err()
}

func (s *RedisStorage) GetMonthlyIncome(w Wallet) (int, error) {
    log.Printf("Getting monthly income")
    income := make(map[string]int, 10)
    scanMatch := fmt.Sprintf("wallet:%s:monthly:*", w.ID)
    var cursor uint64 = 0
    for {
        keys, newcursor, err := s.client.Scan(cursor, scanMatch, 10).Result()
        log.Printf("Monthly income scan by match %s has returned %d keys with cursor %d", scanMatch, len(keys), newcursor)
        cursor = newcursor
        if err != nil {
            log.Printf("Error happened during scanning with match: %s; error: %s", scanMatch, err)
            return 0, err
        }

        for _, k := range keys {
            _, found := income[k]
            if found {
                log.Printf("Key %s has already been used for monthly income calclation, skipping it", k)
                continue
            }

            log.Printf("Getting income values for key %s", k)
            values, err := s.client.LRange(k, math.MinInt64, math.MaxInt64).Result()
            if err != nil {
                log.Printf("Cannot get list for key %s; error: %s", k, err)
                return 0, err
            }

            for _, v := range values {
                val, err := strconv.Atoi(v)
                if err != nil {
                    log.Printf("Could not convert value %s to integer due to error: %s", v, err)
                    return 0, err
                }
                income[k] += val
            }

            log.Printf("Total income for key %s is %d", k, income[k])
        }

        if cursor == 0 {
            log.Printf("Scanning finished")
            break
        }
    }

    totalIncome := 0
    for _, v := range income {
        totalIncome += v
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

    monthly, err := s.GetMonthlyIncome(w)
    if err != nil {
        log.Printf("Could not calculate monthly income for wallet %s with error: %s", w.ID, err)
        return 0, err
    }

    if monthly <= 0 {
        log.Printf("Monthly income is %d (not positive), bot cannot work with such values", monthly)
        return 0, errors.New("Negative monthly income")
    }

    log.Printf("Got monthly income for wallet %s equal to %d", w.ID, monthly)

    monthStart, err := s.getMonthStart(w)
    if err != nil {
        log.Printf("Month start for wallet %s could not be retrieved due to error: %s", w.ID, err)
        return 0, err
    }

    log.Printf("Month start for wallet %s is %d", w.ID, monthStart)

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

            expenseResult := s.client.Get(k)
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
    return int(totalExpense), nil

}

func (s *RedisStorage) GetWalletForOwner(ownerId int64) (*Wallet, error) {
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

func (s *RedisStorage) CreateWalletOwner(ownerId int64) error {
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
