package mmb

import (
	"context"
	"fmt"
	"os"

	"github.com/go-numb/go-ftx/realtime"
	"github.com/go-numb/go-ftx/rest/private/fills"
	"github.com/go-numb/go-ftx/rest/public/markets"
)

type Mmb struct {
	market         string
	feeRate        float64
	priceIncrement float64
}

func NewMmb(market string) *Mmb {
	return &Mmb{
		market:         market,
		feeRate:        0.000388,
		priceIncrement: 0.001,
	}
}

func (m *Mmb) Run() {
	m.websocketRun()
}

func (m *Mmb) HandleTicker(ticker markets.Ticker, bid, ask float64) (float64, float64) {
	if ask != ticker.Ask || bid != ticker.Bid {
		ask = ticker.Ask
		bid = ticker.Bid
		rate := (ask - bid) / bid
		rate = rate - m.feeRate
		rate *= 100
		fmt.Printf("%s	%.5f%% Bid: %.5f, Ask: %.5f, BidUSD: %.3f, AskUSD: %.3f, %+v\n", m.market, rate, ticker.Bid, ticker.Ask, bid*ticker.BidSize, ask*ticker.AskSize, ticker)
		if rate > 0 {

		}
		return bid, ask
	}
	return bid, ask
}

func (m *Mmb) HandleExecution(fills fills.Fill) {

}

func (m *Mmb) websocketRun() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan realtime.Response)
	go realtime.Connect(ctx, ch, []string{"ticker", "trades"}, []string{m.market}, nil)

	var ask float64
	var bid float64
	for {
		select {
		case v := <-ch:
			switch v.Type {
			case realtime.TICKER:
				bid, ask = m.HandleTicker(v.Ticker, bid, ask)

			case realtime.TRADES:
				fmt.Printf("%s	%+v\n", v.Symbol, v.Trades)
				for i := range v.Trades {
					if v.Trades[i].Liquidation {
						fmt.Printf("-----------------------------%+v\n", v.Trades[i])
					}
				}

			case realtime.ORDERBOOK:
				fmt.Printf("%s	%+v\n", v.Symbol, v.Orderbook)

			case realtime.UNDEFINED:
				fmt.Printf("UNDEFINED %s	%s\n", v.Symbol, v.Results.Error())
			}
		}
	}
}

func (m *Mmb) privateWebsocketRun() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan realtime.Response)
	go realtime.ConnectForPrivate(ctx, ch, os.Getenv("FTX_KEY"), os.Getenv("FTX_SECRET"), []string{"fills"}, nil, os.Getenv("FTX_SUBACCOUNT"))

	for {
		select {
		case v := <-ch:
			switch v.Type {
			case realtime.FILLS:
				fmt.Printf("%s	%+v\n", "FILLS", v.Fills)
				m.HandleExecution(v.Fills)

			case realtime.UNDEFINED:
				fmt.Printf("UNDEFINED %s	%s\n", v.Symbol, v.Results.Error())
			}
		}
	}

}
