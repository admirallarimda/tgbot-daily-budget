package bot

import "log"
import "fmt"
import "time"
import "sort"
import "math"
import "gopkg.in/telegram-bot-api.v4"

import "../budget"

type ownerReminder struct {
    t time.Time
    ownerId budget.OwnerId
}

type dailyReminder struct {
    baseHandler
}

func (d *dailyReminder) register(out_msg_chan chan<- tgbotapi.MessageConfig,
                                 service_chan chan<- serviceMsg) handlerTrigger {
    d.out_msg_chan = out_msg_chan

    d.storageconn = budget.CreateStorageConnection()

    return handlerTrigger{}
}

func processDailyReminders(reminders []ownerReminder, now time.Time) (newReminders []ownerReminder, ownersToBeNotified []budget.OwnerId) {
    ownersToBeNotified = make([]budget.OwnerId, 0, 0)

    sort.Slice(reminders, func(x, y int) bool { return reminders[x].t.Before(reminders[y].t)})
    lastNotifIx := -1
    for i, reminder := range reminders {
        t := reminder.t
        if t.After(now) {
            log.Printf("Will wait for reminder times finished at %s", t)
            break
        }
        lastNotifIx = i

        log.Printf("Need to send daily notifications for user %d with notification time at %s", reminder.ownerId, t)
        ownersToBeNotified = append(ownersToBeNotified, reminder.ownerId)

        nextNotifTime := t.Add(time.Duration(24) * time.Hour)
        reminders = append(reminders, ownerReminder{t: nextNotifTime, ownerId: reminder.ownerId})
    }

    if lastNotifIx == -1 {
        newReminders = reminders
    } else {
        newReminders = reminders[lastNotifIx + 1:]
    }
    return
}

func (d *dailyReminder) sendDailyAvailableBalance(owner budget.OwnerId, wallet *budget.Wallet, t time.Time) {
    log.Printf("Sending daily available balance to owner %d with wallet '%s'", owner, wallet.ID)
    availMoney, err := wallet.GetBalance(t)
    if err != nil {
        log.Printf("Could not get balance for wallet '%s' due to error: %s", wallet.ID, err)
        return
    }
    monthSplit := budget.SplitWalletMonth(t, wallet.MonthStart)
    msg := fmt.Sprintf("New day has come! Currently available money: %d; there are %d days till month end", availMoney, monthSplit.DaysRemaining)
    if availMoney > 0 {
        d.out_msg_chan<- tgbotapi.NewMessage(int64(owner), msg)
    } else {
        // TODO: consider not only planned, but 'actual' income for current month
        plannedIncome, err := wallet.GetPlannedMonthlyIncome()
        if err != nil {
            log.Printf("Could not get planned income for wallet '%s' due to error: %s", wallet.ID, err)
            d.out_msg_chan<- tgbotapi.NewMessage(int64(owner), msg)
            return
        }
        plannedDailyIncome := plannedIncome / monthSplit.DaysInCurMonth
        daysTillPositive := int(math.Ceil(math.Abs(float64(availMoney) / float64(plannedDailyIncome))))
        d.out_msg_chan<- tgbotapi.NewMessage(int64(owner), fmt.Sprintf("%s\n In order to make positive balance with current daily income %d, you should not spend any money for %d days", msg, plannedDailyIncome, daysTillPositive))
    }
}

func (d *dailyReminder) sendMonthlySummary(owner budget.OwnerId, wallet *budget.Wallet, t time.Time) {
    log.Printf("Sending monthly stats to owner %d with wallet '%s'", owner, wallet.ID)
    summary, err := wallet.GetMonthlySummary(t.Add(time.Duration(time.Hour * -24)))
    if err != nil {
        return
    }

    type keyValue struct {
        key string
        value int
    }
    var sortedExpenses []keyValue
    for k, v := range summary.ExpenseSummary {
        sortedExpenses = append(sortedExpenses, keyValue{key: k, value: v})
    }
    sort.Slice(sortedExpenses, func(i, j int) bool {
                            return sortedExpenses[i].value < sortedExpenses[j].value // lowest value will be the first
    })

    msg := fmt.Sprintf("Last month summary (for dates from %s to %s):", summary.TimeStart, summary.TimeEnd)
    for _, kv := range sortedExpenses {
        label_txt := "unlabeled category"
        if kv.key != "" {
            label_txt = fmt.Sprintf("category labeled '%s'", kv.key)
        }
        msg = fmt.Sprintf("%s\nSpent %d for %s", msg, -(kv.value), label_txt)
    }

    d.out_msg_chan<- tgbotapi.NewMessage(int64(owner), msg)

    d.sendDailyAvailableBalance(owner, wallet, t)
}


func (d *dailyReminder) run() {
    ownerDataMap, err := d.storageconn.GetAllOwners()
    if err != nil {
        log.Panicf("Cannot start daily reminder as it is impossible to get owner data due to error: %s", err)
    }

    log.Printf("Starting daily reminder using a map of %d wallet owners", len(ownerDataMap))

    reminders := make([]ownerReminder, 0, len(ownerDataMap))
    // preparing structures for sorted reminders
    now := time.Now()
    startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
    for id, data := range ownerDataMap {
        if data.DailyReminderTime == nil {
            log.Printf("Owner %d doesn't have reminder settings; it will not be added into notificaiton list", id)
            continue
        }
        reminderTime := startOfDay.Add(*data.DailyReminderTime)
        reminders = append(reminders, ownerReminder{t: reminderTime, ownerId: id})
    }
    log.Printf("Running one 'fake' daily reminder processing in order to skip all reminders for current day")
    reminders, _ = processDailyReminders(reminders, time.Now())

    // main notif cycle
    for {
        checkTime := time.Now()
        var ownersToBeNotified []budget.OwnerId
        reminders, ownersToBeNotified = processDailyReminders(reminders, checkTime)
        for _, owner := range ownersToBeNotified {
            wallet, err := budget.GetWalletForOwner(owner, false, d.storageconn)
            if err != nil {
                log.Printf("Could not get wallet for owner %d with error: %s", owner, err)
                continue
            }
            if wallet.MonthStart == checkTime.Day() {
                d.sendMonthlySummary(owner, wallet, checkTime)
            } else {
                d.sendDailyAvailableBalance(owner, wallet, checkTime)
            }
        }
        time.Sleep(time.Minute)
    }
}
