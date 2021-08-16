package inago

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/egapool/egamifi/internal"
	"github.com/egapool/egamifi/internal/indicators"
)

const taker_fee float64 = 0.000679
const maker_fee float64 = 0.000194
const slippage float64 = 0.0007

const entry_volatility_rate float64 = 1

type BotLogger interface {
	Log(l string)
	GetLogs() []string
	Output()
	Logfile() string
}

type InagoClient interface {
	MarketOrder(market string, side string, size float64, time time.Time, price float64) Position
	Close(market string, position Position, price float64) float64
}

// Inago Bot
type Bot struct {
	client        InagoClient
	logger        BotLogger
	market        string
	recentTrades  RecentTrades
	state         int // 1: open position, 2: cool down time
	position      Position
	lastCloseTime time.Time
	candles       []internal.Candle
	result        Result
	config        Config // parameter
	loc           *time.Location
	nunpin        int
	paidFee       float64 // OpenからCloseまでの累積Fee
	firstOfCandle bool    // 足で一回だけ判定

	// markete data
	volatility  float64
	middlePrice float64
	upperPrice  float64
	lowerPrice  float64
}

func NewBot(client InagoClient, market string, config Config, logger BotLogger) *Bot {
	jst, _ := time.LoadLocation("Asia/Tokyo")

	return &Bot{
		client:        client,
		logger:        logger,
		market:        market,
		config:        config,
		loc:           jst,
		firstOfCandle: true,
	}
}

func (b *Bot) Market() string {
	return b.market
}

func (b *Bot) InitBot() {}

func (b *Bot) Result() {
	b.result.render()
	fmt.Println("")
	fmt.Println("----- Log ------")
	for _, l := range b.logger.GetLogs() {
		fmt.Println(l)
	}
}

func (b *Bot) ResultOneline() {
	if b.result.tradeCount() <= 10 {
		// return
	}
	// start, end, avg_volume_period, against_avg_volume_rate, minimum_rate, profit, pnl, fee, long_count, short_count, win, lose, total, entry, rate
	fmt.Printf("%s,%s,%d,%.1f,%.0f,%.3f,%.1f,%.3f,%.3f,%.3f,%d,%d,%d,%d,%.3f,%.3f,%s\n",
		b.result.startTime.Format("20060102150405"),
		b.result.endTime.Format("20060102150405"),
		b.config.avgVolumePeriod,
		b.config.againstAvgVolumeRate,
		b.config.minimumVolume,
		b.config.guardOverBb3,
		b.config.minimumCandleLength,
		b.result.totalPnl-b.result.totalFee,
		b.result.totalPnl,
		b.result.totalFee,
		b.result.tradeCount(),
		b.result.winCount(),
		b.result.tradeCount()-b.result.winCount(),
		b.result.longCount+b.result.shortCount,
		b.result.winRate(),
		b.result.pf(),
		b.logger.Logfile(),
	)
	b.logger.Output()
}

type Trade struct {
	Time        time.Time
	Side        string
	Size        float64
	Price       float64
	Liquidation bool
}

type Position struct {
	Time    time.Time
	Side    string
	Size    float64
	Price   float64
	Reverse bool
}

type RecentTrades []Trade

func (b *Bot) getCandle(offset int) internal.Candle {
	return b.candles[len(b.candles)-1-offset]
}

