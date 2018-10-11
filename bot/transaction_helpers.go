package bot

import "time"
import "fmt"
import "github.com/admirallarimda/tgbot-daily-budget/budget"

func constructIncomeMessage(w *budget.Wallet) string {
	plannedIncomeMsg := ""
	if plannedIncome, err := w.GetPlannedMonthlyIncome(); err == nil {
		if correctedMonthlyIncome, correctedDailyIncome, err := w.GetCorrectedMonthlyIncome(time.Now()); err == nil {
			plannedIncomeMsg = fmt.Sprintf("Planned monthly income: %d", plannedIncome)
			if plannedIncome != correctedMonthlyIncome {
				plannedIncomeMsg = fmt.Sprintf("%s (with corrections for current month: monthly: %d; daily: %d)", plannedIncomeMsg, correctedMonthlyIncome, correctedDailyIncome)
			}
		}
	}
	return plannedIncomeMsg
}
