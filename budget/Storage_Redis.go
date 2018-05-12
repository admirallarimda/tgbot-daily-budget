package budget

import "log"
import "fmt"
import "strconv"
import "errors"
import "github.com/go-redis/redis"
import "github.com/satori/go.uuid"

type RedisStorage struct {
    client redis.Client
}

func NewRedisStorage() Storage {
    s := &RedisStorage{}
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

func (s *RedisStorage) AddIncome(w Wallet, val AmountChange) error {
    key := fmt.Sprintf("wallet:%s:in:%d", w.ID, val.Time.Unix())
    value := strconv.Itoa(val.Value)

    return s.set(key, value)
}

func (s *RedisStorage) AddExpense(w Wallet, val AmountChange) error {
    key := fmt.Sprintf("wallet:%s:out:%d", w.ID, val.Time.Unix())
    value := strconv.Itoa(val.Value)

    return s.set(key, value)
}

func (s *RedisStorage) GetWalletForUser(userId int) (*Wallet, error) {
    key := fmt.Sprintf("user:%d", userId)
    log.Printf("Getting wallet for user via key %s", key)
    result := s.client.HGetAll(key)
    if result == nil {
        log.Printf("Could not get user info for user with key %s", key)
        // TODO: new user info?
        return nil, errors.New("No user info")
    }

    log.Printf("Got info about user key %s. Info: %v", key, result.Val())
    // TODO: add teams
    walletIdStr, found := result.Val()["wallet"]
    if !found {
        log.Printf("No wallet found for user key %s", key)
        return nil, errors.New("No wallet for user")
        //TODO: request new wallet?
    }

    walletId, err := uuid.FromString(walletIdStr)
    if err != nil {
        log.Printf("Could not convert wallet ID %s to uuid, error: %s", walletIdStr, err)
        return nil, err
    }
    return &Wallet{ID: walletId}, nil
}