func (b *Bot) avgVolume(side string, period, offset int) (avg_volume float64) {
	rate := 0.05
	i := 0
	for _, c := range b.candles[len(b.candles)-period-offset-1 : len(b.candles)-offset-1] {
		if side == "buy" {
			avg_volume += c.BuyVolume * (1 - (rate * float64(period-i)))
		} else {
			avg_volume += c.SellVolume * (1 - (rate * float64(period-i)))
		}
		i++
	}
	return avg_volume / float64(period)
}

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
		// fmt.Println("Candle created: ", b.getCandle(0))
		b.firstOfCandle = true
		// tp := internal.NewMinuteFromTime(trade.Time)
		tp := internal.NewMinutesFromTime(trade.Time, 5)
		candle := internal.NewCandle(tp)
		candle.AddTrade(trade.Size, trade.Price, trade.Side)
		b.candles = append(b.candles, *candle)
		if len(b.candles) < b.config.avgVolumePeriod+1 {
			return
		}
		// ボラティリティを計算
		var mf indicators.Mfloat
		for _, c := range b.candles[len(b.candles)-(b.config.avgVolumePeriod+1) : len(b.candles)-1] {
			mf = append(mf, c.Close)
		}
		b.volatility = indicators.Std(mf)
		middle, upper, lower := indicators.BollingerBands(mf, b.config.avgVolumePeriod, 3)
		middle_price := middle[len(middle)-1:][0]
		upper_price := upper[len(upper)-1:][0]
		lower_price := lower[len(lower)-1:][0]
		b.middlePrice = middle_price
		b.upperPrice = upper_price
		b.lowerPrice = lower_price
		if len(b.candles) > 100 {
			b.candles = b.candles[1:]
		}
	}
}

// Backtestデータのハンドラー
func (b *Bot) HandleBacktest(t, side, price, size, liquidation string) {
	parseTime, _ := time.ParseInLocation("2006-01-02 15:04:05.00000", t, b.loc)
	parsePrice, _ := strconv.ParseFloat(price, 64)
	parseSize, _ := strconv.ParseFloat(size, 64)
	trade := Trade{
		Time:        parseTime,
		Side:        strings.TrimSpace(side),
		Price:       parsePrice,
		Size:        parseSize,
		Liquidation: (liquidation == "true"),
	}
	b.process(trade)
}

// 日運用時に発生するデータのハンドラー
func (b *Bot) Handle(t time.Time, side string, price, size float64, liquidation bool) {
	trade := Trade{
		Time:        t.In(b.loc),
		Side:        side,
		Price:       price,
		Size:        size,
		Liquidation: liquidation,
	}
	b.process(trade)
}

