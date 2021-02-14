package domain

import "time"

type Ohlc struct {
	ID         uint      `json:"id"`
	Market     string    `json:"market"`
	Open       float64   `json:"open"`
	High       float64   `json:"high"`
	Low        float64   `json:"low"`
	Close      float64   `json:"close"`
	Volume     float64   `json:"volume"`
	StartTime  time.Time `json:"startTime"`
	Resolution int       `json:"resolution"`
	Exchanger  string    `json:"exchanger"`
}
