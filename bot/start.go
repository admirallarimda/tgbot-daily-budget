package bot

import "log"
import "gopkg.in/telegram-bot-api.v4"
import "golang.org/x/net/proxy"
import "net/http"

import "../botcfg"

// panics internally if something goes wrong
func setupBot(cfg botcfg.Config) (*tgbotapi.BotAPI, *tgbotapi.UpdatesChannel) {
    botToken := cfg.TGBot.Token
    log.Printf("Setting up a bot with token: %s", botToken)

    var bot *tgbotapi.BotAPI = nil
    server := cfg.Proxy_SOCKS5.Server
    user := cfg.Proxy_SOCKS5.User
    pass := cfg.Proxy_SOCKS5.Pass
    if server != "" {
        log.Printf("Proxy is set, connecting to '%s' with credentials '%s':'%s'", server, user, pass)
        auth := proxy.Auth { User: user,
                             Password: pass}
        dialer, err := proxy.SOCKS5("tcp", server, &auth, proxy.Direct)
        if err != nil {
            log.Panicf("Could get proxy dialer, error: %s", err)
        }
        httpTransport := &http.Transport{}
        httpTransport.Dial = dialer.Dial
        httpClient := &http.Client{Transport: httpTransport}
        bot, err = tgbotapi.NewBotAPIWithClient(botToken, httpClient)
        if err != nil {
            log.Panicf("Could not connect via proxy, error: %s", err)
        }
    } else {
        log.Printf("No proxy is set, going without any proxy")
        var err error
        bot, err = tgbotapi.NewBotAPI(botToken)
        if err != nil {
            log.Panicf("Could not connect directly, error: %s", err)
        }
    }

    log.Printf("Authorized on account %s", bot.Self.UserName)

    u := tgbotapi.NewUpdate(0)
    u.Timeout = 60

    updates, err := bot.GetUpdatesChan(u)
    if err != nil {
        log.Panic(err)
    }

    return bot, &updates
}

func run(updates *tgbotapi.UpdatesChannel, bot *tgbotapi.BotAPI, cfg botcfg.Config) {
    isRunning := true
    for isRunning {
        select {
            case update := <-*updates:
                log.Printf("Received an update from tgbotapi")
                if update.Message == nil {
                    log.Print("Message: empty. Skipping");
                    continue
                }
        }
    }

    log.Print("Main cycle has been aborted")
}

func Start(cfg botcfg.Config) error {
    log.Print("Starting the bot")

    bot, updates := setupBot(cfg);
    run(updates, bot, cfg)

    log.Print("Stopping the bot")
    return nil
}
