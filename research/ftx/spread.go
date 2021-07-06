package ftx

import (
	"context"
	"fmt"

	"github.com/go-numb/go-ftx/realtime"
	"github.com/go-numb/go-ftx/rest/public/markets"
)

type SpreadUsecase struct {
	market string
}

func NewSpreadUsecase(market string) *SpreadUsecase {
	return &SpreadUsecase{market: market}
}

func (s *SpreadUsecase) Run() {
	s.websocketRun()
}

func (s *SpreadUsecase) HandleTicker(ticker markets.Ticker, bid, ask float64) (float64, float64) {
	fee_rate := 0.000388
	if ask != ticker.Ask || bid != ticker.Bid {
		ask = ticker.Ask
		bid = ticker.Bid
		rate := (ask - bid) / bid
		rate = rate - fee_rate
		rate *= 100
		fmt.Printf("%s	%.5f%% Bid: %.5f, Ask: %.5f, BidUSD: %.3f, AskUSD: %.3f, %+v\n", s.market, rate, ticker.Bid, ticker.Ask, bid*ticker.BidSize, ask*ticker.AskSize, ticker)
		return bid, ask
	}
	return bid, ask
}

func (s *SpreadUsecase) websocketRun() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan realtime.Response)
	go realtime.Connect(ctx, ch, []string{"ticker", "trades"}, []string{s.market}, nil)

	var ask float64
	var bid float64
	for {
		select {
		case v := <-ch:
			switch v.Type {
			case realtime.TICKER:
				bid, ask = s.HandleTicker(v.Ticker, bid, ask)

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
