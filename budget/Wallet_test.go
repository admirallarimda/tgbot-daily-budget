package budget

import "testing"

func TestEmptyWalletPlannedIncome(t *testing.T) {
    s := NewRamStorage()
    w, err := s.CreateWalletOwner(OwnerId(1))
    if err != nil {
        t.FailNow()
    }
    val, err := w.GetPlannedMonthlyIncome()
    if err != nil {
        t.FailNow()
    }
    if val != 0 {
        t.FailNow()
    }
}
