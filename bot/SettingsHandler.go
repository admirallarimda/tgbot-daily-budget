package bot

import "regexp"
import "log"
import "fmt"
import "time"
import "errors"
import "strconv"
import "strings"
import "gopkg.in/telegram-bot-api.v4"

import "github.com/admirallarimda/tgbot-daily-budget/budget"
import "github.com/admirallarimda/tgbotbase"

var monthStartRe *regexp.Regexp = regexp.MustCompile("monthStart (\\d{1,2})")
var notifTimeRe *regexp.Regexp = regexp.MustCompile("notifTime ((\\d{1,2}:\\d{2})|(disable))")

type settingsHandler struct {
	baseHandler
}

func NewWalletSettingsHandler(storage budget.Storage) tgbotbase.IncomingMessageHandler {
	h := &settingsHandler{}
	h.storage = storage
	return h
}

func (h *settingsHandler) Init(outMsgCh chan<- tgbotapi.Chattable, srvCh chan<- tgbotbase.ServiceMsg) tgbotbase.HandlerTrigger {
	h.OutMsgCh = outMsgCh
	return tgbotbase.NewHandlerTrigger(nil, []string{"settings"})
}

func (h *settingsHandler) Name() string {
	return "wallet settings"
}

func (h *settingsHandler) HandleOne(msg tgbotapi.Message) {
	text := msg.Text
	chatId := msg.Chat.ID
	ownerId := budget.OwnerId(chatId)
	h.parseCmd(text, chatId, ownerId)
}

func (h *settingsHandler) setMonthStart(ownerId budget.OwnerId, date int) error {
	if date < 1 || date > 28 {
		return errors.New("Date must be between 1 and 28")
	}

	wallet, err := budget.GetWalletForOwner(ownerId, true, h.storage)
	if err != nil {
		return err
	}

	err = wallet.SetMonthStart(date)
	if err != nil {
		return err
	}

	log.Printf("Month start for wallet '%s' has been successfully modified to %d", wallet.ID, date)
	return nil
}

func (h *settingsHandler) changeMonthStart(text string, chatId int64, ownerId budget.OwnerId) {
	matches := monthStartRe.FindStringSubmatch(text)
	dateStr := matches[1]
	date, err := strconv.Atoi(dateStr)
	if err != nil {
		log.Printf("Could not convert date '%s' for month start setting due to error: %s", dateStr, err)
		h.OutMsgCh <- tgbotapi.NewMessage(chatId, fmt.Sprintf("Incorrect format of date for month start modification"))
		return
	}
	err = h.setMonthStart(ownerId, date)
	if err != nil {
		log.Printf("Could not set month start %d for owner %d due to error: %s", date, ownerId, err)
		h.OutMsgCh <- tgbotapi.NewMessage(chatId, fmt.Sprintf("Could not set month start due to the following reason: %s", err))
		return
	}
	h.OutMsgCh <- tgbotapi.NewMessage(chatId, fmt.Sprintf("Month start has been successfully modified"))
}

func (h *settingsHandler) setNotificationTime(ownerId budget.OwnerId, enabled bool, hour, minute int) error {
	if (hour < 0 || hour > 23) || (minute < 0 || minute > 59) {
		return errors.New(fmt.Sprintf("Incorrect notification hour %d or minute %d", hour, minute))
	}

	var notifTime *time.Duration = nil
	if enabled {
		t := time.Duration(time.Hour*time.Duration(hour) + time.Minute*time.Duration(minute))
		notifTime = &t
	}

	return h.storage.SetOwnerDailyNotificationTime(ownerId, notifTime)
}

func (h *settingsHandler) changeNotificationTime(text string, chatId int64, ownerId budget.OwnerId) {
	matches := notifTimeRe.FindStringSubmatch(text)
	enabled := true
	hour, minute := 0, 0
	var err error
	if matches[1] == "disable" {
		enabled = false
	} else {
		timeParts := strings.Split(matches[1], ":")
		// len == 2 due to regexp, no need to check
		hour, err = strconv.Atoi(timeParts[0])
		if err != nil {
			log.Printf("Could not convert hour '%s' to int due to error: %s", timeParts[0], err)
			h.OutMsgCh <- tgbotapi.NewMessage(chatId, fmt.Sprintf("Hour part in requested notification time '%s' cannot be converted to integer", matches[1]))
			return
		}
		minute, err = strconv.Atoi(timeParts[1])
		if err != nil {
			log.Printf("Could not convert minute '%s' to int due to error: %s", timeParts[1], err)
			h.OutMsgCh <- tgbotapi.NewMessage(chatId, fmt.Sprintf("Minute part in requested notification time '%s' cannot be converted to integer", matches[1]))
			return
		}
	}
	err = h.setNotificationTime(ownerId, enabled, hour, minute)
	if err != nil {
		log.Printf("Something went wrong with disabling notification time for owner %d, error: %s", ownerId, err)
		h.OutMsgCh <- tgbotapi.NewMessage(chatId, "Something went wrong - cannot modify notification time :( ")
	} else {
		h.OutMsgCh <- tgbotapi.NewMessage(chatId, "Your daily notification settings have been successfully modified")
	}
}

func (h *settingsHandler) parseCmd(text string, chatId int64, ownerId budget.OwnerId) {
	if monthStartRe.MatchString(text) {
		h.changeMonthStart(text, chatId, ownerId)
	} else if notifTimeRe.MatchString(text) {
		h.changeNotificationTime(text, chatId, ownerId)
	}
}
