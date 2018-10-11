package bot

import "testing"
import "time"

func TestOneReminder(t *testing.T) {
	t1 := time.Now()
	times := make([]ownerReminder, 0, 1)
	times = append(times, ownerReminder{t: t1, ownerId: 200})

	t.Log("Checking notif for tBefore")
	tBefore := t1.Add(time.Minute * (-1))
	times2, ownersNotif := processDailyReminders(times, tBefore)
	if len(times2) != 1 || len(ownersNotif) != 0 {
		t.Errorf("Before time 't' test failed: len(times)=%d; len(ownerNotif)=%d", len(times2), len(ownersNotif))
	}

	t.Log("Checking notif for t1")
	times3, ownersNotif := processDailyReminders(times, t1)
	if len(times3) != 1 || len(ownersNotif) != 1 {
		t.Errorf("Time 't' test failed: len(times)=%d; len(ownerNotif)=%d", len(times3), len(ownersNotif))
	}

	t.Log("Checking notif for tAfter")
	tAfter := t1.Add(time.Minute)
	times4, ownersNotif := processDailyReminders(times, tAfter)
	if len(times4) != 1 || len(ownersNotif) != 1 {
		t.Errorf("After time 't' test failed: len(times)=%d; len(ownerNotif)=%d", len(times4), len(ownersNotif))
	}

	t.Log("Checking that next notif will happen after 24h - check at some time before")
	t2 := t1.Add(time.Hour * 24)
	t2Before := t2.Add(time.Minute * (-1))
	_, ownersNotif = processDailyReminders(times4, t2Before)
	if len(ownersNotif) != 0 {
		t.Errorf("24h 'before' test failed: len(ownerNotif)=%d", len(ownersNotif))
	}

	t.Log("Checking that next notif will happen after 24h")
	_, ownersNotif = processDailyReminders(times4, t2)
	if len(ownersNotif) != 1 {
		t.Errorf("24h test failed: len(ownerNotif)=%d", len(ownersNotif))
	}
}
