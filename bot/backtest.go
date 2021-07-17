package bot

import (
	"encoding/csv"
	"os"
)

type Backtest struct {
	bot Bot
}

type Bot interface {
	InitBot()
	Handle(t, side, size, price, liquidation string)
	Result()
}

func NewBacktest(bot Bot) Backtest {
	return Backtest{
		bot: bot,
	}
	// using test-mode client
	// logger?
}

func (t *Backtest) Run() {
	// loop by price data set
	// filepath := "../data/ftx-trades-20210715230700-20210715230000.csv"
	filepath := "../data/sample.csv"
	file, err := os.Open(filepath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	var line []string

	for {
		line, err = reader.Read()
		if err != nil {
			break
		}
		// run a strategy
		t.bot.Handle(line[1], line[2], line[3], line[4], line[5])
	}

	// 集計
	t.bot.Result()
}
