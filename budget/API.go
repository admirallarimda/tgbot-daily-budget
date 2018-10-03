package budget

import "log"
import "github.com/admirallarimda/tgbot-daily-budget/botcfg"

var redisServer string
var redisDB int
var redisPassword string

func Init(cfg botcfg.Config) {
	redisServer = cfg.Redis.Server
	redisDB = cfg.Redis.DB
	redisPassword = cfg.Redis.Pass
}

func CreateStorageConnection() Storage {
	return NewRedisStorage(redisServer, redisDB, redisPassword)
}

func GetWalletForOwner(owner OwnerId, createIfAbsent bool, storageconn Storage) (*Wallet, error) {
	log.Printf("Acquiring wallet for owner %d", owner)
	wallet, err := storageconn.GetWalletForOwner(owner, createIfAbsent)
	if err != nil {
		log.Printf("Could not get wallet for owner %d due to error: %s", owner, err)
		return nil, err
	}

	return wallet, nil
}
