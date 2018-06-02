package budget

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
