package domain

import "github.com/go-numb/go-ftx/rest/public/markets"

type Ohlc struct {
    markets.Candle
    Resolution int
}
