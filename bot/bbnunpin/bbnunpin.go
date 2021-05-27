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

	// 損切り後逆方向トレンド継続中
	Trending mode = 3
)

type side string

const (
	BUY  side = "buy"
	SELL side = "sell"
)

type notifer interface {
	Notify(message string)
}

type BbNunpin struct {
	client   *rest.Client
	market   string
	mode     mode
	orders   Orders
	position position
	// 初回エントリー縦玉数
	initialSize float64
	limitSize   float64
	nunpinCnt   int32
	// 足の幅
	resolution int
	// bbの基準値
	middlePrice float64
	volatility  float64
	notifer     notifer
}

func NewBbNunpin(client *rest.Client, market string, size float64) *BbNunpin {
	var n notifer
	if os.Getenv("SLACK_CHANNEL") == "" {
		n = notification.NewNone()
	} else {
		n = notification.NewNotifer(os.Getenv("SLACK_CHANNEL"), os.Getenv("SLACK_WEBHOOK"))
	}
	return &BbNunpin{
		client:      client,
		market:      market,
		mode:        Normal,
		initialSize: size,
		limitSize:   0.04,
		resolution:  60,
		orders:      Orders{},
		notifer:     n,
	}
}

func (b *BbNunpin) Run() {
	go b.continueOrders()
	b.websocketRun()
}

func (b *BbNunpin) continueOrders() {
	for {
		// filter := b.emaFilter()
		mf := b.fetchCandles(b.resolution)
		tick := mf[len(mf)-1:][0]

		middle, upper, lower := indicators.BollingerBands(mf, 20, 2)
		middle_price := middle[len(middle)-1:][0]
		upper_price := upper[len(upper)-1:][0]
		lower_price := lower[len(lower)-1:][0]
		b.middlePrice = middle_price
		// update volatility
		b.volatility = (upper_price - middle_price) / 2

		if b.mode != Normal {
			time.Sleep(time.Second * (time.Duration(b.resolution) / 3))
			continue
		}

		_, upper3, lower3 := indicators.BollingerBands(mf, 20, 3)
		upper_price3 := upper3[len(upper3)-1:][0]
		lower_price3 := lower3[len(lower3)-1:][0]

		b.cancelAll()
		// place bollinger order
		if tick > middle_price {
			b.PlaceOrder(b.market, "sell", upper_price, b.initialSize, InitOrder)
			time.Sleep(time.Microsecond * 50)
			b.PlaceOrder(b.market, "sell", upper_price3, b.initialSize*3, NunpinOrder)
		} else {
			b.PlaceOrder(b.market, "buy", lower_price, b.initialSize, InitOrder)
			time.Sleep(time.Microsecond * 50)
			b.PlaceOrder(b.market, "buy", lower_price3, b.initialSize*3, NunpinOrder)
		}

		// TODO 約定したらplace stop loss order
		//bb3の外側volatility1つ分にstop loss orderいれる
		time.Sleep(time.Second * (time.Duration(b.resolution) / 3))
	}
}

func (b *BbNunpin) handler(orderID int, side string, size float64, price float64) {
	switch b.orders[orderID].purpose {
	case InitOrder:
		b.handleStartPosition(orderID, side, size, price)
	case NunpinOrder:
		b.handleNunpin(orderID, side, size, price)
	case SettleOrder:
		// some
	case StopLossOrder:
		b.handleStopLoss(orderID, side, size, price)
	}
}

// ポジションスタート
func (b *BbNunpin) handleStartPosition(orderID int, side string, size float64, price float64) {
	b.notifer.Notify("ポジション開始")
	Log("約定", "ポジションを持ちました", side, "Size:", size, "Price:", price)
	Log("反対側のオーダーを閉じます")
	b.mode = Positioning
	b.updatePosition(side, size, price)
	b.cancelOneSide(b.position.oppositeSide())

	// 決済用
	b.placeSettleOrder(price, size)

	// 損切り用
	// TODO どうせ下にナンピンオーダーがいるから必要ない？
	// b.PlaceStopLossOrder(b.volatility * 2)

	// 全決済したらordersから消す
	delete(b.orders, orderID)
}

// ナンピン注文約定
// TODO ナンピンが部分約定の時の対応が必要
func (b *BbNunpin) handleNunpin(orderID int, side string, size float64, price float64) {
	b.notifer.Notify("約定")
	Log("約定", "ナンピンしました", side, "Size: ", size, "Price: ", price)
	fmt.Println("一旦注文中のオーダーを全て閉じます")
	b.cancelAll()
	b.updatePosition(side, size, price)
	Log("合計ポジション", b.position.side, "Size:", b.position.size, "AveragePrice:", b.position.avgPrice)

	// 決済用
	// TODO ナンピン後の決済注文どうやってさすか
	b.placeSettleOrder(b.position.avgPrice, b.position.size)
	// 損切り用
	b.PlaceStopLossOrder(b.volatility)

	// 全決済したらordersから消す
	delete(b.orders, orderID)
}

