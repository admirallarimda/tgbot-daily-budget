package budget

/*
import "testing"
import "strconv"
import "time"
import "math/rand"

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

func TestPlannedMonthlyIncome_IncomeOnly_NoActualCorrelation(t *testing.T) {
    s := NewRamStorage()
    w, err := s.CreateWalletOwner(OwnerId(1))
    if err != nil {
        t.FailNow()
    }

    incomeN := 1 + rand.Int() % 5
    totalPlanned := 0
    for i := 0; i < incomeN; i++ {
        val := 1000 + rand.Int() % 10000
        totalPlanned += val
        regular := NewRegularTransaction(val, i + 1, strconv.Itoa(i + 1))
        err := w.AddRegularTransaction(*regular)
        if err != nil {
            t.FailNow()
        }
    }
    if val, err := w.GetPlannedMonthlyIncome(); val != totalPlanned || err != nil {
        t.FailNow()
    }
}

func TestPlannedMonthlyIncome_IncomeAndExpense_NoActualCorrelation(t *testing.T) {
    s := NewRamStorage()
    w, err := s.CreateWalletOwner(OwnerId(1))
    if err != nil {
        t.FailNow()
    }

    transactionsN := 1 + rand.Int() % 5
    totalPlanned := 0
    for i := 0; i < transactionsN; i++ {
        val_pos := 1000 + rand.Int() % 10000
        totalPlanned += val_pos
        regular_pos := NewRegularTransaction(val_pos, i + 1, strconv.Itoa(i + 1) + "pos")
        err := w.AddRegularTransaction(*regular_pos)
        if err != nil {
            t.FailNow()
        }

        val_neg := -1 * (1000 + rand.Int() % 10000)
        totalPlanned += val_neg
        regular_neg := NewRegularTransaction(val_neg, i + 1, strconv.Itoa(i + 1) + "neg")
        err = w.AddRegularTransaction(*regular_neg)
        if err != nil {
            t.FailNow()
        }
    }
    if val, err := w.GetPlannedMonthlyIncome(); val != totalPlanned || err != nil {
        t.FailNow()
    }
}

func TestPlannedMonthlyIncome_IncomeAndExpense_FullActualCorrelation(t *testing.T) {
    s := NewRamStorage()
    w, err := s.CreateWalletOwner(OwnerId(1))
    if err != nil {
        t.FailNow()
    }

    trRegPos := NewRegularTransaction(10000, 1, "pos1")
    trRegNeg := NewRegularTransaction(-3300, 5, "neg1")
    totalPlanned := trRegPos.Value + trRegNeg.Value

    if w.AddRegularTransaction(*trRegPos) != nil || w.AddRegularTransaction(*trRegNeg) != nil {
        t.FailNow()
    }

    trActualPos := NewActualTransaction(8800, time.Now(), "pos1", "")
    trActualNeg := NewActualTransaction(-1200, time.Now(), "neg1", "")
    totalActual := trActualPos.Value + trActualNeg.Value

    if w.AddTransaction(*trActualPos) != nil ||
       w.AddTransaction(*trActualNeg) != nil {
           t.FailNow()
       }

    if val, err := w.GetPlannedMonthlyIncome(); val != totalPlanned || err != nil {
        t.FailNow()
    }
    if val, err := w.GetActualMonthlyIncome(); val != totalActual || err != nil {
        t.FailNow()
    }
}

func TestPlannedMonthlyIncome_IncomeAndExpense_PartialActualCorrelation(t *testing.T) {
    s := NewRamStorage()
    w, err := s.CreateWalletOwner(OwnerId(1))
    if err != nil {
        t.FailNow()
    }

    trRegPos := NewRegularTransaction(10000, 1, "pos1")
    trRegNeg := NewRegularTransaction(-3300, 5, "neg1")
    totalPlanned := trRegPos.Value + trRegNeg.Value

    if w.AddRegularTransaction(*trRegPos) != nil || w.AddRegularTransaction(*trRegNeg) != nil {
        t.FailNow()
    }

    trActualPos := NewActualTransaction(8800, time.Now(), "pos1", "")
    totalActual := trActualPos.Value + trRegNeg.Value
    if w.AddTransaction(*trActualPos) != nil {
           t.FailNow()
       }

    if val, err := w.GetPlannedMonthlyIncome(); val != totalPlanned || err != nil {
       t.FailNow()
    }
    if val, err := w.GetActualMonthlyIncome(); val != totalActual || err != nil {
       t.FailNow()
    }
}

func TestPlannedMonthlyIncome_IncomeAndExpense_FullActualCorrelation_AdditionalIncomeExpenseSameLabel(t *testing.T) {
    s := NewRamStorage()
    w, err := s.CreateWalletOwner(OwnerId(1))
    if err != nil {
        t.FailNow()
    }

    trRegPos := NewRegularTransaction(10000, 1, "pos1")
    trRegNeg := NewRegularTransaction(-3300, 5, "neg1")
    totalPlanned := trRegPos.Value + trRegNeg.Value

    if w.AddRegularTransaction(*trRegPos) != nil || w.AddRegularTransaction(*trRegNeg) != nil {
        t.FailNow()
    }

    trActualPos1 := NewActualTransaction(8800, time.Now(), "pos1", "")
    trActualPos2 := NewActualTransaction(500, time.Now(), "pos1", "")
    trActualNeg1 := NewActualTransaction(-1200, time.Now(), "neg1", "")
    trActualNeg2 := NewActualTransaction(-200, time.Now(), "neg1", "")
    totalActual := trActualPos1.Value + trActualPos2.Value + trActualNeg1.Value + trActualNeg2.Value

    if w.AddTransaction(*trActualPos1) != nil ||
       w.AddTransaction(*trActualPos2) != nil ||
       w.AddTransaction(*trActualNeg1) != nil ||
       w.AddTransaction(*trActualNeg2) != nil {
           t.FailNow()
       }

    if val, err := w.GetPlannedMonthlyIncome(); val != totalPlanned || err != nil {
      t.FailNow()
    }
    if val, err := w.GetActualMonthlyIncome(); val != totalActual || err != nil {
      t.FailNow()
    }
}

func TestPlannedMonthlyIncome_IncomeAndExpense_FullActualCorrelation_AdditionalIncomeOtherLabel(t *testing.T) {
    s := NewRamStorage()
    w, err := s.CreateWalletOwner(OwnerId(1))
    if err != nil {
        t.FailNow()
    }

    trRegPos := NewRegularTransaction(10000, 1, "pos1")
    trRegNeg := NewRegularTransaction(-3300, 5, "neg1")
    totalPlanned := trRegPos.Value + trRegNeg.Value

    if w.AddRegularTransaction(*trRegPos) != nil || w.AddRegularTransaction(*trRegNeg) != nil {
        t.FailNow()
    }

    trActualPos1 := NewActualTransaction(8800, time.Now(), "pos1", "")
    trActualPos2 := NewActualTransaction(500, time.Now(), "pos2", "")
    trActualNeg1 := NewActualTransaction(-1200, time.Now(), "neg1", "")
    trActualNeg2 := NewActualTransaction(-200, time.Now(), "neg2", "")
    totalActual := trActualPos1.Value + trActualPos2.Value + trActualNeg1.Value // trActualNeg2.Value is not used for income calc

    if w.AddTransaction(*trActualPos1) != nil ||
       w.AddTransaction(*trActualPos2) != nil ||
       w.AddTransaction(*trActualNeg1) != nil ||
       w.AddTransaction(*trActualNeg2) != nil {
           t.FailNow()
       }

    if val, err := w.GetPlannedMonthlyIncome(); val != totalPlanned || err != nil {
        t.FailNow()
    }
    if val, err := w.GetActualMonthlyIncome(); val != totalActual || err != nil {
        t.FailNow()
    }
}

func TestAvailableAmount_NoTransactions(t * testing.T) {
    s := NewRamStorage()
    w, err := s.CreateWalletOwner(OwnerId(1))
    if err != nil {
        t.FailNow()
    }

    val, err := w.GetActualMonthlyIncomeTillDate(time.Now())
    if err != nil {
        t.FailNow()
    }
    if val != 0 {
        t.Error(val)
    }
}

func TestAvailableAmount_OnlyRegularTransactions(t *testing.T) {
    s := NewRamStorage()
    w, err := s.CreateWalletOwner(OwnerId(1))
    if err != nil {
        t.FailNow()
    }

    trRegPos := NewRegularTransaction(10000, 1, "pos1")
    trRegNeg := NewRegularTransaction(-3300, 5, "neg1")
    totalPlanned := trRegPos.Value + trRegNeg.Value

    if w.AddRegularTransaction(*trRegPos) != nil || w.AddRegularTransaction(*trRegNeg) != nil {
        t.FailNow()
    }

    t.Log("HERE STARTS a test for 31 days")
    t_days31 := time.Date(2018, 1, 10, 0, 0, 0, 0, time.UTC)
    val31, err := w.GetActualMonthlyIncomeTillDate(t_days31)
    if err != nil {
        t.FailNow()
    }
    expected_val31 := int(float32(totalPlanned) / 31 * 10) // 10 days from 1 (default monthStart) to t_days31
    if val31 != expected_val31 {
        t.Errorf("31 days: actual=%d; expected=%d", val31, expected_val31)
    }

    t.Log("HERE STARTS a test for 30 days")
    t_days30 := time.Date(2018, 4, 3, 0, 0, 0, 0, time.UTC)
    val30, err := w.GetActualMonthlyIncomeTillDate(t_days30)
    if err != nil {
        t.FailNow()
    }
    expected_val30 := int(float32(totalPlanned) / 30 * 3)
    if val30 != expected_val30 {
        t.Errorf("30 days: actual=%d; expected=%d", val30, expected_val30)
    }

    t.Log("HERE STARTS a test for 28 days")
    t_days28 := time.Date(2018, 2, 20, 0, 0, 0, 0, time.UTC)
    val28, err := w.GetActualMonthlyIncomeTillDate(t_days28)
    if err != nil {
        t.FailNow()
    }
    expected_val28 := int(float32(totalPlanned) / 28 * 20)
    if val28 != expected_val28 {
        t.Errorf("28 days: actual=%d; expected=%d", val28, expected_val28)
    }

    // TODO: correct this test when leap year is handled correctly
    t.Log("HERE STARTS a test for 29 (leap year) days")
    t_days29 := time.Date(2004, 2, 20, 0, 0, 0, 0, time.UTC)
    val29, err := w.GetActualMonthlyIncomeTillDate(t_days29)
    if err != nil {
        t.FailNow()
    }
    expected_val29 := expected_val28 // to be corrected when leap year is handled correctly
    if val29 != expected_val29 {
        t.Errorf("29 days: actual=%d; expected=%d", val29, expected_val29)
    }
}

func TestAvailableAmount_RegularThenActual(t *testing.T) {
    s := NewRamStorage()
    w, err := s.CreateWalletOwner(OwnerId(time.Now().Unix()))
    if err != nil {
        t.FailNow()
    }

    trRegPos := NewRegularTransaction(10000, 1, "pos1")
    trRegNeg := NewRegularTransaction(-3300, 5, "neg1")
    totalPlanned := float32(trRegPos.Value + trRegNeg.Value)

    if w.AddRegularTransaction(*trRegPos) != nil || w.AddRegularTransaction(*trRegNeg) != nil {
        t.FailNow()
    }

    t1 := time.Date(2018, 06, 20, 0, 0, 0, 0, time.UTC)
    var daysInJune float32 = 30
    trActual1 := NewActualTransaction(-500, t1, "food", "")
    if w.AddTransaction(*trActual1) != nil {
        t.FailNow()
    }

    tBefore := t1.Add(time.Duration(time.Hour * (-5)))
    valBefore, err := w.GetBalance(tBefore)
    if err != nil {
        t.FailNow()
    }
    expectedValBefore := int(totalPlanned / daysInJune * float32(tBefore.Day()))
    if valBefore != expectedValBefore {
        t.Errorf("BEFORE mismatch: actual=%d; expected=%d", valBefore, expectedValBefore)
    }

    tAfter := t1.Add(time.Duration(time.Hour * 2))
    valAfter, err := w.GetBalance(tAfter)
    if err != nil {
        t.FailNow()
    }
    expectedValAfter := int(totalPlanned / daysInJune * float32(tAfter.Day())) - 500
    if valAfter != expectedValAfter {
        t.Errorf("AFTER mismatch: actual=%d; expected=%d", valAfter, expectedValAfter)
    }

    tExact := t1
    valExactTime, err := w.GetBalance(tExact)
    if err != nil {
        t.FailNow()
    }
    expectedValExact := expectedValAfter
    if valExactTime != expectedValExact {
        t.Errorf("EXACT mismatch: actual=%d; expected=%d", valExactTime, expectedValExact)
    }
}

func TestAvailableAmount_CorrectionAfterNewRegular(t *testing.T) {
    s := NewRamStorage()
    w, err := s.CreateWalletOwner(OwnerId(time.Now().Unix()))
    if err != nil {
        t.FailNow()
    }

    trRegPos := NewRegularTransaction(3000, 1, "pos1")
    totalPlanned := float32(trRegPos.Value)

    if w.AddRegularTransaction(*trRegPos) != nil {
        t.FailNow()
    }

    t1 := time.Date(2018, 06, 20, 0, 0, 0, 0, time.UTC)
    var daysInJune float32 = 30
    trActual1 := NewActualTransaction(-800, t1, "food", "")
    if w.AddTransaction(*trActual1) != nil {
        t.FailNow()
    }

    tAfter := t1.Add(time.Duration(time.Hour * 2))
    valAfter, err := w.GetBalance(tAfter)
    if err != nil {
        t.FailNow()
    }
    expectedValAfter := int(totalPlanned / daysInJune * float32(tAfter.Day())) + trActual1.Value
    if valAfter != expectedValAfter {
        t.Errorf("AFTER mismatch: actual=%d; expected=%d", valAfter, expectedValAfter)
    }

    trRegNeg := NewRegularTransaction(-3300, 5, "neg1")
    totalPlanned += float32(trRegNeg.Value)

    if w.AddRegularTransaction(*trRegNeg) != nil {
        t.FailNow()
    }

    valAfter, err = w.GetBalance(tAfter)
    if err != nil {
        t.FailNow()
    }
    expectedValAfter = int(totalPlanned / daysInJune * float32(tAfter.Day())) + trActual1.Value
    if valAfter != expectedValAfter {
        t.Errorf("AFTER mismatch: actual=%d; expected=%d", valAfter, expectedValAfter)
    }
}

func TestAvailableAmount_CorrectionAfterNewRegularWithLabelMatch(t *testing.T) {
    t.Skip("TODO: impleminent (add new regular transaction after actual, and new regular matches label of an actual transaction)")
}

func TestAvailableAmount_ModifiedMonthStart_January(t *testing.T) {
    t.Skip("TODO: implement. There's a hack for january, try monthStart 10, transactions and get amount on dates 1-10")
}
*/
