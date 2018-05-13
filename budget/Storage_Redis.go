package budget

import "log"
import "fmt"
import "strconv"
import "errors"
import "time"
import "github.com/go-redis/redis"
import "github.com/satori/go.uuid"

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

func (s *RedisStorage) attachWalletToUser(userKey string, walletId string) error {
    res := s.client.HSet(userKey, "wallet", walletId)

    if res != nil && res.Val() == false {
        log.Printf("Could not attach user '%s' and wallet '%s'", userKey, walletId)
        return errors.New("Could not attach wallet to user")
    }

    log.Printf("Attached user with key '%s' and wallet '%s'", userKey, walletId)
    return nil
}

func (s *RedisStorage) CreateUser(userId int) error {
    log.Printf("Starting creation of user %d", userId)

    key := fmt.Sprintf("user:%d", userId)
    user := s.client.HGetAll(key)
    if user != nil && len(user.Val()) > 0 {
        log.Printf("User %d has been already created", userId)
        return errors.New("User exists")
    }

    walletId, err := s.createWallet()
    if err != nil {
        log.Printf("Could not create wallet for user %d with error: %s", userId, err)
        return err
    }
    log.Printf("Wallet %s has been created for user %d", walletId, userId)

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
