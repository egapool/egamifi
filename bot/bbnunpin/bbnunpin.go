package bbnunpin

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/egapool/egamifi/internal/indicators"
	"github.com/egapool/egamifi/internal/notification"
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
	client    *rest.Client
	market    string
	mode      mode
	orders    map[string]order
	position  position
	size      float64
	limitSize float64
	nunpinCnt int32
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

func NewBbNunpin(client *rest.Client, market string, size float64) *BbNunpin {
	return &BbNunpin{
		client:    client,
		market:    market,
		mode:      Normal,
		size:      size,
		limitSize: 0.06,
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
		if b.mode != Normal {
			time.Sleep(time.Second * 30)
			continue
		}
		b.cancelAll()
		mf := b.fetchCandles()

		// place bollinger order
		_, upper, lower := indicators.BollingerBands(mf, 20, 2)
		upper_price := upper[len(upper)-1:][0]
		lower_price := lower[len(lower)-1:][0]
		b.PlaceOrder(b.market, "sell", upper_price, b.size, 1)
		b.PlaceOrder(b.market, "buy", lower_price, b.size, 1)
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
					notification.Notify("ポジション開始")
					Log("ポジションを持ちました", v.Fills.Side, "Size:", v.Fills.Size, "Price:", v.Fills.Price)
					Log("一旦注文中のオーダーを全て閉じます")
					b.cancelAll()
					b.mode = Positioning
					b.updatePosition()

					// 決済用
					b.placeSettleOrder(v.Fills.Price, v.Fills.Size)

					// ナンピン用
					b.placeNunpinOrder(v.Fills.Side, v.Fills.Price, v.Fills.Size)
				} else if b.mode == Positioning {

					// ナンピン
					if v.Fills.Side == b.position.side {
						notification.Notify("約定")
						Log("ナンピンしました", v.Fills.Side, "Size: ", v.Fills.Size, "Price: ", v.Fills.Price)
						fmt.Println("一旦注文中のオーダーを全て閉じます")
						b.cancelAll()
						b.updatePosition()
						Log("合計ポジション", b.position.side, "Size:", b.position.size, "AveragePrice:", b.position.avgPrice)

						// 決済用
						b.placeSettleOrder(b.position.avgPrice, b.position.size)

						// ナンピン用
						b.placeNunpinOrder(v.Fills.Side, b.position.avgPrice, b.position.size)

						// 決済
					} else {
						// 完全約定
						if v.Fills.Size == b.position.size {
							notification.Notify("完全約定")
							Log("決済しました", v.Fills.Side, "Size:", v.Fills.Size, "Price:", v.Fills.Price)
							Log("一旦注文中のオーダーを全て閉じます")
							b.cancelAll()
							b.resetPosition()
							b.mode = Normal
							Log("ボリンジャーバンドにオーダーを出し続けます...")

							// 部分約定
						} else {
							notification.Notify("部分約定")
							b.updatePosition()
							Log("決済オーダーに対して部分約定しました", v.Fills.Side, "TotalSize:", b.position.size, "部分約定Size", v.Fills.Size)
						}
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

func (b *BbNunpin) placeNunpinOrder(side string, base_price float64, base_size float64) {

	// ポジション上限を超えている場合は何もしない
	if b.position.size > b.limitSize {
		return
	}
	b.nunpinCnt++

	magnification := 0.5

	if side == "buy" {
		base_price = base_price - 150
	} else {
		base_price = base_price + 150
	}
	b.PlaceOrder(b.market, b.position.side, base_price, base_size*magnification, 1)
	Log("ナンピン用オーダー", b.position.side, "Size:", base_size*magnification, "Price:", base_price)
}

func (b *BbNunpin) placeSettleOrder(base_price float64, base_size float64) {

	order_side := b.oppositeSide()
	var close_price float64
	if order_side == "buy" {
		close_price = base_price - 100
	} else {
		close_price = base_price + 100
	}
	b.PlaceOrder(b.market, order_side, close_price, base_size, 1)
	Log("決済用オーダー", order_side, "Size:", base_size, "Price:", close_price)
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
		_, err := c.client.PlaceOrder(req)
		if err != nil {
			log.Fatal(err)
		}
		// fmt.Println(res)
		if side == "sell" {
			base_price = base_price * 1.002
		} else {
			base_price = base_price - (base_price * 0.002)
		}
	}
}

func (b *BbNunpin) updatePosition() {
	positions, err := b.client.Positions(&account.RequestForPositions{})
	if err != nil {
		log.Fatal(err)
	}
	for _, p := range *positions {
		fmt.Println(p)
		if p.Future == b.market {
			b.position = position{
				side:     p.Side,
				size:     p.Size,
				avgPrice: p.EntryPrice,
			}
		}
	}
}

func (b *BbNunpin) resetPosition() {
	b.position = position{}
}
