package inago

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/go-numb/go-ftx/rest"
)

const taker_fee float64 = 0.000679
const slippage float64 = 0.0005

// Inago Bot
type Bot struct {
	client       *rest.Client
	market       string
	recentTrades RecentTrades
	lot          float64
	state        int
	position     Position
	result       Result
	config       Config // parameter
	log          []string
}

func NewBot(market string, config Config) *Bot {
	return &Bot{
		// client:       client,
		market:       market,
		recentTrades: RecentTrades{},
		lot:          3,
		state:        0,
		config:       config,
	}
}

func (b *Bot) InitBot() {}

func (b *Bot) Result() {
	fmt.Println("期間", b.result.startTime, "〜", b.result.endTime)
	fmt.Println("トータルPnl:", b.result.totalPnl)
	fmt.Println("トータルFee:", b.result.totalFee)
	fmt.Println("トータルProfit:", b.result.totalPnl-b.result.totalFee)
	fmt.Println("ロング回数:", b.result.longCount)
	fmt.Println("ショート回数:", b.result.shortCount)
	fmt.Println("Win:", b.result.winCount)
	fmt.Println("Lose:", b.result.loseCount)
	winRate := float64(b.result.winCount) / (float64(b.result.longCount) + float64(b.result.shortCount))
	fmt.Println("勝率:", winRate)
	fmt.Println("")
	fmt.Println("----- Log ------")
	for _, l := range b.log {
		fmt.Println(l)
	}
}

type Trade struct {
	Time        time.Time
	Side        string
	Size        float64
	Price       float64
	Liquidation bool
}

type Position struct {
	Time  time.Time
	Side  string
	Size  float64
	Price float64
}

// TODO ryota-trade見る
type Result struct {
	totalPnl   float64
	totalFee   float64
	longCount  int
	shortCount int
	winCount   int
	loseCount  int
	startTime  time.Time
	endTime    time.Time
}

type RecentTrades []Trade

func (b *Bot) Handle(t, side, size, price, liquidation string) {
	parseTime, _ := time.Parse("2006-01-02 15:04:05.00000", t)
	parseSize, _ := strconv.ParseFloat(size, 64)
	parsePrice, _ := strconv.ParseFloat(price, 64)
	trade := Trade{
		Time:        parseTime,
		Side:        side,
		Size:        parseSize,
		Price:       parsePrice,
		Liquidation: (liquidation == "true"),
	}
	zero := time.Time{}
	if b.result.startTime == zero {
		b.result.startTime = trade.Time
	}
	b.result.endTime = trade.Time
	switch b.state {
	case 0:
		b.handleWaitForOpenPosition(trade)
	case 1:
		b.handleWaitForSettlement(trade)
	}
}

func (b *Bot) handleWaitForOpenPosition(trade Trade) {
	b.recentTrades = append(b.recentTrades, trade)

	var buyV, sellV float64
	for _, item := range b.recentTrades {
		// scope秒すぎたものは消していく
		if item.Time.Unix() <= (trade.Time.Unix() - b.config.scope) {
			b.recentTrades = b.recentTrades[1:]
			continue
		}
		if item.Side == "buy" {
			buyV += item.Size * item.Price
		} else {
			sellV += item.Size * item.Price
		}
	}

	if !b.isEntry(buyV, sellV) {
		return
	}

	if buyV > sellV {
		if b.config.reverse {
			b.entry("sell", buyV, trade)
		} else {
			b.entry("buy", buyV, trade)
		}
	} else {
		if b.config.reverse {
			b.entry("buy", sellV, trade)
		} else {
			b.entry("sell", buyV, trade)
		}
	}
}

func (b *Bot) handleWaitForSettlement(trade Trade) {

	// TODO Important logic
	if trade.Time.Unix() > b.position.Time.Unix()+b.config.settleTerm {
		b.settle(trade)
	}
}

func (b *Bot) isEntry(buyVolume, sellVolume float64) bool {
	return math.Max(buyVolume, sellVolume) > b.config.volumeTriger
}

func (b *Bot) entry(side string, v float64, trade Trade) {
	if b.state != 0 {
		return
	}
	if side == "buy" {
		trade.Price *= (1 + slippage)
		b.log = append(b.log, fmt.Sprintf("%s, volume: %.4f ロングエントリー Size: %.4f, Price: %.3f, Liquidation: %t",
			trade.Time,
			v,
			trade.Size,
			trade.Price,
			trade.Liquidation,
		))
		b.openPosition(side, trade)
		b.result.longCount++
	} else {
		trade.Price *= (1 - slippage)
		b.log = append(b.log, fmt.Sprintf("%s, volume: %.4f ショートエントリー Size: %.4f, Price: %.3f, Liquidation: %t",
			trade.Time,
			v,
			trade.Size,
			trade.Price,
			trade.Liquidation,
		))
		b.openPosition(side, trade)
		b.result.shortCount++
	}
}

func (b *Bot) settle(trade Trade) {
	if b.state != 1 {
		return
	}

	// market close order
	var pnl float64
	if b.position.Side == "buy" {
		pnl = (trade.Price - b.position.Price) * b.lot
	} else {
		pnl = (b.position.Price - trade.Price) * b.lot
	}
	// Fee
	fee := b.lot * taker_fee * 2
	b.log = append(b.log, fmt.Sprintf("%s, 決済しました  Size: %.4f, Price: %.3f, Pnl: %.4f",
		trade.Time,
		trade.Size,
		trade.Price,
		pnl-fee,
	))
	b.result.totalPnl += pnl
	if pnl-fee > 0 {
		b.result.winCount++
	} else {
		b.result.loseCount++
	}
	b.result.totalFee += fee
	b.state = 0
}

func (b *Bot) openPosition(side string, trade Trade) {
	// req := &orders.RequestForPlaceOrder{
	// 	Market: b.market,
	// 	Type:   "market",
	// 	Side:   side,
	// 	Size:   b.lot,
	// 	Ioc:    true}
	// o, err := b.client.PlaceOrder(req)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	b.position = Position{
		Time:  trade.Time,
		Side:  side,
		Size:  trade.Size,
		Price: trade.Price,
	}
	b.state = 1
}
