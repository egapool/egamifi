package internal

type Candle struct {
	Period     TimePeriod
	Open       float64
	High       float64
	Low        float64
	Close      float64
	Volume     float64
	TradeCount uint
}

// NewCandle returns a new *Candle for a given time period
func NewCandle(period TimePeriod) (c *Candle) {
	return &Candle{
		Period: period,
		Open:   0.0,
		High:   0.0,
		Low:    0.0,
		Close:  0.0,
		Volume: 0.0,
	}
}

// AddTrade adds a trade to this candle. It will determine if the current price is higher or lower than the min or max
// price and increment the tradecount.
func (c *Candle) AddTrade(tradeAmount, tradePrice float64) {
	if c.Open == 0.0 {
		c.Open = tradePrice
	}
	c.Close = tradePrice

	if c.High == 0.0 {
		c.High = tradePrice
	} else if tradePrice > c.High {
		c.High = tradePrice
	}

	if c.Low == 0.0 {
		c.Low = tradePrice
	} else if tradePrice < c.Low {
		c.Low = tradePrice
	}

	if c.Volume == 0.0 {
		c.Volume = tradeAmount
	} else {
		c.Volume += tradeAmount
	}

	c.TradeCount++
}
