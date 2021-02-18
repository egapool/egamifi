package domain

import (
	"log"
	"time"
)

type DeviationRate struct {
	QuarterMarket string
	QuarterPrice  float64
	PerpMarket    string
	PerpPrice     float64
	Time          time.Time
	DeviationRate float64
}

type DeviationRates []DeviationRate

func NewDeviationRate(quarter, perp Ohlc) DeviationRate {
	if quarter.StartTime != perp.StartTime {
		log.Fatal("Both time is not same!")
	}
	// ohlcのtimeがStartTimeなので価格もOpenを利用する
	dr := DeviationRate{
		QuarterMarket: quarter.Market,
		QuarterPrice:  quarter.Open,
		PerpMarket:    perp.Market,
		PerpPrice:     perp.Open,
		Time:          quarter.StartTime,
	}
	dr.DeviationRate = dr.calcDeviationRate()
	return dr
}

func (dr *DeviationRate) calcDeviationRate() float64 {
	return (dr.QuarterPrice - dr.PerpPrice) * 2 / (dr.QuarterPrice + dr.PerpPrice)
}
