package main

import (
	"context"
	"fmt"
	"time"

	"github.com/egapool/egamifi/internal/client"
	"github.com/egapool/egamifi/internal/notification"
	"github.com/egapool/egamifi/internal/strategy"
	"github.com/go-numb/go-ftx/realtime"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan realtime.Response)
	// go realtime.ConnectForPrivate(ctx, ch, "<key>", "<secret>", []string{"orders", "fills"}, nil)

	long_ask := 0.0
	long_ask_size := 0.0
	short_bid := 0.0
	short_bid_size := 0.0
	client := client.NewSubRestClient("shit")
	future := "TRU"
	crossorder := strategy.NewCrossOrder(client, 0.0535, future+"-0625", future+"-PERP")
	go realtime.Connect(ctx, ch, []string{"ticker"}, []string{crossorder.Long, crossorder.Short}, nil)
	for {
		select {
		case v := <-ch:
			switch v.Type {
			case realtime.TICKER:
				if v.Symbol == crossorder.Long {
					if long_ask == v.Ticker.Ask {
						continue
					}
					long_ask = v.Ticker.Ask
					long_ask_size = v.Ticker.AskSize
				}
				if v.Symbol == crossorder.Short {
					if short_bid == v.Ticker.Bid {
						continue
					}
					short_bid = v.Ticker.Bid
					short_bid_size = v.Ticker.BidSize
				}
				if short_bid == 0.0 || long_ask == 0.0 {
					continue
				}
				// diff := (long_ask - short_bid) * 2 / (long_ask + short_bid)
				diff := (long_ask - short_bid) / short_bid
				fmt.Printf("%.5f %.4f %s ask %.4f (%.3f) / %s bid %.4f (%.3f) %s\n", diff, long_ask-short_bid, crossorder.Long, long_ask, long_ask_size, crossorder.Short, short_bid, short_bid_size, time.Now().Format(time.UnixDate))

				if crossorder.ShouldEntery(diff) {
					var size float64
					notification.Notify(fmt.Sprintf("%f", diff), "general", "https://hooks.slack.com/services/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")
					if long_ask_size < short_bid_size {
						size = long_ask_size
					} else {
						size = short_bid_size
					}
					fmt.Println("entry", size)
					// size = 0.001
					crossorder.UpdateTicker(diff, long_ask, size)
					break
				}

				// fmt.Printf("%s	%+v\n", v.Symbol, v.Ticker)

			case realtime.TRADES:
				fmt.Printf("%s	%+v\n", v.Symbol, v.Trades)
				for i := range v.Trades {
					if v.Trades[i].Liquidation {
						fmt.Printf("-----------------------------%+v\n", v.Trades[i])
					}
				}

			case realtime.ORDERBOOK:
				fmt.Printf("%s	%+v\n", v.Symbol, v.Orderbook)

			case realtime.ORDERS:
				fmt.Printf("%d	%+v\n", v.Type, v.Orders)

			case realtime.FILLS:
				fmt.Printf("%d	%+v\n", v.Type, v.Fills)

			case realtime.UNDEFINED:
				fmt.Printf("UNDEFINED %s	%s\n", v.Symbol, v.Results.Error())
			}
		}
	}
}
