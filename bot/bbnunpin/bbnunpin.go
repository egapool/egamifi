package bbnunpin

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
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
const division_num = 4

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
	nunpinCnt   int32
	// 足の幅
	resolution int
	// bbの基準値
	middlePrice float64
	upperPrice  float64
	lowerPrice  float64
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
		resolution:  60,
		orders:      Orders{},
		notifer:     n,
	}
}

func (b *BbNunpin) Run() {
	b.cleanUp()
	go b.continueOrders()
	b.websocketRun()
}

// マーケットのオーダーをリセット
func (b *BbNunpin) cleanUp() {
	orders, err := b.client.OpenOrder(&orders.RequestForOpenOrder{
		ProductCode: b.market,
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, order := range *orders {
		b.cancelOrder(order.ID)
	}
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
		b.upperPrice = upper_price
		b.lowerPrice = lower_price
		// update volatility
		b.volatility = (upper_price - middle_price) / 2

		if b.mode != Normal {
			if b.mode == Positioning {
				b.adjustSettleOrder()
			}
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
			b.PlaceOrder(b.market, "sell", upper_price3+(b.volatility*0.5), b.initialSize*3, NunpinOrder)
		} else {
			b.PlaceOrder(b.market, "buy", lower_price, b.initialSize, InitOrder)
			time.Sleep(time.Microsecond * 50)
			b.PlaceOrder(b.market, "buy", lower_price3-(b.volatility*0.5), b.initialSize*3, NunpinOrder)
		}

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
		b.handleSettle(orderID, side, size, price)
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
	b.placeSettleOrder(price, size, 4)

	// 全決済したらordersから消す
	delete(b.orders, orderID)
}

// ナンピン注文約定
func (b *BbNunpin) handleNunpin(orderID int, side string, size float64, price float64) {
	b.notifer.Notify("約定")
	Log("約定", "ナンピンしました", side, "Size: ", size, "Price: ", price)
	fmt.Println("一旦注文中のオーダーを全て閉じます")
	b.cancelAll()
	b.updatePosition(side, size, price)
	Log("合計ポジション", b.position.side, "Size:", b.position.size, "AveragePrice:", b.position.avgPrice)

	// 決済用
	b.placeSettleOrder(b.position.avgPrice, b.position.size, division_num)
	// 損切り用
	b.PlaceStopLossOrder(b.volatility * 1.5)

	// 全決済したらordersから消す
	delete(b.orders, orderID)
}

// 決済注文約定
func (b *BbNunpin) handleSettle(orderID int, side string, size float64, price float64) {
	b.notifer.Notify("一部利確しました")
	Log("約定", "決済しました", side, "Size: ", size, "Price: ", price)
	b.updatePosition(side, size, price)

	// TODO 一旦トレイリングストップは保留
	// 損切り用
	// r :=
	// 	b.PlaceStopLossOrder(b.volatility)

	// TODO どこまで決済したか、あるいは価格が移動したかでナンピン注文消すかどうか決める

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

// 残っている利確オーダーが、ミドルラインより上にきた場合に、
// ポジション価格とミドルラインの間にくるように動かす
func (b *BbNunpin) adjustSettleOrder() {
	settle_orders := b.orders.SettleOrder()
	division_num := len(settle_orders)
	for _, order := range settle_orders {
		b.cancelOrder(order.ID)
	}
	b.placeSettleOrder(b.position.avgPrice, b.position.size, float64(division_num))

}

func (b *BbNunpin) cancelAll() {
	for _, order := range b.orders {
		b.cancelOrder(order.ID)
	}
}

// 片方のオーダーをキャンセル
func (b *BbNunpin) cancelOneSide(side string) {
	for _, order := range b.orders.OneSide(side) {
		b.cancelOrder(order.ID)
	}
}

func (b *BbNunpin) cancelOrder(orderID int) {
	_, err := b.client.CancelByID(&orders.RequestForCancelByID{OrderID: orderID})
	delete(b.orders, orderID)
	if err != nil {
		// log.Fatal(err)
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

func (b *BbNunpin) placeSettleOrder(price float64, base_size float64, division_num float64) {

	order_side := b.position.oppositeSide()
	var price_range float64
	if order_side == "sell" {
		price_range = (b.upperPrice - price) / division_num // BB上線までの分割幅
	} else {
		price_range = (b.lowerPrice - price) / division_num // BB下線までの分割幅
	}
	division_size := base_size / division_num

	// TODO 端数が出ないようにポジションサイズを調整する
	for i := 0; i < int(division_num); i++ {
		price = price + price_range
		b.PlaceOrder(b.market, order_side, price, division_size, SettleOrder)
		time.Sleep(time.Millisecond * 50)
	}
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
	id, err := strconv.Atoi(o.OrderID)
	if err != nil {
		// nothing
	}
	b.orders[id] = Order{
		ID:      id,
		side:    o.Side,
		size:    o.Size,
		purpose: StopLossOrder,
	}
	Log("損切りオーダーをだしました", b.orders[id])
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
			side:      side,
			size:      size,
			avgPrice:  price,
			initPrice: price,
			settleCnt: 0,
		}
		Log("ポジションを更新しました", b.position)
		return
	}
	if b.position.side != side {
		b.position.size = b.position.size - size
		// 利確オーダーを何回食ったか
		b.position.settleCnt = b.position.settleCnt + 1
	} else {
		b.position.avgPrice = (b.position.size*b.position.avgPrice + size*price) / (b.position.size + size)
		b.position.size = b.position.size + size
	}
	Log("ポジションを更新しました", b.position)
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
