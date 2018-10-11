package budget

import "testing"
import "strconv"
import "time"

var walletId int = 1

func testNewWalletId() string {
	res := strconv.Itoa(walletId)
	walletId += 1
	return res
}

func testNewDate(month time.Month) time.Time {
	return time.Date(2018, month, 1, 12, 0, 0, 0, time.UTC)
}

func TestNoTransactions(t *testing.T) {
	w := NewWalletFromStorage(testNewWalletId(), 1, nil)
	txs := newTransactionCollection()
	income := w.calcMonthlyIncomeTillDate(*txs, testNewDate(6))
	if income != 0 {
		t.Errorf("income: %d", income)
	}
}

func TestRegularIncomeOnly(t *testing.T) {
	w := NewWalletFromStorage(testNewWalletId(), 1, nil)
	txs := newTransactionCollection()

	val1 := 100
	val2 := 200
	transactions := make([]RegularTransaction, 0, 2)
	transactions = append(transactions, *NewRegularTransaction(val1, 1, "pos1"), *NewRegularTransaction(val2, 5, "pos2"))
	txs.regular_txs = transactions

	income := w.calcMonthlyIncomeTillDate(*txs, testNewDate(6))
	expected := (val1 + val2) / 30 // 30 days in month 6 - June
	if income != expected {
		t.Errorf("income: %d; expected: %d", income, expected)
	}
}

func TestRegularExpenseOnly(t *testing.T) {
	w := NewWalletFromStorage(testNewWalletId(), 1, nil)
	txs := newTransactionCollection()

	val1 := -100
	val2 := -200
	transactions := make([]RegularTransaction, 0, 2)
	transactions = append(transactions, *NewRegularTransaction(val1, 1, "pos1"), *NewRegularTransaction(val2, 5, "pos2"))
	txs.regular_txs = transactions

	income := w.calcMonthlyIncomeTillDate(*txs, testNewDate(6))
	// might seem strange that expected is < 0, but it is not this function's responsibility
	expected := (val1 + val2) / 30 // 30 days in month 6 - June
	if income != expected {
		t.Errorf("income: %d; expected: %d", income, expected)
	}
}

func TestRegularMixed_IncomeGreater(t *testing.T) {
	w := NewWalletFromStorage(testNewWalletId(), 1, nil)
	txs := newTransactionCollection()

	valPos1 := 1200
	valPos2 := 200
	valNeg1 := -100
	valNeg2 := -50
	valNeg3 := -200
	transactions := make([]RegularTransaction, 0, 5)
	transactions = append(transactions, *NewRegularTransaction(valPos1, 1, "1"),
		*NewRegularTransaction(valPos2, 5, "2"),
		*NewRegularTransaction(valNeg1, 1, "3"),
		*NewRegularTransaction(valNeg2, 19, "4"),
		*NewRegularTransaction(valNeg3, 4, "5"))
	txs.regular_txs = transactions

	income := w.calcMonthlyIncomeTillDate(*txs, testNewDate(6))
	expected := (valPos1 + valPos2 + valNeg1 + valNeg2 + valNeg3) / 30 // 30 days in month 6 - June
	if income != expected {
		t.Errorf("income: %d; expected: %d", income, expected)
	}
}

func TestRegularMixed_ExpenseGreater(t *testing.T) {
	w := NewWalletFromStorage(testNewWalletId(), 1, nil)
	txs := newTransactionCollection()

	values := []int{1000, -500, -200, 100, -700}
	transactions := make([]RegularTransaction, 0, len(values))
	for i, v := range values {
		transactions = append(transactions, *NewRegularTransaction(v, 1+i%30, strconv.Itoa(i)))
	}
	txs.regular_txs = transactions

	income := w.calcMonthlyIncomeTillDate(*txs, testNewDate(6))
	// might seem strange that expected is < 0, but it is not this function's responsibility
	valSum := 0
	for _, v := range values {
		valSum += v
	}
	expected := valSum / 30 // 30 days in month 6 - June
	if income != expected {
		t.Errorf("income: %d; expected: %d", income, expected)
	}
}