func (b *Bot) process(trade Trade) {
	// 約定履歴からOHLC作成
	b.updateCandle(trade)

	if len(b.candles) < b.config.avgVolumePeriod+3 {
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
		if b.firstOfCandle {
			b.handleWaitForOpenPosition(trade)
		}
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
	is_entry, entry_side, trigger_volume, reverse := b.isEntry4(trade)
	if !is_entry {
		return
	}

	b.entry(entry_side, b.config.lot, trigger_volume, trade, reverse)
}

// TODO Important logic
func (b *Bot) handleWaitForSettlement(trade Trade) {

	// 建値より1Volatility分逆さやになったらナンピン
	var max_nunpin_cnt int = 2
	var nunpin_rate float64 = 2
	var nunpin_offtime int64 = 5
	if b.position.Reverse && b.nunpin < max_nunpin_cnt && trade.Time.Unix() > b.position.Time.Unix()+nunpin_offtime {
		if b.position.Side == "buy" {
			if trade.Price < b.position.Price-b.volatility*nunpin_rate*(1+float64(b.nunpin)) {
				b.state = 0
				b.logger.Log(fmt.Sprintf("%s ナンピンしました, ナンピン価格: %.5f, 建値: %.5f",
					trade.Time,
					trade.Price,
					b.position.Price,
				))
				b.entry(b.position.Side, b.config.lot*(1+float64(b.nunpin)), trade.Size, trade, true)
				b.nunpin++
				return
			}
		} else {
			if trade.Price > b.position.Price+b.volatility*nunpin_rate*(1+float64(b.nunpin)) {
				b.state = 0
				b.logger.Log(fmt.Sprintf("%s ナンピンしました, ナンピン価格: %.5f, 建値: %.5f",
					trade.Time,
					trade.Price,
					b.position.Price,
				))
				b.entry(b.position.Side, b.config.lot*(1+float64(b.nunpin)), trade.Size, trade, true)
				b.nunpin++
				return
			}
		}
	}

	// StopLoss
	// if b.nunpin >= max_nunpin_cnt {
	// 	loss_cut_range := b.volatility * 1
	// 	if b.position.Side == "buy" {
	// 		if trade.Price < b.position.Price-loss_cut_range {
	// 			b.logger.Log(fmt.Sprintf("価格が建値 (%.5f) より1Volatility下 (%.5f) を下回ったので損切りします", b.position.Price, b.position.Price-loss_cut_range))
	// 			b.settle(trade, true)
	// 			return
	// 		}
	// 	} else {
	// 		if trade.Price > b.position.Price+loss_cut_range {
	// 			b.logger.Log(fmt.Sprintf("価格が建値 (%.5f) より1Volatility上 (%.5f) を上回ったので損切りします", b.position.Price, b.position.Price+loss_cut_range))
	// 			b.settle(trade, true)
	// 			return
	// 		}
	// 	}
	// }

	// X%さやが開いたらclose
	var price_range float64
	if b.position.Side == "buy" {
		price_range = (trade.Price - b.position.Price) / b.position.Price
	} else {
		price_range = (b.position.Price - trade.Price) / b.position.Price
	}
	// var ProfitabilityRange float64 = 0.03
	var ProfitabilityRange float64 = 0.005
	if price_range > ProfitabilityRange {
		b.logger.Log("利確幅到達につき close")
		b.settle(trade, false)
		return
	}

	var maxKeppTime int64 = 60 * 3
	// 制限時間過ぎたら強制close
	if trade.Time.Unix() > b.position.Time.Unix()+maxKeppTime {
		b.settle(trade, true)
		return
	}
}

func (b *Bot) handleCoolDownTime(trade Trade) {
	if trade.Time.Unix() < b.lastCloseTime.Unix()+60 {
		return
	}
	// cool down time finish
	b.state = 0
	return
}
func (b *Bot) isEntry4(trade Trade) (is_entry bool, entry_side string, trigger_volume float64, reverse bool) {
	b.firstOfCandle = false
	candle := b.getCandle(1)
	totalBV := b.avgVolume("buy", b.config.avgVolumePeriod, 1) * float64(b.config.avgVolumePeriod)
	totalSV := b.avgVolume("sell", b.config.avgVolumePeriod, 1) * float64(b.config.avgVolumePeriod)
	var which string
	// diff_rate := math.Abs(totalBV-totalSV) * 2 / (totalBV + totalSV)
	diff := math.Abs(totalBV - totalSV)
	if totalBV > totalSV {
		which = "買い優勢"
		i := 0
		for i < int(diff/totalBV*10) {
			i++
			which = which + "🟩"
		}
	} else {
		which = "売り優勢"
		i := 0
		for i < int(diff/totalSV*10) {
			i++
			which = which + "🟥"
		}
	}
	var dir string
	if candle.Open < candle.Close {
		dir = "🟩"
	} else {
		dir = "🟥"
	}
	b.logger.Log(fmt.Sprintf("%s, %s, %.3f, %.3f, %.2f, %.0f, %.0f, %.0f, %s",
		candle.Period.Start,
		dir,
		candle.Open,
		candle.Close,
		(candle.Close-candle.Open)/candle.Open*100,
		totalBV,
		totalSV,
		totalBV-totalSV,
		which,
	))
	return false, "", 0, false
}

func (b *Bot) isEntry3(trade Trade) (is_entry bool, entry_side string, trigger_volume float64, reverse bool) {
	b.firstOfCandle = false
	prev_candle := b.getCandle(1)
	var moving_side string
	if prev_candle.BodyLength() > 0 {
		moving_side = "buy"
	} else {
		moving_side = "sell"
	}

	prev_prev_candle := b.getCandle(2)
	if math.Abs(prev_candle.BodyLength()) < math.Abs(prev_prev_candle.BodyLength())*2 {
		return false, "", 0, false
	}

	var v, avgV float64
	if moving_side == "buy" {
		v = prev_candle.BuyVolume
		if v < b.config.minimumVolume {
			return false, "", 0, false
		}
		avgV = b.avgVolume(moving_side, b.config.avgVolumePeriod, 1)
		if v < avgV*b.config.againstAvgVolumeRate {
			return false, "", 0, false
		}
	} else {
		v = prev_candle.SellVolume
		if v < b.config.minimumVolume {
			return false, "", 0, false
		}
		avgV = b.avgVolume(moving_side, b.config.avgVolumePeriod, 1)
		if v < avgV*b.config.againstAvgVolumeRate {
			return false, "", 0, false
		}
	}

	b.logger.Log(fmt.Sprintf("%s %s出来高 %.2f が過去%d足の出来高平均%.2f x %.1fを超えました",
		trade.Time,
		moving_side,
		v,
		b.config.avgVolumePeriod,
		avgV,
		b.config.againstAvgVolumeRate,
	))

	return true, moving_side, v, false
}

func (b *Bot) isEntry2(trade Trade) (is_entry bool, entry_side string, trigger_volume float64, reverse bool) {

	candle := b.getCandle(0)
	var moving_side string
	if candle.BodyLength() > 0 {
		moving_side = "buy"
	} else {
		moving_side = "sell"
	}

	// 方向はBB3にタッチするかどうかで決まる
	// 条件x BB2にタッチしていたらにする
	var over_bb2 bool = false
	if trade.Price > b.upperPrice-(b.volatility*1.5) { //bb2の内側0.5volatility分はもう外とみなす
		entry_side = "sell"
		over_bb2 = true
	} else if trade.Price < b.lowerPrice+(b.volatility*1.5) {
		entry_side = "buy"
		over_bb2 = true
	} else {
		return false, "", 0, false
		// entry_side = moving_side
	}

	// 条件1 candle bodyが2Std以上
	if math.Abs(candle.BodyLength()) < b.config.minimumCandleLength*b.volatility {
		return false, "", 0, false
	}
	// b.logger.Log(fmt.Sprintf("%s ボラが規定量を超えました", trade.Time))
	// b.log = append(b.log, fmt.Sprintf("%s ロウソクの長さが %.3f x 2 を超えました", trade.Time, b.volatility))
	// 条件2 外向きの場合はBB3にタッチしていること
	r := b.config.guardOverBb3
	if over_bb2 {
		// b.logger.Log(fmt.Sprintf("%s %.3f, %.3f, %.3f, %s", trade.Time, trade.Price, b.upperPrice, b.lowerPrice, entry_side))
		if entry_side == "sell" && trade.Price < b.upperPrice*(1+r) {
			return false, "", 0, false
		}
		if entry_side == "buy" && trade.Price > b.lowerPrice*(1-r) {
			return false, "", 0, false
		}
	}
	// b.logger.Log(fmt.Sprintf("%s BB3を超えました", trade.Time))

	// 条件3 出来高が過去X足の平均よりY倍あること
	// 条件4 最低出来高を上回っていること
	var v, avgV float64
	if moving_side == "buy" {
		v = b.getCandle(0).BuyVolume
		if v < b.config.minimumVolume {
			return false, "", 0, false
		}
		avgV = b.avgVolume(moving_side, b.config.avgVolumePeriod, 0)
		if v < avgV*b.config.againstAvgVolumeRate {
			return false, "", 0, false
		}
	} else {
		v = b.getCandle(0).SellVolume
		if v < b.config.minimumVolume {
			return false, "", 0, false
		}
		avgV = b.avgVolume(moving_side, b.config.avgVolumePeriod, 0)
		if v < avgV*b.config.againstAvgVolumeRate {
			return false, "", 0, false
		}
	}
	b.logger.Log(fmt.Sprintf("%s %s出来高 %.2f が過去%d足の出来高平均%.2f x %.1fを超えました",
		trade.Time,
		moving_side,
		v,
		b.config.avgVolumePeriod,
		avgV,
		b.config.againstAvgVolumeRate,
	))

	return true, entry_side, v, entry_side != moving_side
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
		b.logger.Log(fmt.Sprintf("出来高はあるが変動幅がVolatility以下なのでスルー %+v", trade))
		return false, "", 0
	}

	if buyV > sellV {
		return is_entry, "buy", buyV
	} else {
		return is_entry, "sell", sellV
	}
}

