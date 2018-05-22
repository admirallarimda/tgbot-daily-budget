package bot

import "log"
import "fmt"
import "time"
import "sort"
import "gopkg.in/telegram-bot-api.v4"

import "../budget"

type dailyReminder struct {
    baseHandler
}

func (d *dailyReminder) register(out_msg_chan chan<- tgbotapi.MessageConfig,
                                 service_chan chan<- serviceMsg) handlerTrigger {
    d.out_msg_chan = out_msg_chan

    return handlerTrigger{}
}


func (d *dailyReminder) run() {
    ownerDataMap, err := budget.GetStorage().GetAllOwners()
    if err != nil {
        log.Panicf("Cannot start daily reminder as it is impossible to get owner data due to error: %s", err)
    }

    log.Printf("Starting daily reminder using a map of %d wallet owners", len(ownerDataMap))

    reminderTimes := make([]int64, 0, len(ownerDataMap))
    reminderTimeOwners := make(map[int64][]budget.OwnerId, len(ownerDataMap))
    // preparing structures for sorted reminders
    now := time.Now()
    startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
    for id, data := range ownerDataMap {
        durationFromMidnight := time.Duration(9) * time.Hour
        if data.DailyReminderTime != nil {
            durationFromMidnight = *data.DailyReminderTime
        }

        reminderTime := startOfDay.Add(durationFromMidnight)
        if reminderTime.Before(now) { // already happened this day - let's remind about next day
            reminderTime.Add(time.Duration(24) * time.Hour)
        }

        reminderTimeUnix := reminderTime.Unix()
        // TODO: check whether we already have this time
        reminderTimes = append(reminderTimes, reminderTimeUnix)
        owners := reminderTimeOwners[reminderTimeUnix]
        if owners == nil {
            owners = make([]budget.OwnerId, 0, 1)
        }
        owners = append(owners, id)
        reminderTimeOwners[reminderTimeUnix] = owners // TODO: do we need this operation in golang?
    }

    // main notif cycle
    for {
        sort.Slice(reminderTimes, func(x, y int) bool { return reminderTimes[x] < reminderTimes[y]})
        lastNotifIx := 0
        for i, t := range reminderTimes {
            lastNotifIx = i
            t1 := time.Unix(t, 0)
            if t1.After(now) {
                log.Printf("Will wait for reminder times finished at %s", t1)
                break
            }

            log.Printf("Sending daily notifications for users with notification time at %s", t1)
            for _, owner := range reminderTimeOwners[t] {
                // TODO: separate function?  owner -> Income minus Expences till date
                wallet, err := budget.GetStorage().GetWalletForOwner(owner)
                if err != nil {
                    log.Printf("Could not get wallet for owner %d with error: %s", owner, err)
                    continue
                }

                // getting current available money
                curAvailIncome, err := budget.GetStorage().GetMonthlyIncomeTillDate(*wallet, now)
                if err != nil {
                    log.Printf("Unable to get current available amount due to error: %s", err)
                    continue
                }

                curExpenses, err := budget.GetStorage().GetMonthlyExpenseTillDate(*wallet, now)
                if err != nil {
                    log.Printf("Unable to get current expenses due to error: %s", err)
                    continue
                }

                availMoney := curAvailIncome - curExpenses
                log.Printf("Currently available money: %d (income: %d; expenses: %d)", availMoney, curAvailIncome, curExpenses)
                d.out_msg_chan<- tgbotapi.NewMessage(int64(owner), fmt.Sprintf("Currently available money: %d", availMoney))
            }

            nextNotifTime := t1.Add(time.Duration(24) * time.Hour)
            reminderTimeOwners[nextNotifTime.Unix()] = reminderTimeOwners[t]
            delete(reminderTimeOwners, t)

        }
        reminderTimes = reminderTimes[lastNotifIx:]
        time.Sleep(time.Duration(1) * time.Minute)
    }
}
