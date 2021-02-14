package domain

import "time"

type Ohlc struct {
	ID         uint      `json:"id"`
	Market     string    `json:"market"`
	Close      float64   `json:"close"`
	High       float64   `json:"high"`
	Low        float64   `json:"low"`
	Open       float64   `json:"open"`
	Volume     float64   `json:"volume"`
	StartTime  time.Time `json:"startTime"`
	Resolution int       `json:"resolution"`
}
