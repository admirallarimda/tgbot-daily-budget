package bot

import "log"
import "gopkg.in/telegram-bot-api.v4"
import "golang.org/x/net/proxy"
import "net/http"

import "../botcfg"
import "../budget"

type botChannels struct {
    out_msg_chan chan tgbotapi.MessageConfig
    service_chan chan serviceMsg
}

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

func addHandler(h msgHandler, name string, channels botChannels, triggers []handlerTrigger) []handlerTrigger {
    log.Printf("Preparing '%s' handler", name)
    triggers = append(triggers, h.register(channels.out_msg_chan, channels.service_chan))
    go h.run()
    return triggers
}

func setupHandlers(channels botChannels) []handlerTrigger {
    triggers := make([]handlerTrigger, 0, 10)

    triggers = addHandler(&startHandler{}, "start", channels, triggers)
    triggers = addHandler(&expenseHandler{}, "expense", channels, triggers)
    triggers = addHandler(&monthlyHandler{}, "monthly wallet settings", channels, triggers)

    return triggers
}

func setupStorage(cfg botcfg.Config) {
    redisServer := cfg.Redis.Server

    storage := budget.NewRedisStorage(redisServer)
    budget.SetStorage(storage)
}

func run(updates *tgbotapi.UpdatesChannel,
         bot *tgbotapi.BotAPI,
         cfg botcfg.Config,
         channels botChannels,
         handlers []handlerTrigger) {
    isRunning := true
    for isRunning {
        select {
            case update := <-*updates:
                log.Printf("Received an update from tgbotapi")
                if update.Message == nil {
                    log.Print("Message: empty. Skipping");
                    continue
                }
                for _, h := range handlers {
                    h.Handle(*update.Message)
                }
            case msg := <- channels.out_msg_chan:
                log.Printf("Received reply")
                _, err := bot.Send(msg)
                if err != nil {
                    log.Printf("Could not sent reply %+v due to error: %s", err)
                }
                continue
            case _ = <- channels.service_chan:
                log.Printf("Received service message")
                continue
        }
    }

    log.Print("Main cycle has been aborted")
}

func Start(cfg botcfg.Config) error {
    log.Print("Starting the bot")

    log.Printf("Setting up bot")
    bot, updates := setupBot(cfg)

    log.Printf("Setting up storage")
    setupStorage(cfg)

    replies := make(chan tgbotapi.MessageConfig, 0)
    serviceCh := make(chan serviceMsg, 0)
    channels := botChannels{
        out_msg_chan: replies,
        service_chan: serviceCh }

    log.Printf("Setting up handlers")
    handlers := setupHandlers(channels)

    log.Printf("Running the bot...")
    run(updates, bot, cfg, channels, handlers)

    log.Print("Stopping the bot")
    return nil
}
