package budget

import "fmt"

func keyOwner(owner OwnerId) string {
	return fmt.Sprintf("owner:%d", owner)
}

func keyActualTransaction(wId WalletId, operation string, tUnix int64) string {
	return fmt.Sprintf("wallet:%s:%s:%d", wId, operation, tUnix)
}

func keyRegularTransaction(wId WalletId, operation string, regularDate int, addDateUnix int64) string {
	return fmt.Sprintf("wallet:%s:monthly:%s:%d:%d", wId, operation, regularDate, addDateUnix)
}

func keyWallet(wId WalletId) string {
	return fmt.Sprintf("wallet:%s", wId)
}

func scannerRegularTransactions(wId WalletId) string {
	return fmt.Sprintf("wallet:%s:monthly:*", wId)
}
