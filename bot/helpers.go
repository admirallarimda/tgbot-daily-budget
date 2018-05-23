package bot

import "fmt"
import "log"
import "time"
import "gopkg.in/telegram-bot-api.v4"

import "../budget"

func dumpMsgUserInfo(msg tgbotapi.Message) string {
    return fmt.Sprintf("chat ID: %d (type '%s'), message issued by user ID: %d (username: '%s')", msg.Chat.ID,
                                                                                                  msg.Chat.Type,
                                                                                                  msg.From.ID,
                                                                                                  msg.From.UserName)
}

func getCurrentAvailableAmount(owner budget.OwnerId, t time.Time) (int, error) {
    log.Printf("Starting to calculate available amount for owner %d for time %s", owner, t)
    wallet, err := budget.GetStorage().GetWalletForOwner(owner)
    if err != nil {
        log.Printf("Could not get wallet for owner %d with error: %s", owner, err)
        return 0, err
    }

    // getting current available money
    curAvailIncome, err := budget.GetStorage().GetMonthlyIncomeTillDate(*wallet, t)
    if err != nil {
        log.Printf("Unable to get current available amount due to error: %s", err)
        return 0, err
    }

    curExpenses, err := budget.GetStorage().GetMonthlyExpenseTillDate(*wallet, t)
    if err != nil {
        log.Printf("Unable to get current expenses due to error: %s", err)
        return 0, err
    }
    availMoney := curAvailIncome - curExpenses
    log.Printf("Currently available money for owner %d: %d (income: %d; expenses: %d)", owner, availMoney, curAvailIncome, curExpenses)
    return availMoney, nil
}
