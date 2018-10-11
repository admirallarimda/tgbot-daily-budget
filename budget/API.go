package budget

import "log"
import "github.com/admirallarimda/tgbotbase"

func CreateStorageConnection(pool tgbotbase.RedisPool) Storage {
	return NewRedisStorage(pool.GetConnByName("budget"))
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
