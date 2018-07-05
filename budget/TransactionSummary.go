package budget

import "time"

type TransactionSummary struct {
    TimeStart, TimeEnd time.Time

    ExpenseSummary map[string]int
}

func NewTransactionSummary(start, end time.Time) *TransactionSummary {
    result := &TransactionSummary{TimeStart: start,
                                  TimeEnd: end}
    result.ExpenseSummary = make(map[string]int, 0)

    return result
}
