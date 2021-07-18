package bot

import (
	"testing"

	"github.com/egapool/egamifi/bot/strategy/inago"
)

func TestBacktest(t *testing.T) {
	var scope int64 = 100
	var volumeTriger float64 = 300000
	var settleTerm int64 = 60
	var reverse bool = true
	inago_config := inago.NewConfig(scope, volumeTriger, settleTerm, reverse)
	backtest := NewBacktest(inago.NewBot("AXS-PERP", inago_config))
	trade_file := "../data/ftx-trades-AXS-PERP-20210718220000-20210717220000.csv"
	backtest.Run(trade_file, false)
}