package internal

import (
	"time"
)

// TimePeriod is a simple struct that describes a period of time with a Start and End time
type TimePeriod struct {
	Start time.Time
	End   time.Time
}

func NewMinuteFromTime(t time.Time) TimePeriod {
	jst, _ := time.LoadLocation("Asia/Tokyo")
	start := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, jst)
	return TimePeriod{
		Start: start,
		End:   start.Add(time.Second * time.Duration(59)),
	}
}

func (tp *TimePeriod) Contain(t time.Time) bool {
	return tp.Start.Unix() <= t.Unix() && t.Unix() <= tp.End.Unix()
}
