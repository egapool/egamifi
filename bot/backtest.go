package bot

type Backtest struct {
	bot BotTest
}

type BotTest interface {
	InitBot()
	HandleBacktest(t, side, size, price, liquidation string)
	Result()
	ResultOneline()
}

func NewBacktest(bot BotTest) Backtest {
	return Backtest{
		bot: bot,
	}
	// using test-mode client
	// logger?
}

func (t *Backtest) Run(trades [][]string, is_combination bool) {
	func() {
		for _, line := range trades {
			t.bot.HandleBacktest(line[1], line[2], line[3], line[4], line[5])
		}
		if is_combination {
			t.bot.ResultOneline()
		} else {
			t.bot.Result()
		}
	}()
}
