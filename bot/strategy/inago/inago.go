package inago

import (
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/go-numb/go-ftx/rest"
)

const taker_fee float64 = 0.000679
const slippage float64 = 0.0005

// Inago Bot
type Bot struct {
	client        *rest.Client
	market        string
	recentTrades  RecentTrades
	lot           float64
	state         int // 1: open position, 2: cool down time
	position      Position
	lastCloseTime time.Time
	result        Result
	config        Config // parameter
	log           []string
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
	fmt.Println("勝率:", b.result.winRate())

	fmt.Println("")
	fmt.Println("----- Log ------")
	for _, l := range b.log {
		fmt.Println(l)
	}
}

func (b *Bot) ResultOneline() {
	// start, end, triger_volume, scope, settle_term, reverse, profit, pnl, fee, long_count, short_count, win, lose, total, entry, rate
	fmt.Printf("%s,%s,%.0f,%d,%d,%.5f,%t,%.3f,%.3f,%.3f,%d,%d,%d,%d,%d,%.3f\n",
		b.result.startTime.Format("20060102150405"),
		b.result.endTime.Format("20060102150405"),
		b.config.volumeTriger,
		b.config.scope,
		b.config.settleTerm,
		b.config.settleRange,
		b.config.reverse,
		b.result.totalPnl-b.result.totalFee,
		b.result.totalPnl,
		b.result.totalFee,
		b.result.longCount,
		b.result.shortCount,
		b.result.winCount,
		b.result.loseCount,
		b.result.longCount+b.result.shortCount,
		b.result.winRate(),
	)

	// logging into file
	result_dir := fmt.Sprintf("result/inago/%s-%s", b.result.startTime.Format("20060102150405"), b.result.endTime.Format("20060102150405"))
	if _, err := os.Stat(result_dir); os.IsNotExist(err) {
		os.Mkdir(result_dir, 0777)
	}

	filepath := fmt.Sprintf(result_dir+"/%s.log", b.config.Serialize())
	if err := os.Remove(filepath); err != nil {
	}
	file, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	writer := csv.NewWriter(file)

	for _, l := range b.log {
		writer.Write([]string{l})
	}
	writer.Flush()
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

func (r *Result) winRate() float64 {
	return float64(r.winCount) / (float64(r.longCount) + float64(r.shortCount))
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
	// neutral
	case 0:
		b.handleWaitForOpenPosition(trade)
		// open position
	case 1:
		b.handleWaitForSettlement(trade)
		// waiting for cool down time
	case 2:
		b.handleCoolDownTime(trade)
	}
}

func (b *Bot) handleWaitForOpenPosition(trade Trade) {
	b.recentTrades = append(b.recentTrades, trade)

	var buyV, sellV float64
	// scope秒すぎたものは消していく
	for _, item := range b.recentTrades {
		if item.Time.Unix() <= (trade.Time.Unix() - b.config.scope) {
			b.recentTrades = b.recentTrades[1:]
			continue
		}
	}
	first_in_scope := b.recentTrades[0]
	for _, item := range b.recentTrades {
		// IDEA 荷重加算しても良いかも/直近ほど重い
		// IDEA Done scope前と価格が開いていたらボーナスを付与する
		var r float64 = 5
		if item.Side == "buy" {
			price_diff := (item.Price / first_in_scope.Price) - 1 // 0.01 or -0.01
			// TODO レート調整
			rate := price_diff * r // 0.2 or -0.2
			rate += 1              // 1.2 or 0.8
			buyV += item.Size * item.Price * rate
		} else {
			price_diff := (first_in_scope.Price / item.Price) - 1 // 0.01 or -0.01
			// TODO レート調整
			rate := price_diff * r // 0.2 or -0.2
			rate += 1              // 1.2 or 0.8
			sellV += item.Size * item.Price * rate
		}
	}

	if !b.isEntry(buyV, sellV) {
		return
	}

	// fmt.Println(buyV, sellV)
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

// TODO Important logic
func (b *Bot) handleWaitForSettlement(trade Trade) {

	// TODO 損切り?

	// positionの方向に進んでいる以上は決済しない
	// if trade.Side == b.position.Side {
	// 	return
	// }

	// X%さやが開いたらclose
	// TODO トレンドフォロー
	var price_range float64
	if b.position.Side == "buy" {
		price_range = (trade.Price - b.position.Price) / b.position.Price
	} else {
		price_range = (b.position.Price - trade.Price) / b.position.Price
	}
	if price_range > b.config.settleRange {
		b.settle(trade)
		return
	}

	// 制限時間過ぎたら強制close
	if trade.Time.Unix() > b.position.Time.Unix()+b.config.settleTerm {
		b.settle(trade)
		return
	}
}

func (b *Bot) handleCoolDownTime(trade Trade) {
	if trade.Time.Unix() < b.lastCloseTime.Unix()+b.config.scope {
		return
	}
	// cool down time finish
	b.state = 0
	return
}

// IDEA フィルタリング等いれて改良する
func (b *Bot) isEntry(buyVolume, sellVolume float64) bool {
	return math.Max(buyVolume, sellVolume) > b.config.volumeTriger
}

func (b *Bot) entry(side string, v float64, trade Trade) {
	if b.state != 0 {
		return
	}
	if side == "buy" {
		trade.Price *= (1 + slippage)
		b.log = append(b.log, fmt.Sprintf("%s, volume: %.4f ロングエントリー Size: %.4f, Price: %.3f",
			trade.Time,
			v,
			trade.Size,
			trade.Price,
		))
		b.openPosition(side, trade)
		b.result.longCount++
	} else {
		trade.Price *= (1 - slippage)
		b.log = append(b.log, fmt.Sprintf("%s, volume: %.4f ショートエントリー Size: %.4f, Price: %.3f",
			trade.Time,
			v,
			trade.Size,
			trade.Price,
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
	b.log = append(b.log, fmt.Sprintf("%s, 決済しました  Size: %.4f, Price: %.3f, OpenTime: %s, Pnl: %.4f",
		trade.Time,
		trade.Size,
		trade.Price,
		trade.Time.Sub(b.position.Time),
		pnl-fee,
	))
	b.result.totalPnl += pnl
	if pnl-fee > 0 {
		b.result.winCount++
	} else {
		b.result.loseCount++
	}
	b.result.totalFee += fee

	// 最後に決済した時刻を保存
	b.lastCloseTime = trade.Time

	// IDEA 閾値どうする？
	if pnl > 0 {
		b.state = 2
	} else {
		b.state = 0
	}
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
