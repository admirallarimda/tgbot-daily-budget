package main

import "log"
import "gopkg.in/gcfg.v1"
import "github.com/admirallarimda/tgbotbase"

import "github.com/admirallarimda/tgbot-daily-budget/bot"
import "github.com/admirallarimda/tgbot-daily-budget/budget"

type config struct {
	tgbotbase.Config
	Redis tgbotbase.RedisConfig
}

func readGcfg(filename string) config {
	log.Printf("Reading configuration from: %s", filename)

	var cfg config

	err := gcfg.ReadFileInto(&cfg, filename)
	if err != nil {
		log.Printf("Could not correctly parse configuration file: %s; error: %s", filename, err)
		panic(err)
	}

	log.Printf("Configuration has been successfully read from %s: %+v", filename, cfg)
	return cfg
}

func main() {
	log.Print("Starting daily budget bot")

	cfg := readGcfg("bot.cfg")
	botCfg := tgbotbase.Config{TGBot: cfg.TGBot, Proxy_SOCKS5: cfg.Proxy_SOCKS5}
	tgbot := tgbotbase.NewBot(botCfg)

	pool := tgbotbase.NewRedisPool(cfg.Redis)

	tgbot.AddHandler(tgbotbase.NewIncomingMessageDealer(bot.NewTransactionHandler(budget.CreateStorageConnection(pool))))
	tgbot.AddHandler(tgbotbase.NewIncomingMessageDealer(bot.NewRegularTransactionHandler(budget.CreateStorageConnection(pool))))

	/*
		triggers = addHandler(&startHandler{}, "start", channels, triggers)

		triggers = addHandler(&regularTransactionHandler{}, "regular transactions", channels, triggers)
		triggers = addHandler(&dailyReminder{}, "daily wallet status notification", channels, triggers)
		triggers = addHandler(&settingsHandler{}, "wallet settings", channels, triggers)
		triggers = addHandler(&lastTransactionsListHandler{}, "list of last transactions", channels, triggers)
	*/

	tgbot.Start()

	log.Print("Daily budget bot has stopped")
}
