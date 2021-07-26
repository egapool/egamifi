package inago

import (
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/egapool/egamifi/internal"
	"github.com/egapool/egamifi/internal/indicators"
	"github.com/go-numb/go-ftx/rest"
)

const taker_fee float64 = 0.000679
const slippage float64 = 0.0005
const volatility_period int = 20
const entry_volatility_rate float64 = 1

// Inago Bot
type Bot struct {
	client        *rest.Client
	market        string
	recentTrades  RecentTrades
	lot           float64
	state         int // 1: open position, 2: cool down time
	position      Position
	lastCloseTime time.Time
	candles       []internal.Candle
	result        Result
	config        Config // parameter
	log           []string
	loc           *time.Location
	volatility    float64
}

func NewBot(market string, config Config) *Bot {
	jst, _ := time.LoadLocation("Asia/Tokyo")

	return &Bot{
		// client:       client,
		market:       market,
		recentTrades: RecentTrades{},
		lot:          0.25,
		state:        0,
		config:       config,
		loc:          jst,
	}
}

func (b *Bot) InitBot() {}

func (b *Bot) Result() {
	b.result.render()
	fmt.Println("")
	fmt.Println("----- Log ------")
	for _, l := range b.log {
		fmt.Println(l)
	}
}

func (b *Bot) ResultOneline() {
	if b.result.tradeCount() <= 10 {
		// return
	}
	// start, end, triger_volume, scope, settle_term, price_ratio, reverse, profit, pnl, fee, long_count, short_count, win, lose, total, entry, rate
	fmt.Printf("%s,%s,%.0f,%d,%d,%.5f,%.1f, %t,%.3f,%.3f,%.3f,%d,%d,%d,%d,%.3f,%.3f\n",
		b.result.startTime.Format("20060102150405"),
		b.result.endTime.Format("20060102150405"),
		b.config.volumeTriger,
		b.config.scope,
		b.config.settleTerm,
		b.config.settleRange,
		b.config.priceRatio,
		b.config.reverse,
		b.result.totalPnl-b.result.totalFee,
		b.result.totalPnl,
		b.result.totalFee,
		b.result.tradeCount(),
		b.result.winCount(),
		b.result.tradeCount()-b.result.winCount(),
		b.result.longCount+b.result.shortCount,
		b.result.winRate(),
		b.result.pf(),
	)

	// logging into file
	result_dir := fmt.Sprintf("result/inago/%s/%s-%s", b.market, b.result.startTime.Format("20060102150405"), b.result.endTime.Format("20060102150405"))
	if _, err := os.Stat(result_dir); os.IsNotExist(err) {
		os.MkdirAll(result_dir, 0777)
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

type RecentTrades []Trade

func (b *Bot) updateCandle(trade Trade) {
	if len(b.candles) == 0 {
		tp := internal.NewMinuteFromTime(trade.Time)
		candle := internal.NewCandle(tp)
		candle.AddTrade(trade.Size, trade.Price, trade.Side)
		b.candles = append(b.candles, *candle)
		return
	}
	latest_candle := b.candles[len(b.candles)-1]
	if latest_candle.Period.Contain(trade.Time) {
		latest_candle.AddTrade(trade.Size, trade.Price, trade.Side)
		b.candles[len(b.candles)-1] = latest_candle
	} else {
		tp := internal.NewMinuteFromTime(trade.Time)
		candle := internal.NewCandle(tp)
		candle.AddTrade(trade.Size, trade.Price, trade.Side)
		b.candles = append(b.candles, *candle)
		if len(b.candles) < volatility_period+1 {
			return
		}
		// ボラティリティを計算
		var mf indicators.Mfloat
		for _, c := range b.candles[len(b.candles)-(volatility_period+1) : len(b.candles)-1] {
			mf = append(mf, c.Close)
		}
		b.volatility = indicators.Std(mf)
		if len(b.candles) > 100 {
			b.candles = b.candles[1:]
		}
	}
}

func (b *Bot) Handle(t, side, price, size, liquidation string) {
	// jst, _ := time.LoadLocation("Asia/Tokyo")
	parseTime, _ := time.ParseInLocation("2006-01-02 15:04:05.00000", t, b.loc)
	parseSize, _ := strconv.ParseFloat(size, 64)
	parsePrice, _ := strconv.ParseFloat(price, 64)
	trade := Trade{
		Time:        parseTime,
		Side:        strings.TrimSpace(side),
		Size:        parseSize,
		Price:       parsePrice,
		Liquidation: (liquidation == "true"),
	}

	// 約定履歴からOHLC作成
	b.updateCandle(trade)

	if len(b.candles) < volatility_period+1 {
		return
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
	is_entry, entry_side, trigger_volume := b.isEntry(trade)
	if !is_entry {
		return
	}

	if b.config.reverse {
		b.entry(opSide(entry_side), trigger_volume, trade)
	} else {
		b.entry(entry_side, trigger_volume, trade)
	}
}

// TODO Important logic
func (b *Bot) handleWaitForSettlement(trade Trade) {

	// 建値より逆さやになったら損切り
	// 損切り入れると極端に勝率がさがる
	// if b.position.Side == "buy" && trade.Price < b.position.Price*(1-slippage*5) {
	// 	b.log = append(b.log, fmt.Sprintf("%s, 損切り(ロング) market: %.3f, open: %.3f",
	// 		trade.Time,
	// 		trade.Price,
	// 		b.position.Price,
	// 	))
	// 	b.settle(trade)
	//
	// 	// 損切りしたら一旦cooldown
	// 	b.state = 2
	// 	return
	// } else if b.position.Side == "sell" && trade.Price > b.position.Price*(1+slippage*5) {
	// 	b.log = append(b.log, fmt.Sprintf("%s, 損切り(ショート) market: %.3f, open: %.3f",
	// 		trade.Time,
	// 		trade.Price,
	// 		b.position.Price,
	// 	))
	// 	b.settle(trade)
	//
	// 	// 損切りしたら一旦cooldown
	// 	b.state = 2
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
func (b *Bot) isEntry(trade Trade) (is_entry bool, entry_side string, trigger_volume float64) {
	var buyV, sellV float64
	// scope秒すぎたものは消していく
	for _, item := range b.recentTrades {
		if item.Time.Unix() <= (trade.Time.Unix() - b.config.scope) {
			b.recentTrades = b.recentTrades[1:]
			continue
		}
	}

	// first_in_scope := b.recentTrades[0]
	// for _, item := range b.recentTrades {
	// 	// IDEA 荷重加算しても良いかも/直近ほど重い
	// 	// IDEA Done scope前と価格が開いていたらボーナスを付与する
	// 	if item.Side == "buy" {
	// 		price_diff := (item.Price / first_in_scope.Price) - 1 // 0.01 or -0.01
	// 		// TODO レート調整
	// 		rate := price_diff * b.config.priceRatio // 0.2 or -0.2
	// 		rate += 1                                // 1.2 or 0.8
	// 		buyV = buyV + item.Size*rate
	// 	} else {
	// 		price_diff := (first_in_scope.Price / item.Price) - 1 // 0.01 or -0.01
	// 		// TODO レート調整
	// 		rate := price_diff * b.config.priceRatio // 0.2 or -0.2
	// 		rate += 1                                // 1.2 or 0.8
	// 		sellV = sellV + item.Size*rate
	// 	}
	// }

	// 指数増加var
	previous_candle_close := b.candles[len(b.candles)-2].Close
	for _, item := range b.recentTrades {
		// IDEA 荷重加算しても良いかも/直近ほど重い
		// IDEA Done 前の分足の終値と価格が開いていたら指数荷重ボーナスを付与する
		if item.Side == "buy" {
			price_diff_rate := (item.Price / previous_candle_close) // ex. 0.01 or -0.01
			buyV = buyV + item.Size*math.Pow(price_diff_rate, b.config.priceRatio)
		} else {
			price_diff_rate := (previous_candle_close / item.Price) // ex. 0.01 or -0.01
			sellV = sellV + item.Size*math.Pow(price_diff_rate, b.config.priceRatio)
		}
	}

	if buyV == sellV {
		return false, "", 0
	}
	is_entry = math.Max(buyV, sellV) > b.config.volumeTriger
	if !is_entry {
		return false, "", 0
	}

	// 現行足の変動幅がボラティリティ以下ならエントリーしないfilter
	candle_body := b.candles[len(b.candles)-1].BodyLength()
	if math.Abs(candle_body) < b.volatility*entry_volatility_rate {
		// fmt.Println("出来高はあるが変動幅がVolatility以下なのでスルー", trade)
		return false, "", 0
	}

	if buyV > sellV {
		return is_entry, "buy", buyV
	} else {
		return is_entry, "sell", sellV
	}
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

	// TODO 決済注文から結果を抽出

	// Fee
	fee := trade.Price * b.lot * taker_fee * 2

	// market close order
	var pnl float64
	if b.position.Side == "buy" {
		pnl = (trade.Price - b.position.Price) * b.lot
		b.result.longProfit = append(b.result.longProfit, pnl-fee)
		b.result.longReturn = append(b.result.longReturn, 100*(pnl-fee)/b.position.Price)
		if pnl-fee > 0 {
			b.result.longWinning++
		}
	} else {
		pnl = (b.position.Price - trade.Price) * b.lot
		b.result.shortProfit = append(b.result.shortProfit, pnl-fee)
		b.result.shortReturn = append(b.result.shortReturn, 100*(pnl-fee)/b.position.Price)
		if pnl-fee > 0 {
			b.result.shortWinning++
		}
	}
	b.log = append(b.log, fmt.Sprintf("%s, 決済しました  Size: %.4f, Price: %.3f, OpenTime: %s, Pnl: %.4f",
		trade.Time,
		trade.Size,
		trade.Price,
		trade.Time.Sub(b.position.Time),
		pnl-fee,
	))
	b.result.totalPnl += pnl
	b.result.totalFee += fee

	// 最後に決済した時刻を保存
	b.lastCloseTime = trade.Time

	// IDEA 閾値どうする？
	// if pnl > 0 {
	// 	b.state = 2
	// } else {
	// 	b.state = 0
	// }
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
