[![Build Status](https://travis-ci.com/admirallarimda/tgbot-daily-budget.svg?branch=master)](https://travis-ci.com/admirallarimda/tgbot-daily-budget)
# tgbot-daily-budget
Telegram bot which follows 'daily available budget' approach

**TBD**: add link/description of this approach

## Available commands and operations
__/start__ is the initial entry point to start working with the bot. Though most other operations which involve writing some information into bot's wallets do the same preparing actions, this command will likely be expanded into a tutorial over how to work with the bot or a simple step-by-step wizard to setup initial configuration

__/regular__ allows to manage planned operations, like salaries, planned bills etc. Though usage of this command is not mandatory but it is essential for normal bot functioning and available money calculation

Each __/regular__ command must be followed by:
* type of a planned operation + amount. Type could be an '_income_' or an '_expense_'
* day of the month it typically occurs using '_date XX_' format. Note that only dates from 1 to 28 are accepted
* label in a '*#this_is_a_label*' format. Label is necessary to associate actual income/expense values over the planned ones (e.g. you're planning to get a salary of 1000 but for some reason you got only 950). Also note that each regular transaction must have a unique label

Therefore, a __/regular__ command might look like '_/regular income 1000 date 7 #salary_' or '_/regular expense 2000 #kindergaten date 16_'

If __/regular__ is issued without any arguments, it prints a list of all planned operations

If __/regular__ command has a '_delete_' keyword, then the transaction with this amount + date + label is removed.

**General transaction** could be added via simple '_AMOUNT_' or '_AMOUNT #somelabel_' statement. Here if **no sign** or '-' sign is used for AMOUNT, then this transaction is considered to be an expense. Only explicit '+' sign is considered to be an income.

When label is entered for a transaction, it is attempted to be matched to the planned incomes/expenses. **TBD description of matching rules**

__/last__ command allows to print N latest transactions. If N is omitted, it prints out 10 latest transactions

__/set__ command allows setting and removing various bot settings for current chat. The following options are available:
* monthStart instructs the bot in which date a new month should be started. Calculations for available money will consider this date as month start. By default equals to 1
