package bot

import "log"

import "fmt"
import "time"
import "sort"

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
	if wallet.MonthStart == scheduledWhen.Day() {
		job.sendMonthlySummary(job.ownerID, wallet, scheduledWhen)
	}
	job.sendDailyNotification(job.ownerID, wallet, scheduledWhen, job.ownerData)

	nextNotifTime := scheduledWhen.Add(time.Duration(24) * time.Hour)
	cron.AddJob(nextNotifTime, job)
}

func (job *dailyReminderJob) sendDailyNotification(owner budget.OwnerId, wallet *budget.Wallet, t time.Time, ownerData budget.OwnerData) {
	log.Printf("Sending daily available balance to owner %d with wallet '%s'", owner, wallet.ID)
	availMoney, err := wallet.GetBalance(t)
	if err != nil {
		log.Printf("Could not get balance for wallet '%s' due to error: %s", wallet.ID, err)
		return
	}
	monthSplit := budget.SplitWalletMonth(t, wallet.MonthStart)
	msg := fmt.Sprintf("New day has come! Currently available money: %d; there are %d days till month end", availMoney, monthSplit.DaysRemaining)
	if availMoney > 0 {
		job.OutMsgCh <- tgbotapi.NewMessage(int64(owner), msg)
	} else {
		// TODO: consider not only planned, but 'actual' income for current month
		plannedIncome, err := wallet.GetPlannedMonthlyIncome()
		if err != nil {
			log.Printf("Could not get planned income for wallet '%s' due to error: %s", wallet.ID, err)
			job.OutMsgCh <- tgbotapi.NewMessage(int64(owner), msg)
			return
		}
		plannedDailyIncome := plannedIncome / monthSplit.DaysInCurMonth
		daysTillPositive := int(math.Ceil(math.Abs(float64(availMoney) / float64(plannedDailyIncome))))
		job.OutMsgCh <- tgbotapi.NewMessage(int64(owner), fmt.Sprintf("%s\nIn order to make positive balance with current daily income %d, you should not spend any money for %d days", msg, plannedDailyIncome, daysTillPositive))
	}

	log.Printf("Checking and sending reminders for regular transactions for current day for owner %d with wallet '%s' (has %d dates for reminding)", owner, wallet.ID, len(ownerData.RegularTxs))
	if txs, found := ownerData.RegularTxs[t.Day()]; found {
		msg := fmt.Sprintf("You have the following regular transactions to be fulfilled today:")
		for _, tx := range txs {
			msg = fmt.Sprintf("%s\n%d labeled by '%s'", msg, tx.Value, tx.Label)
		}
		job.OutMsgCh <- tgbotapi.NewMessage(int64(owner), msg)
	}
}

func (job *dailyReminderJob) sendMonthlySummary(owner budget.OwnerId, wallet *budget.Wallet, t time.Time) {
	log.Printf("Sending monthly stats to owner %d with wallet '%s'", owner, wallet.ID)
	summary, err := wallet.GetMonthlySummary(t.Add(time.Duration(time.Hour * -24)))
	if err != nil {
		return
	}

	type keyValue struct {
		key   string
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

	job.OutMsgCh <- tgbotapi.NewMessage(int64(owner), msg)
}