// 決済注文約定
func (b *BbNunpin) handleSettle(orderID int, side string, size float64, price float64) {
	b.notifer.Notify("一部利確しました")
	Log("約定", "決済しました", side, "Size: ", size, "Price: ", price)
	b.updatePosition(side, size, price)

	// TODO 一旦決済後の損切りラインアップデートは保留

	b.updateStopLossOrder()

	// 全決済したらordersから消す
	delete(b.orders, orderID)

	if b.position.size == 0 {
		b.notifer.Notify("ポジションを閉じました")
		Log("約定", "ポジション閉じる", side, "Size: ", size, "Price: ", price)
		b.mode = Normal
	}
}

// 損切り注文約定
func (b *BbNunpin) handleStopLoss(orderID int, side string, size float64, price float64) {
	b.notifer.Notify("損切りしました")
	Log("約定", "損切りしました", side, "Size: ", size, "Price: ", price)
	fmt.Println("一旦注文中のオーダーを全て閉じます")
	b.cancelAll()
	b.resetPosition()
	b.mode = Trending

	// 全決済したらordersから消す
	delete(b.orders, orderID)
}

func (b *BbNunpin) cancelAll() {
	for _, order := range b.orders {
		_, err := b.client.CancelByID(&orders.RequestForCancelByID{OrderID: order.ID})
		delete(b.orders, order.ID)
		if err != nil {
			// log.Fatal(err)
		}
	}
}

// 片方のオーダーをキャンセル
func (b *BbNunpin) cancelOneSide(side string) {
	for _, order := range b.orders.OneSide(side) {
		_, err := b.client.CancelByID(&orders.RequestForCancelByID{OrderID: order.ID})
		delete(b.orders, order.ID)
		if err != nil {
			// log.Fatal(err)
		}
	}
}

func (b *BbNunpin) websocketRun() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan realtime.Response)
	go realtime.ConnectForPrivate(ctx, ch, os.Getenv("FTX_KEY"), os.Getenv("FTX_SECRET"), []string{"fills"}, nil, os.Getenv("FTX_SUBACCOUNT"))

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
				b.handler(v.Fills.OrderID, v.Fills.Side, v.Fills.Size, v.Fills.Price)

			case realtime.UNDEFINED:
				fmt.Printf("UNDEFINED %s	%s\n", v.Symbol, v.Results.Error())
			}
		}
	}

}

func (c *BbNunpin) fetchCandles(resolution int) indicators.Mfloat {
	var mf indicators.Mfloat

	req := &markets.RequestForCandles{
		ProductCode: c.market,
		Resolution:  int(time.Duration(resolution)),
		Limit:       80,
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

func (b *BbNunpin) placeSettleOrder(price float64, base_size float64) {
	var division_num float64 = 4

	order_side := b.position.oppositeSide()
	price_range := (b.middlePrice - price) / division_num // 中央線までの4分割幅
	division_size := base_size / division_num
	for i := 0; i < int(division_num); i++ {
		price = price + price_range
		b.PlaceOrder(b.market, order_side, price, division_size, SettleOrder)
	}
	// Log("注文", "決済用オーダー", order_side, "Size:", base_size, "Price:", close_price)
}

func (c *BbNunpin) PlaceOrder(market string, side string, price float64, lot float64, purpose purpose) {
	req := &orders.RequestForPlaceOrder{
		Market: market,
		Type:   "limit",
		Side:   side,
		Price:  price,
		Size:   lot,
		Ioc:    false}
	o, err := c.client.PlaceOrder(req)
	if err != nil {
		log.Fatal(err)
	}
	c.orders[o.ID] = Order{
		ID:      o.ID,
		side:    o.Side,
		size:    o.Size,
		purpose: purpose,
	}
}

func (b *BbNunpin) PlaceStopLossOrder(price_range float64) {
	req := &orders.RequestForPlaceTriggerOrder{
		Market:       b.market,
		Type:         "stop",
		Side:         b.position.oppositeSide(),
		TriggerPrice: b.position.stopLossPrice(price_range),
		Size:         b.position.size}
	o, err := b.client.PlaceTriggerOrder(req)
	if err != nil {
		log.Fatal(err)
	}
	b.orders[o.ID] = Order{
		ID:      o.ID,
		side:    o.Side,
		size:    o.Size,
		purpose: StopLossOrder,
	}
}

func (b *BbNunpin) updateStopLossOrder() {
	order, err := b.orders.StopLossOrder()
	if err == nil {
		b.client.ModifyTriggerOrder(&orders.RequestForModifyTriggerOrder{
			OrderID: fmt.Sprintf("%d", order.ID),
			Size:    b.position.size,
		})
		Log("損切りオーダーをリサイズしました", b.position.size)
	}
}

func (b *BbNunpin) updatePosition(side string, size float64, price float64) {
	if b.position.size == 0 {
		b.position = position{
			side:     side,
			size:     size,
			avgPrice: price,
		}
		return
	}
	if b.position.side != side {
		b.position.size = b.position.size - size
	} else {
		b.position.avgPrice = (b.position.size*b.position.avgPrice + size*price) / (b.position.size + size)
		b.position.size = b.position.size + size
	}
}

func (b *BbNunpin) resetPosition() {
	b.position = position{}
}

// 15分足EMAの直近の傾斜を返す
func (b *BbNunpin) emaFilter() string {
	mf := b.fetchCandles(900)
	ema := mf.EMA(20)
	last := ema[len(ema)-2:]
	if last[1]/last[0] >= 1 {
		return "buy"
	} else {
		return "sell"
	}
}
