package budget

type transactionCollection struct {
	regular_txs []RegularTransaction
	actual_txs  []ActualTransaction

	// below are cached results of the functions
	matched_actual_txs *map[string][]ActualTransaction
	actual_income_txs  *[]ActualTransaction
	actual_expense_txs *[]ActualTransaction
}

func (txs *transactionCollection) getActualTransactions() []ActualTransaction {
	return txs.actual_txs
}

func (txs *transactionCollection) getActualIncomeTransactions() []ActualTransaction {
	if txs.actual_income_txs == nil {
		result := make([]ActualTransaction, 0, len(txs.actual_txs))
		for _, t := range txs.actual_txs {
			if t.Value > 0 {
				result = append(result, t)
			}
		}
		txs.actual_income_txs = &result
	}
	return *txs.actual_income_txs
}

func (txs *transactionCollection) getActualExpenseTransactions() []ActualTransaction {
	if txs.actual_expense_txs == nil {
		result := make([]ActualTransaction, 0, len(txs.actual_txs))
		for _, t := range txs.actual_txs {
			if t.Value < 0 {
				result = append(result, t)
			}
		}
		txs.actual_expense_txs = &result
	}
	return *txs.actual_expense_txs
}

func (txs *transactionCollection) getRegularTransactions() []RegularTransaction {
	return txs.regular_txs
}

func (txs *transactionCollection) getMatchedActualTransactions() map[string][]ActualTransaction {
	if txs.matched_actual_txs == nil {
		matched := make(map[string][]ActualTransaction, len(txs.regular_txs))
		for _, regular := range txs.regular_txs {
			matched[regular.Label] = make([]ActualTransaction, 0, 0)
		}
		for _, actual := range txs.actual_txs {
			if actual.Label == "" {
				continue
			}
			if _, found := matched[actual.Label]; !found {
				continue
			}
			matched[actual.Label] = append(matched[actual.Label], actual)
		}
		txs.matched_actual_txs = &matched
	}
	return *txs.matched_actual_txs
}

func newTransactionCollection() *transactionCollection {
	txs := &transactionCollection{
		regular_txs: make([]RegularTransaction, 0, 0),
		actual_txs:  make([]ActualTransaction, 0, 0)}
	return txs
}
