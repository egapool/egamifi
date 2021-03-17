package bbnunpin

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/egapool/egamifi/internal/indicators"
	"github.com/go-numb/go-ftx/realtime"
	"github.com/go-numb/go-ftx/rest"
	"github.com/go-numb/go-ftx/rest/private/account"
	"github.com/go-numb/go-ftx/rest/private/orders"
	"github.com/go-numb/go-ftx/rest/public/markets"
)

// Mode is Bot mode
type mode int

const (
	// bollinger bandにorderを起き続けるmode
	Normal mode = 1

	// positionを持っているmode
	Positioning mode = 2
)

type side string

const (
	BUY  side = "buy"
	SELL side = "sell"
)

type BbNunpin struct {
	client   *rest.Client
	market   string
	mode     mode
	orders   map[string]order
	position position
}

type order struct {
	ID   string
	side side
}

type position struct {
	ID       int
	side     string
	size     float64
	avgPrice float64
}

func NewBbNunpin(client *rest.Client, market string) *BbNunpin {
	return &BbNunpin{
		client: client,
		market: market,
		mode:   Normal,
	}
}

func (c *BbNunpin) Run() {
	go c.continueOrders()
	c.websocketRun()
}

func (b *BbNunpin) oppositeSide() string {
	if b.position.side == "buy" {
		return "sell"
	} else {
		return "buy"
	}
}

func (b *BbNunpin) continueOrders() {
	for {
		fmt.Println("ボリンジャーバンドにオーダー貼り直し")
		if b.mode != Normal {
			fmt.Println(" └ ポジション持っているのでやっぱりやめる")
			time.Sleep(time.Second * 30)
			continue
		}
		b.cancelAll()
		mf := b.fetchCandles()

		// place bollinger order
		_, upper, lower := indicators.BollingerBands(mf, 20, 2)
		upper_price := upper[len(upper)-1:][0]
		lower_price := lower[len(lower)-1:][0]
		b.PlaceOrder(b.market, "sell", upper_price, 0.005, 1)
		b.PlaceOrder(b.market, "buy", lower_price, 0.005, 1)
		time.Sleep(time.Second * 30)
	}
}

func (b *BbNunpin) cancelAll() {
	_, err := b.client.CancelAll(&orders.RequestForCancelAll{
		ProductCode: b.market})
	if err != nil {
		log.Fatal(err)
	}
}

func (b *BbNunpin) websocketRun() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan realtime.Response)
	go realtime.ConnectForPrivate(ctx, ch, os.Getenv("FTX_KEY"), os.Getenv("FTX_SECRET"), []string{"orders", "fills"}, nil, os.Getenv("FTX_SUBACCOUNT"))

	for {
		select {
		case v := <-ch:
			switch v.Type {
			case realtime.TICKER:

				fmt.Printf("%s	%+v\n", v.Symbol, v.Ticker)

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
				// fmt.Printf("%s	%+v\n", "ORDERS", v.Orders)

			case realtime.FILLS:
				fmt.Printf("%s	%+v\n", "FILLS", v.Fills)
				if b.mode == Normal {
					fmt.Println("ポジションを持ちました", "Size: ", v.Fills.Size, "Price: ", v.Fills.Price)
					fmt.Println("一旦注文中のオーダーを全て閉じます")
					b.cancelAll()
					b.mode = Positioning
					b.updatePosition()

					// positionに応じてorder投げる
					// 手動でやってみる

					// 決済用
					var close_price float64
					if v.Fills.Side == "buy" {
						close_price = v.Fills.Price + 50
					} else {
						close_price = v.Fills.Price - 50
					}
					b.PlaceOrder(b.market, b.oppositeSide(), close_price, v.Fills.Size, 1)

					// ナンピン用
					if v.Fills.Side == "buy" {
						close_price = v.Fills.Price - 200
					} else {
						close_price = v.Fills.Price + 200
					}
					b.PlaceOrder(b.market, b.position.side, close_price, v.Fills.Size*1.5, 1)
				} else if b.mode == Positioning {

					// ナンピン
					if v.Fills.Side == b.position.side {
						fmt.Println("ナンピンしました", "Size: ", v.Fills.Size, "Price: ", v.Fills.Price)
						fmt.Println("一旦注文中のオーダーを全て閉じます")
						b.cancelAll()
						b.updatePosition()

						// 決済用
						var close_price float64
						if v.Fills.Side == "buy" {
							close_price = v.Fills.Price + 50
						} else {
							close_price = v.Fills.Price - 50
						}
						b.PlaceOrder(b.market, b.oppositeSide(), close_price, v.Fills.Size, 1)

						// ナンピン用
						if v.Fills.Side == "buy" {
							close_price = v.Fills.Price - 200
						} else {
							close_price = v.Fills.Price + 200
						}
						b.PlaceOrder(b.market, b.position.side, close_price, v.Fills.Size*1.5, 1)

						// 決済
					} else {
						fmt.Println("決済しました", "Size: ", v.Fills.Size, "Price: ", v.Fills.Price)
						fmt.Println("一旦注文中のオーダーを全て閉じます")
						b.cancelAll()
						b.resetPosition()
						b.mode = Normal
					}
				}

			case realtime.UNDEFINED:
				fmt.Printf("UNDEFINED %s	%s\n", v.Symbol, v.Results.Error())
			}
		}
	}

}

func (c *BbNunpin) fetchCandles() indicators.Mfloat {
	var mf indicators.Mfloat

	req := &markets.RequestForCandles{
		ProductCode: c.market,
		Resolution:  60,
		Limit:       40,
	}
	candles, err := c.client.Candles(req)
	if err != nil {
		log.Fatal(err)
	}
	for _, c := range *candles {
		mf = append(mf, c.Close)
	}
	return mf
}

func (c *BbNunpin) PlaceOrder(market string, side string, price float64, lot float64, quantity int) {
	base_price := price
	for i := 0; i < quantity; i++ {
		req := &orders.RequestForPlaceOrder{
			Market: market,
			Type:   "limit",
			Side:   side,
			Price:  base_price,
			Size:   lot,
			Ioc:    false}
		res, err := c.client.PlaceOrder(req)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(res)
		if side == "sell" {
			base_price = base_price * 1.002
		} else {
			base_price = base_price - (base_price * 0.002)
		}
	}
}

func (b *BbNunpin) updatePosition() {
	positions, err := b.client.Positions(&account.RequestForPositions{})
	fmt.Println(positions)
	if err != nil {
		log.Fatal(err)
	}
	for _, p := range *positions {
		if p.Future == b.market {
			b.position = position{
				side:     p.Side,
				size:     p.Size,
				avgPrice: p.EntryPrice,
			}
			fmt.Println(b.position)
		}
	}
}

func (b *BbNunpin) resetPosition() {
	b.position = position{}
}
