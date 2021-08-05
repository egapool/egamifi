package bot

import (
	"context"
	"fmt"
	"time"

	"github.com/go-numb/go-ftx/realtime"
)

type Runner struct {
	bot Bot
}

type Bot interface {
	Market() string
	InitBot()
	Handle(t time.Time, side string, size, price float64, liquidation bool)
}

func NewRunner(bot Bot) Runner {
	return Runner{
		bot: bot}
}

func (r *Runner) Run() {
	r.bot.InitBot()

	// websocketで取得
	// DataFlowInterfaceとして別の場所で実装いれるひつようがある
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan realtime.Response)
	go realtime.Connect(ctx, ch, []string{"trades"}, []string{r.bot.Market()}, nil)

	for {
		select {
		case v := <-ch:
			switch v.Type {
			case realtime.TRADES:
				for _, trade := range v.Trades {
					fmt.Printf("%s	%+v\n", v.Symbol, trade)
					r.bot.Handle(trade.Time, trade.Side, trade.Size, trade.Price, trade.Liquidation)
				}

			case realtime.ORDERBOOK:
				fmt.Printf("%s	%+v\n", v.Symbol, v.Orderbook)

			case realtime.UNDEFINED:
				fmt.Printf("UNDEFINED %s	%s\n", v.Symbol, v.Results.Error())
			}
		}
	}

}
