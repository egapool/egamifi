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
	go realtime.Connect(ctx, ch, []string{"ticker"}, []string{"SHIT-0326", "SHIT-PERP"}, nil)
	// go realtime.ConnectForPrivate(ctx, ch, "<key>", "<secret>", []string{"orders", "fills"}, nil)

	shihanki_ask := 0.0
	shihanki_ask_size := 0.0
	perp_bid := 0.0
	perp_bid_size := 0.0
	client := client.NewSubClient("shit").Rest
	crossorder := strategy.NewCrossOrder(client, 15, "SHIT-0326", "SHIT-PERP")
	for {
		select {
		case v := <-ch:
			switch v.Type {
			case realtime.TICKER:
				if v.Symbol == "SHIT-0326" {
					if shihanki_ask == v.Ticker.Ask {
						continue
					}
					shihanki_ask = v.Ticker.Ask
					shihanki_ask_size = v.Ticker.AskSize
				}
				if v.Symbol == "SHIT-PERP" {
					if perp_bid == v.Ticker.Bid {
						continue
					}
					perp_bid = v.Ticker.Bid
					perp_bid_size = v.Ticker.BidSize
				}
				// fmt.Println(perp_bid-shihanki_ask, "bid", perp_bid, perp_bid_size, "ask", shihanki_ask, shihanki_ask_size)
				if perp_bid == 0.0 || shihanki_ask == 0.0 {
					continue
				}
				diff := perp_bid - shihanki_ask
				fmt.Printf("%.2f bid %.3f (%.3f) / ask %.3f (%.3f) %s\n", diff, perp_bid, perp_bid_size, shihanki_ask, shihanki_ask_size, time.Now().Format(time.UnixDate))

				if diff > 15 {
					var size float64
					notification.Notify(diff)
					if shihanki_ask_size < perp_bid_size {
						size = shihanki_ask_size
					} else {
						size = perp_bid_size
					}
					size = 0.001
					crossorder.UpdateTicker(diff, shihanki_ask, size)
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
