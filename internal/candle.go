package internal

type Candle struct {
	Period     TimePeriod
	Open       float64
	High       float64
	Low        float64
	Close      float64
	BuyVolume  float64
	SellVolume float64
	TradeCount uint
}

// NewCandle returns a new *Candle for a given time period
func NewCandle(period TimePeriod) (c *Candle) {
	return &Candle{
		Period:     period,
		Open:       0.0,
		High:       0.0,
		Low:        0.0,
		Close:      0.0,
		BuyVolume:  0.0,
		SellVolume: 0.0,
	}
}

// AddTrade adds a trade to this candle. It will determine if the current price is higher or lower than the min or max
// price and increment the tradecount.
func (c *Candle) AddTrade(tradeAmount, tradePrice float64, side string) {
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

	if side == "buy" {
		if c.BuyVolume == 0.0 {
			c.BuyVolume = tradeAmount
		} else {
			c.BuyVolume += tradeAmount
		}
	} else if side == "sell" {
		if c.SellVolume == 0.0 {
			c.SellVolume = tradeAmount
		} else {
			c.SellVolume += tradeAmount
		}
	}

	c.TradeCount++
}

func (c *Candle) BodyLength() float64 {
	return c.Close - c.Open
}
