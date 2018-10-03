package main

import "log"
import "github.com/admirallarimda/tgbot-daily-budget/botcfg"
import "github.com/admirallarimda/tgbot-daily-budget/bot"

const cfg_filename = "bot.cfg"

func main() {
	cfg, err := botcfg.Read(cfg_filename)
	if err != nil {
		log.Fatalf("Could not read config file %s, exiting with error: %s", cfg_filename, err)
	}

	log.Printf("Starting the bot using config: %+v", cfg)
	bot.Start(cfg)
	log.Printf("Bot has finished working")
}
