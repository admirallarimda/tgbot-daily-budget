package budget

import "testing"
import "time"
import "strconv"

func TestExtraIncome(t *testing.T) {
    w := NewWalletFromStorage("ExtraIncome", 1, nil)
    txs := newTransactionCollection()

    expected := 0
    regular_vals := []int{100, 200, -50}
    for i, v := range regular_vals {
        txs.regular_txs = append(txs.regular_txs, *NewRegularTransaction(v, 1 + i % 28, strconv.Itoa(i)))
        expected += v
    }

    t1 := time.Date(2018, 6, 1, 12, 0, 0, 0, time.UTC)

    extra_income := []int{50, 70}
    for _, v := range extra_income {
        txs.actual_txs = append(txs.actual_txs, *NewActualTransaction(v, t1, "", ""))
        expected += v
    }

    income := w.calcMonthlyIncomeTillDate(*txs, t1)
    expected = expected / 30 // TODO: use time-getting after this test merge
    if income != expected {
        t.Errorf("income: %d; expected: %d", income, expected)
    }
}
