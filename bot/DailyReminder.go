package bot

import "log"

import "fmt"
import "time"

import "math"
import "gopkg.in/telegram-bot-api.v4"

import "github.com/admirallarimda/tgbot-daily-budget/budget"
import "github.com/admirallarimda/tgbotbase"

type ownerReminder struct {
	t       time.Time
	ownerId budget.OwnerId
}

type dailyReminderHandler struct {
	baseHandler
	cron tgbotbase.Cron
}

func NewDailyReminder(storage budget.Storage) tgbotbase.BackgroundMessageHandler {
	r := &dailyReminderHandler{}
	r.storage = storage
	return r
}

func (d *dailyReminderHandler) Init(outMsgCh chan<- tgbotapi.MessageConfig, srvCh chan<- tgbotbase.ServiceMsg) {
	d.OutMsgCh = outMsgCh
}

func (d *dailyReminderHandler) Name() string {
	return "daily reminder"
}

func (d *dailyReminderHandler) Run() {
	d.cron = tgbotbase.NewCron()
	d.initialLoad()
	// TODO: add channels for add/remove of new reminders
}

func (d *dailyReminderHandler) initialLoad() {
	ownerDataMap, err := d.storage.GetAllOwners()
	if err != nil {
		log.Panicf("Cannot start daily reminder as it is impossible to get owner data due to error: %s", err)
	}
	log.Printf("Starting daily reminder using a map of %d wallet owners", len(ownerDataMap))

	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	for id, data := range ownerDataMap {
		if data.DailyReminderTime == nil {
			log.Printf("Owner %d doesn't have reminder settings; it will not be added into notificaiton list", id)
			continue
		}
		reminderTime := startOfDay.Add(*data.DailyReminderTime)
		if reminderTime.Before(now) {
			reminderTime = reminderTime.Add(24 * time.Hour)
		}
		job := &dailyReminderJob{}
		job.OutMsgCh = d.OutMsgCh
		job.storage = d.storage
		job.ownerID = id
		job.ownerData = data
		d.cron.AddJob(reminderTime, job)
	}
}

type dailyReminderJob struct {
	baseHandler
	ownerID   budget.OwnerId
	ownerData budget.OwnerData
}

func (job *dailyReminderJob) Do(scheduledWhen time.Time, cron tgbotbase.Cron) {
	wallet, err := budget.GetWalletForOwner(job.ownerID, false, job.storage)
	if err != nil {
		log.Printf("Could not get wallet for owner %d with error: %s", job.ownerID, err)
		return
	}

	replies := make([]string, 0, 2)
	if wallet.MonthStart == scheduledWhen.Day() {
		if reply, err := prepareMonthlySummary(job.ownerID, wallet, scheduledWhen.Add(time.Hour*-24)); err == nil && len(reply) > 0 {
			replies = append(replies, reply)
		}
	}
	if dailyMsgs := prepareDailyNotification(job.ownerID, wallet, scheduledWhen, job.ownerData); len(dailyMsgs) > 0 {
		replies = append(replies, dailyMsgs...)
	}

	for _, txt := range replies {
		job.OutMsgCh <- tgbotapi.NewMessage(int64(job.ownerID), txt)
	}

	nextNotifTime := scheduledWhen.Add(time.Duration(24) * time.Hour)
	cron.AddJob(nextNotifTime, job)
}

func prepareDailyNotification(owner budget.OwnerId, wallet *budget.Wallet, t time.Time, ownerData budget.OwnerData) []string {
	log.Printf("Preparing daily available balance to owner %d with wallet '%s'", owner, wallet.ID)
	msgs := make([]string, 0, 3)
	availMoney, err := wallet.GetBalance(t)
	if err != nil {
		log.Printf("Could not get balance for wallet '%s' due to error: %s", wallet.ID, err)
		return msgs
	}
	monthSplit := budget.SplitWalletMonth(t, wallet.MonthStart)
	msgs = append(msgs, fmt.Sprintf("New day has come! Currently available money: %d; there are %d days till month end", availMoney, monthSplit.DaysRemaining))
	if availMoney < 0 {
		// TODO: consider not only planned, but 'actual' income for current month
		plannedIncome, err := wallet.GetPlannedMonthlyIncome()
		if err != nil {
			log.Printf("Could not get planned income for wallet '%s' due to error: %s", wallet.ID, err)
			return msgs
		}
		plannedDailyIncome := plannedIncome / monthSplit.DaysInCurMonth
		daysTillPositive := int(math.Ceil(math.Abs(float64(availMoney) / float64(plannedDailyIncome))))
		msgs = append(msgs, fmt.Sprintf("In order to make positive balance with current daily income %d, you should not spend any money for %d days", plannedDailyIncome, daysTillPositive))
	}

	log.Printf("Checking and sending reminders for regular transactions for current day for owner %d with wallet '%s' (has %d dates for reminding)", owner, wallet.ID, len(ownerData.RegularTxs))
	if txs, found := ownerData.RegularTxs[t.Day()]; found {
		msg := fmt.Sprintf("You have the following regular transactions to be fulfilled today:")
		for _, tx := range txs {
			msg = fmt.Sprintf("%s\n%d labeled by '%s'", msg, tx.Value, tx.Label)
		}
		msgs = append(msgs, msg)
	}

	return msgs
}
