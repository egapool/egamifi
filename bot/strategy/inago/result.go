package inago

import (
	"fmt"
	"time"
)

// TODO ryota-trade見る
type Result struct {
	totalPnl     float64
	totalFee     float64
	longCount    int
	longWinning  int
	longReturn   []float64
	longProfit   []float64
	shortCount   int
	shortWinning int
	shortReturn  []float64
	shortProfit  []float64
	startTime    time.Time
	endTime      time.Time
}

func (r *Result) winRate() float64 {
	return float64(r.winCount()) / float64(r.tradeCount())
}

func (r *Result) tradeCount() int {
	return r.longCount + r.shortCount
}

func (r *Result) winCount() int {
	return r.longWinning + r.shortWinning
}

func (r *Result) pf() float64 {
	var profit, loss float64
	for _, p := range r.longProfit {
		if p > 0 {
			profit += p
		} else {
			loss += p
		}
	}
	for _, p := range r.shortProfit {
		if p > 0 {
			profit += p
		} else {
			loss += p
		}
	}
	return -1 * profit / loss

}

func (r *Result) render() {
	fmt.Println("期間", r.startTime, "〜", r.endTime)
	fmt.Println("トータルPnl:", r.totalPnl)
	fmt.Println("トータルFee:", r.totalFee)
	fmt.Println("トータルProfit:", r.totalPnl-r.totalFee)
	fmt.Println("ロング回数:", r.longCount)
	fmt.Println("ショート回数:", r.shortCount)
	fmt.Println("Win:", r.winCount)
	fmt.Println("Lose:", r.tradeCount()-r.winCount())
	fmt.Println("勝率:", r.winRate())
}