func (b *Bot) entry(side string, lot float64, v float64, trade Trade, reverse bool) {
	if b.state != 0 {
		return
	}
	if side == "buy" {
		trade.Price *= (1 + slippage)
		b.logger.Log(fmt.Sprintf("%s, volume: %.4f ロングエントリー Size: %.4f, Price: %.5f",
			trade.Time,
			v,
			trade.Size,
			trade.Price,
		))
		b.openPosition(side, lot, trade, reverse)
	} else {
		trade.Price *= (1 - slippage)
		b.logger.Log(fmt.Sprintf("%s, volume: %.4f ショートエントリー Size: %.4f, Price: %.5f",
			trade.Time,
			v,
			trade.Size,
			trade.Price,
		))
		b.openPosition(side, lot, trade, reverse)
	}
}

func (b *Bot) settle(trade Trade, isTaker bool) {
	if b.state != 1 {
		return
	}

	settle_price := b.client.Close(b.market, b.position, trade.Price)

	// Fe
	if isTaker {
		b.paidFee += trade.Price * b.position.Size * taker_fee
	} else {
		b.paidFee += trade.Price * b.position.Size * maker_fee
	}
	fee := b.paidFee

	// market close order
	var pnl float64
	if b.position.Side == "buy" {
		pnl = (settle_price - b.position.Price) * b.position.Size
		b.result.longProfit = append(b.result.longProfit, pnl-fee)
		b.result.longReturn = append(b.result.longReturn, 100*(pnl-fee)/b.position.Price)
		if pnl-fee > 0 {
			b.result.longWinning++
		}
		b.result.longCount++
	} else {
		pnl = (b.position.Price - settle_price) * b.position.Size
		b.result.shortProfit = append(b.result.shortProfit, pnl-fee)
		b.result.shortReturn = append(b.result.shortReturn, 100*(pnl-fee)/b.position.Price)
		if pnl-fee > 0 {
			b.result.shortWinning++
		}
		b.result.shortCount++
	}
	b.logger.Log(fmt.Sprintf("%s, 決済しました  Size: %.3f, Price: %.5f, 建値: %.5f, OpenTime: %s, Pnl: %.4f\n",
		trade.Time,
		b.position.Size,
		settle_price,
		b.position.Price,
		trade.Time.Sub(b.position.Time),
		pnl-fee,
	))
	b.result.totalPnl += pnl
	b.result.totalFee += fee

	// 最後に決済した時刻を保存
	b.lastCloseTime = trade.Time

	b.state = 2
	b.nunpin = 0
	b.position = Position{}
	b.paidFee = 0
}

func (b *Bot) openPosition(side string, lot float64, trade Trade, reverse bool) {
	new_position := b.client.MarketOrder(b.market, side, lot, trade.Time, trade.Price)
	new_position.Reverse = reverse
	b.position = b.position.add(new_position)
	b.logger.Log(fmt.Sprintf("%s, Position  %v",
		trade.Time,
		b.position))
	b.state = 1
	b.paidFee += trade.Price * lot * taker_fee
}

func (p *Position) add(position Position) Position {
	return Position{
		Time:    position.Time,
		Side:    position.Side,
		Size:    p.Size + position.Size,
		Price:   (p.Price*p.Size + position.Price*position.Size) / (p.Size + position.Size),
		Reverse: position.Reverse,
	}
}
