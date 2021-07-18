package bot

import (
	"testing"

	"github.com/egapool/egamifi/bot/strategy/inago"
)

func TestBacktest(t *testing.T) {
	backtest := NewBacktest(inago.NewBot("AXS-PERP"))
	backtest.Run()
}
